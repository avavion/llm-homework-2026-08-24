package regulation

import (
	"errors"
	"testing"
	"time"

	"llm-homework/backend/internal/product"
)

var expiry = time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

func confirmedRule(dateType product.DateType, postExpiry PostExpiryStatus) Rule {
	return Rule{
		RegulatorGroup:    "test_group",
		CountryCode:       "ZZ",
		DateType:          dateType,
		Status:            StatusEnabled,
		ExpiryTimezone:    "UTC",
		ExpiryInstantRule: ExpiryAtEndOfDay,
		PostExpiryStatus:  postExpiry,
	}
}

func TestUseByAfterExpiryIsExcluded(t *testing.T) {
	useByProduct := product.Product{DateType: product.DateTypeUseBy, ExpiryDate: expiry}
	rule := confirmedRule(product.DateTypeUseBy, PostExpiryExpired)

	status, err := Evaluate(useByProduct, rule, ExpiryInstantOf(t, useByProduct, rule).Add(time.Second))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if status != StatusExpired || EligibleForRecipes(status) {
		t.Fatalf("status = %v, eligible = %v", status, EligibleForRecipes(status))
	}
}

func TestBestBeforeAfterExpiryNeedsAttention(t *testing.T) {
	bestBeforeProduct := product.Product{DateType: product.DateTypeBestBefore, ExpiryDate: expiry}
	rule := confirmedRule(product.DateTypeBestBefore, PostExpiryAttention)

	status, err := Evaluate(bestBeforeProduct, rule, ExpiryInstantOf(t, bestBeforeProduct, rule).Add(time.Second))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if status != StatusAttention || !EligibleForRecipes(status) {
		t.Fatalf("status = %v, eligible = %v", status, EligibleForRecipes(status))
	}
}

func TestBeforeExpiryIsActive(t *testing.T) {
	item := product.Product{DateType: product.DateTypeUseBy, ExpiryDate: expiry}
	rule := confirmedRule(product.DateTypeUseBy, PostExpiryExpired)

	status, err := Evaluate(item, rule, ExpiryInstantOf(t, item, rule).Add(-time.Second))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if status != StatusActive || !EligibleForRecipes(status) {
		t.Fatalf("status = %v, eligible = %v", status, EligibleForRecipes(status))
	}
}

func TestResearchRequiredRuleIsNotAutomated(t *testing.T) {
	item := product.Product{DateType: product.DateTypeUseBy, ExpiryDate: expiry}
	rule := Rule{Status: StatusResearchRequired}

	_, err := Evaluate(item, rule, expiry.Add(24*time.Hour))
	if !errors.Is(err, ErrAutomationUnavailable) {
		t.Fatalf("err = %v, want ErrAutomationUnavailable", err)
	}
}

func TestEURegistryHasNoEnabledRowsYet(t *testing.T) {
	repository := NewRepository()
	rule, ok := repository.RuleFor("DE", product.DateTypeUseBy)
	if !ok {
		t.Fatal("expected a research_required row for DE/use_by")
	}
	if rule.Status != StatusResearchRequired {
		t.Fatalf("status = %v, want research_required until a source confirms the row", rule.Status)
	}
}

func ExpiryInstantOf(t *testing.T, item product.Product, rule Rule) time.Time {
	t.Helper()
	instant, err := ExpiryInstant(item, rule)
	if err != nil {
		t.Fatalf("ExpiryInstant err = %v", err)
	}
	return instant
}
