package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type memoryStore struct {
	profiles map[uuid.UUID]Profile
	settings map[uuid.UUID]map[string]int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{profiles: map[uuid.UUID]Profile{}, settings: map[uuid.UUID]map[string]int{}}
}

func (store *memoryStore) Profile(_ context.Context, accountID uuid.UUID) (Profile, error) {
	return store.profiles[accountID], nil
}

func (store *memoryStore) SaveProfile(_ context.Context, accountID uuid.UUID, input ProfileInput) (Profile, error) {
	result := Profile{CountryCode: input.CountryCode, Language: input.Language}
	store.profiles[accountID] = result
	return result, nil
}

func (store *memoryStore) Settings(_ context.Context, accountID uuid.UUID) (map[string]int, error) {
	return store.settings[accountID], nil
}

func (store *memoryStore) SaveSetting(_ context.Context, accountID uuid.UUID, group string, threshold int) error {
	if store.settings[accountID] == nil {
		store.settings[accountID] = map[string]int{}
	}
	store.settings[accountID][group] = threshold
	return nil
}

func testRouter(store Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	resolver := func(c *gin.Context) (uuid.UUID, bool) {
		accountID, err := uuid.Parse(c.GetHeader("X-Test-Account"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return uuid.UUID{}, false
		}
		return accountID, true
	}
	RegisterRoutes(router, NewService(store, RegulatorGroupFunc(func(country string) string {
		if country == "DE" { return "eu_1169_2011" }
		return ""
	})), resolver)
	return router
}

func requestAs(router *gin.Engine, accountID uuid.UUID, method, path string, body any) *httptest.ResponseRecorder {
	var data []byte
	if body != nil { data, _ = json.Marshal(body) }
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Account", accountID.String())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func TestProfileIsAccountScopedAndReturnsRegistryGroup(t *testing.T) {
	store := newMemoryStore()
	router := testRouter(store)
	owner, stranger := uuid.New(), uuid.New()

	saved := requestAs(router, owner, http.MethodPut, "/v1/profile", map[string]string{"country_code": "DE", "language": "en"})
	if saved.Code != http.StatusOK { t.Fatalf("save status=%d body=%s", saved.Code, saved.Body.String()) }

	loaded := requestAs(router, owner, http.MethodGet, "/v1/profile", nil)
	var profile PublicProfile
	if err := json.Unmarshal(loaded.Body.Bytes(), &profile); err != nil { t.Fatal(err) }
	if profile.CountryCode != "DE" || profile.Language != "en" || profile.RegulatorGroup != "eu_1169_2011" { t.Fatalf("profile=%+v", profile) }

	other := requestAs(router, stranger, http.MethodGet, "/v1/profile", nil)
	var empty PublicProfile
	_ = json.Unmarshal(other.Body.Bytes(), &empty)
	if empty.CountryCode != "" || empty.Language != "" { t.Fatalf("foreign profile leaked: %+v", empty) }
}

func TestProfileRejectsInvalidCountryOrLanguage(t *testing.T) {
	router := testRouter(newMemoryStore())
	accountID := uuid.New()
	for _, input := range []map[string]string{
		{"country_code": "D", "language": "en"},
		{"country_code": "DE", "language": "fr"},
	} {
		response := requestAs(router, accountID, http.MethodPut, "/v1/profile", input)
		if response.Code != http.StatusBadRequest { t.Fatalf("input=%v status=%d", input, response.Code) }
	}
}

func TestNotificationSettingsUseDefaultsAndAreAccountScoped(t *testing.T) {
	store := newMemoryStore()
	router := testRouter(store)
	owner, stranger := uuid.New(), uuid.New()

	defaultResponse := requestAs(router, owner, http.MethodGet, "/v1/notification-settings", nil)
	var defaults NotificationSettingsResponse
	_ = json.Unmarshal(defaultResponse.Body.Bytes(), &defaults)
	if len(defaults.Settings) != len(DefaultThresholdMinutes) { t.Fatalf("default count=%d", len(defaults.Settings)) }

	bad := requestAs(router, owner, http.MethodPut, "/v1/notification-settings", map[string]any{"product_group": "other", "alert_threshold_minutes": 59})
	if bad.Code != http.StatusBadRequest { t.Fatalf("low threshold status=%d", bad.Code) }

	saved := requestAs(router, owner, http.MethodPut, "/v1/notification-settings", map[string]any{"product_group": "other", "alert_threshold_minutes": 60})
	if saved.Code != http.StatusOK { t.Fatalf("save status=%d body=%s", saved.Code, saved.Body.String()) }

	other := requestAs(router, stranger, http.MethodGet, "/v1/notification-settings", nil)
	var strangerSettings NotificationSettingsResponse
	_ = json.Unmarshal(other.Body.Bytes(), &strangerSettings)
	for _, setting := range strangerSettings.Settings {
		if setting.ProductGroup == "other" && setting.AlertThresholdMinutes != DefaultThresholdMinutes["other"] { t.Fatalf("foreign setting leaked: %+v", setting) }
	}
}
