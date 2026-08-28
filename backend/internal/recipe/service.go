// Package recipe suggests simple, deterministic recipes from a user's own
// active products, honoring the regulation package's recipe-eligibility gate.
package recipe

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
	"llm-homework/backend/internal/regulation"
)

// RuleLookup is the subset of regulation.Repository the service depends on.
type RuleLookup interface {
	RuleFor(countryCode string, dateType product.DateType) (regulation.Rule, bool)
}

// Kind identifies which client-side copy template a Recipe pairs with, so the
// frontend can render a localized title instead of a hardcoded English one.
type Kind string

const (
	KindUseUp        Kind = "use_up"
	KindCombineGroup Kind = "combine_group"
)

type Recipe struct {
	Kind        Kind
	ProductName string
	GroupName   string
	ProductIDs  []uuid.UUID
}

type Service struct {
	rules RuleLookup
	now   func() time.Time
}

func NewService(rules RuleLookup) *Service {
	return &Service{rules: rules, now: time.Now}
}

// EligibleProducts returns the caller's active products that the regulation
// registry does not exclude. A product whose country/date-type combination
// has no confirmed rule stays eligible: absence of evidence is never treated
// as evidence of expiry.
func (service *Service) EligibleProducts(items []product.Product) []product.Product {
	now := service.now()
	eligible := make([]product.Product, 0, len(items))
	for _, item := range items {
		if item.LifecycleStatus != product.LifecycleActive {
			continue
		}
		if service.isRecipeEligible(item, now) {
			eligible = append(eligible, item)
		}
	}
	return eligible
}

func (service *Service) isRecipeEligible(item product.Product, now time.Time) bool {
	if item.CountryCode == nil {
		return true
	}
	rule, ok := service.rules.RuleFor(*item.CountryCode, item.DateType)
	if !ok {
		return true
	}

	status, err := regulation.Evaluate(item, rule, now)
	if errors.Is(err, regulation.ErrAutomationUnavailable) {
		return true
	}
	if err != nil {
		return true
	}
	return regulation.EligibleForRecipes(status)
}

// Suggest builds one recipe per eligible product plus, where a product group
// has more than one eligible member, a group recipe. Ordering follows the
// input slice so results stay deterministic for a given product list.
func (service *Service) Suggest(items []product.Product) []Recipe {
	eligible := service.EligibleProducts(items)

	recipes := make([]Recipe, 0, len(eligible))
	groupIDs := map[string][]uuid.UUID{}
	var groupOrder []string

	for _, item := range eligible {
		recipes = append(recipes, Recipe{Kind: KindUseUp, ProductName: item.Name, ProductIDs: []uuid.UUID{item.ID}})

		groupName := "other"
		if item.ProductGroup != nil && *item.ProductGroup != "" {
			groupName = *item.ProductGroup
		}
		if _, seen := groupIDs[groupName]; !seen {
			groupOrder = append(groupOrder, groupName)
		}
		groupIDs[groupName] = append(groupIDs[groupName], item.ID)
	}

	for _, groupName := range groupOrder {
		ids := groupIDs[groupName]
		if len(ids) < 2 {
			continue
		}
		recipes = append(recipes, Recipe{Kind: KindCombineGroup, GroupName: groupName, ProductIDs: ids})
	}

	return recipes
}
