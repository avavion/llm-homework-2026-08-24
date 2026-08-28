// Package profile owns account-scoped country/language and e-mail reminder
// preferences. Regulatory classification remains read-only: it is derived
// from the reviewed registry and is never supplied by a client.
package profile

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

const MinAlertThresholdMinutes = 60

// DefaultCountryCode and DefaultLanguage seed every new account's profile.
// The country selector isn't built yet, so accounts default to Russia (an
// EAEU member) rather than being blocked until they visit Settings; the
// EAEU regulatory rules themselves stay research_required, so this default
// unblocks the UI without inventing an expiry rule.
const (
	DefaultCountryCode = "RU"
	DefaultLanguage    = "ru"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
	ErrInvalidSetting = errors.New("invalid notification setting")
)

// DefaultThresholdMinutes mirrors shared/docs/product-group-alert-policy.md.
// It is deliberately local and explicit so a database migration is not needed
// merely to update a documented default for accounts without an override.
var DefaultThresholdMinutes = map[string]int{
	"refrigerated_perishable": 4320,
	"fresh_produce":           2880,
	"frozen":                  10080,
	"shelf_stable":            20160,
	"other":                   4320,
}

type Profile struct {
	CountryCode string
	Language    string
}

type ProfileInput struct {
	CountryCode string
	Language    string
}

type Store interface {
	Profile(ctx context.Context, accountID uuid.UUID) (Profile, error)
	SaveProfile(ctx context.Context, accountID uuid.UUID, input ProfileInput) (Profile, error)
	Settings(ctx context.Context, accountID uuid.UUID) (map[string]int, error)
	SaveSetting(ctx context.Context, accountID uuid.UUID, productGroup string, thresholdMinutes int) error
}

type RegulatorGroupLookup interface {
	RegulatorGroupForCountry(countryCode string) string
}

type RegulatorGroupFunc func(countryCode string) string

func (fn RegulatorGroupFunc) RegulatorGroupForCountry(countryCode string) string {
	return fn(countryCode)
}

type Service struct {
	store Store
	rules RegulatorGroupLookup
}

func NewService(store Store, rules RegulatorGroupLookup) *Service {
	return &Service{store: store, rules: rules}
}

func (service *Service) Profile(ctx context.Context, accountID uuid.UUID) (Profile, error) {
	return service.store.Profile(ctx, accountID)
}

func (service *Service) SaveProfile(ctx context.Context, accountID uuid.UUID, input ProfileInput) (Profile, error) {
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	input.Language = strings.ToLower(strings.TrimSpace(input.Language))
	if !validCountryCode(input.CountryCode) || (input.Language != "ru" && input.Language != "en") {
		return Profile{}, ErrInvalidProfile
	}
	return service.store.SaveProfile(ctx, accountID, input)
}

// InitializeProfile seeds a newly registered account with the default
// country/language so it is never blocked on the missing country selector.
func (service *Service) InitializeProfile(ctx context.Context, accountID uuid.UUID) error {
	_, err := service.SaveProfile(ctx, accountID, ProfileInput{CountryCode: DefaultCountryCode, Language: DefaultLanguage})
	return err
}

func (service *Service) RegulatorGroup(countryCode string) string {
	if countryCode == "" || service.rules == nil {
		return ""
	}
	return service.rules.RegulatorGroupForCountry(countryCode)
}

func (service *Service) Settings(ctx context.Context, accountID uuid.UUID) ([]Setting, error) {
	overrides, err := service.store.Settings(ctx, accountID)
	if err != nil {
		return nil, err
	}
	settings := make([]Setting, 0, len(DefaultThresholdMinutes))
	for _, group := range supportedGroups {
		threshold := DefaultThresholdMinutes[group]
		if value, ok := overrides[group]; ok {
			threshold = value
		}
		settings = append(settings, Setting{ProductGroup: group, AlertThresholdMinutes: threshold})
	}
	return settings, nil
}

func (service *Service) SaveSetting(ctx context.Context, accountID uuid.UUID, productGroup string, thresholdMinutes int) error {
	if _, ok := DefaultThresholdMinutes[productGroup]; !ok || thresholdMinutes < MinAlertThresholdMinutes {
		return ErrInvalidSetting
	}
	return service.store.SaveSetting(ctx, accountID, productGroup, thresholdMinutes)
}

var supportedGroups = []string{
	"refrigerated_perishable", "fresh_produce", "frozen", "shelf_stable", "other",
}

func validCountryCode(value string) bool {
	return len(value) == 2 && value[0] >= 'A' && value[0] <= 'Z' &&
		value[1] >= 'A' && value[1] <= 'Z'
}

type Setting struct {
	ProductGroup          string `json:"product_group"`
	AlertThresholdMinutes int    `json:"alert_threshold_minutes"`
}
