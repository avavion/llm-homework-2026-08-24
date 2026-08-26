// Package notification schedules idempotent expiry-reminder e-mails from a
// confirmed regulation rule and the product-group alert policy defaults.
package notification

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/regulation"
)

const ChannelEmail = "email"

// MinAlertThreshold matches the product-group alert policy floor.
const MinAlertThreshold = 60 * time.Minute

var ErrThresholdTooSmall = errors.New("alert threshold must be at least 60 minutes")

// defaultAlertWindowMinutesByGroup mirrors shared/docs/product-group-alert-policy.md.
var defaultAlertWindowMinutesByGroup = map[string]int{
	"refrigerated_perishable": 4320,
	"fresh_produce":           2880,
	"frozen":                  10080,
	"shelf_stable":            20160,
	"other":                   4320,
}

// ResolveThresholdMinutes returns the user's own alert threshold, falling
// back to the product's group default and finally to "other" for an unknown
// or empty group.
func ResolveThresholdMinutes(item product.Product) int {
	if item.AlertThresholdMinutes != nil {
		return *item.AlertThresholdMinutes
	}
	group := "other"
	if item.ProductGroup != nil && *item.ProductGroup != "" {
		group = *item.ProductGroup
	}
	if minutes, ok := defaultAlertWindowMinutesByGroup[group]; ok {
		return minutes
	}
	return defaultAlertWindowMinutesByGroup["other"]
}

// NextNotificationAt schedules a reminder a threshold before the confirmed
// expiry instant: alert_at = verified_expiry_instant - alert_window.
func NextNotificationAt(expiryInstant time.Time, threshold time.Duration) (time.Time, error) {
	if threshold < MinAlertThreshold {
		return time.Time{}, ErrThresholdTooSmall
	}
	return expiryInstant.Add(-threshold), nil
}

// Candidate pairs a product with the e-mail address its reminder goes to.
type Candidate struct {
	Product        product.Product
	RecipientEmail string
}

// ProductSource lists active products that may need a reminder scheduled.
type ProductSource interface {
	ActiveWithCountry(ctx context.Context) ([]Candidate, error)
}

// RuleLookup is the subset of regulation.Repository the scheduler depends on.
type RuleLookup interface {
	RuleFor(countryCode string, dateType product.DateType) (regulation.Rule, bool)
}

// DeliveryLog guarantees at-most-once delivery per (product, scheduled
// instant, channel). Claim must be atomic: only the caller that successfully
// claims a slot may call the sender for it.
type DeliveryLog interface {
	Claim(ctx context.Context, productID uuid.UUID, scheduledFor time.Time, channel string) (claimed bool, err error)
}

type Service struct {
	products   ProductSource
	rules      RuleLookup
	deliveries DeliveryLog
	sender     EmailSender
}

func NewService(products ProductSource, rules RuleLookup, deliveries DeliveryLog, sender EmailSender) *Service {
	return &Service{products: products, rules: rules, deliveries: deliveries, sender: sender}
}

// DeliverDue evaluates every candidate against its confirmed rule (if any)
// and sends a reminder for each one whose scheduled instant has arrived and
// has not already been claimed. A candidate with no confirmed rule, or whose
// threshold has not yet elapsed, is silently skipped: that is the intended
// behavior for `research_required` rows and future reminders alike.
func (service *Service) DeliverDue(ctx context.Context, now time.Time) error {
	candidates, err := service.products.ActiveWithCountry(ctx)
	if err != nil {
		return err
	}

	var deliveryErrors []error
	for _, candidate := range candidates {
		if err := service.deliverIfDue(ctx, candidate, now); err != nil {
			deliveryErrors = append(deliveryErrors, err)
		}
	}
	return errors.Join(deliveryErrors...)
}

func (service *Service) deliverIfDue(ctx context.Context, candidate Candidate, now time.Time) error {
	item := candidate.Product
	if item.CountryCode == nil {
		return nil
	}

	rule, ok := service.rules.RuleFor(*item.CountryCode, item.DateType)
	if !ok || rule.Status != regulation.StatusEnabled {
		return nil
	}

	expiryInstant, err := regulation.ExpiryInstant(item, rule)
	if err != nil {
		return nil
	}

	scheduledFor, err := NextNotificationAt(expiryInstant, time.Duration(ResolveThresholdMinutes(item))*time.Minute)
	if err != nil || now.Before(scheduledFor) {
		return nil
	}

	claimed, err := service.deliveries.Claim(ctx, item.ID, scheduledFor, ChannelEmail)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	return service.sender.SendExpiryReminder(ctx, candidate.RecipientEmail, item)
}
