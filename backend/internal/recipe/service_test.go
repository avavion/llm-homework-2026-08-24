package recipe

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/regulation"
)

type fakeRules struct {
	rule regulation.Rule
	has  bool
}

func (rules fakeRules) RuleFor(_ string, _ product.DateType) (regulation.Rule, bool) {
	return rules.rule, rules.has
}

func confirmedUseByRule() regulation.Rule {
	return regulation.Rule{
		DateType:          product.DateTypeUseBy,
		Status:            regulation.StatusEnabled,
		ExpiryTimezone:    "UTC",
		ExpiryInstantRule: regulation.ExpiryAtEndOfDay,
		PostExpiryStatus:  regulation.PostExpiryExpired,
	}
}

func TestExpiredUseByProductIsExcludedFromRecipes(t *testing.T) {
	country := "ZZ"
	expiry := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	item := product.Product{
		ID: uuid.New(), Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: expiry,
		CountryCode: &country, LifecycleStatus: product.LifecycleActive,
	}
	service := NewService(fakeRules{rule: confirmedUseByRule(), has: true})
	service.now = func() time.Time { return expiry.Add(48 * time.Hour) }

	eligible := service.EligibleProducts([]product.Product{item})
	if len(eligible) != 0 {
		t.Fatalf("eligible = %+v, want none", eligible)
	}
}

func TestUnconfirmedRuleKeepsProductEligible(t *testing.T) {
	country := "ZZ"
	item := product.Product{
		ID: uuid.New(), Name: "Milk", DateType: product.DateTypeUseBy,
		ExpiryDate: time.Now().Add(-48 * time.Hour), CountryCode: &country,
		LifecycleStatus: product.LifecycleActive,
	}
	service := NewService(fakeRules{rule: regulation.Rule{Status: regulation.StatusResearchRequired}, has: true})

	eligible := service.EligibleProducts([]product.Product{item})
	if len(eligible) != 1 {
		t.Fatalf("eligible = %+v, want the product to remain visible", eligible)
	}
}

func TestCompletedProductsAreNeverSuggested(t *testing.T) {
	item := product.Product{ID: uuid.New(), Name: "Milk", LifecycleStatus: product.LifecycleUsed}
	service := NewService(fakeRules{})

	if recipes := service.Suggest([]product.Product{item}); len(recipes) != 0 {
		t.Fatalf("recipes = %+v, want none for a completed product", recipes)
	}
}

func TestSuggestGroupsEligibleProductsSharingAGroup(t *testing.T) {
	group := "fresh_produce"
	first := product.Product{ID: uuid.New(), Name: "Apple", ProductGroup: &group, LifecycleStatus: product.LifecycleActive}
	second := product.Product{ID: uuid.New(), Name: "Pear", ProductGroup: &group, LifecycleStatus: product.LifecycleActive}
	service := NewService(fakeRules{})

	recipes := service.Suggest([]product.Product{first, second})
	if len(recipes) != 3 {
		t.Fatalf("recipes = %+v, want 2 single-product recipes plus 1 group recipe", recipes)
	}
	last := recipes[len(recipes)-1]
	if len(last.ProductIDs) != 2 {
		t.Fatalf("group recipe = %+v, want both products", last)
	}
}
