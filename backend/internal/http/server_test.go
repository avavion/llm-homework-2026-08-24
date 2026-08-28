package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/regulation"
)

type fakeRegulationRules struct {
	rule regulation.Rule
	ok   bool
}

func (rules fakeRegulationRules) RuleFor(string, product.DateType) (regulation.Rule, bool) {
	return rules.rule, rules.ok
}

func TestHealthzReturnsOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewServer(nil, nil, "").ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("response body = %q, want %q", body, "{\"status\":\"ok\"}\n")
	}
}

func TestAllowedFrontendOriginGetsCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/v1/products", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", "POST")
	response := httptest.NewRecorder()

	NewServer(nil, []string{"http://localhost:5173"}, "").ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q", got)
	}
}

func TestUnlistedOriginGetsNoCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/v1/products", nil)
	request.Header.Set("Origin", "https://not-allowed.example.com")
	response := httptest.NewRecorder()

	NewServer(nil, []string{"http://localhost:5173"}, "").ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty for an unlisted origin", got)
	}
}

func TestDisplayStatusUsesConfirmedRulesAndLifecyclePriority(t *testing.T) {
	now := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)
	country := "ZZ"
	base := product.Product{
		CountryCode: &country, DateType: product.DateTypeUseBy,
		ExpiryDate: time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		LifecycleStatus: product.LifecycleActive,
	}
	rule := regulation.Rule{
		Status: regulation.StatusEnabled, ExpiryTimezone: "UTC",
		ExpiryInstantRule: regulation.ExpiryAtEndOfDay,
		PostExpiryStatus: regulation.PostExpiryExpired,
	}
	if got := displayStatusFor(base, fakeRegulationRules{rule: rule, ok: true}, now); got != product.DisplayStatusExpired {
		t.Fatalf("display status = %q, want %q", got, product.DisplayStatusExpired)
	}

	rule.PostExpiryStatus = regulation.PostExpiryAttention
	if got := displayStatusFor(base, fakeRegulationRules{rule: rule, ok: true}, now); got != product.DisplayStatusAttention {
		t.Fatalf("display status = %q, want %q", got, product.DisplayStatusAttention)
	}

	base.LifecycleStatus = product.LifecycleUsed
	if got := displayStatusFor(base, fakeRegulationRules{rule: rule, ok: true}, now); got != product.DisplayStatusUsed {
		t.Fatalf("display status = %q, want %q", got, product.DisplayStatusUsed)
	}
}

func TestDisplayStatusIsResearchRequiredWithoutUsableRule(t *testing.T) {
	item := product.Product{DateType: product.DateTypeUseBy, LifecycleStatus: product.LifecycleActive}
	if got := displayStatusFor(item, fakeRegulationRules{}, time.Now()); got != product.DisplayStatusResearchRequired {
		t.Fatalf("display status = %q, want %q", got, product.DisplayStatusResearchRequired)
	}
}
