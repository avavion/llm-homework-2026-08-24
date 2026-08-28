package regulation

import "llm-homework/backend/internal/product"

// euCountryCodes lists the EU member states covered by the eu_1169_2011 rows
// of shared/docs/regulatory-date-rules.md.
var euCountryCodes = []string{
	"AT", "BE", "BG", "HR", "CY", "CZ", "DK", "EE", "FI", "FR", "DE", "GR",
	"HU", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PL", "PT", "RO", "SK",
	"SI", "ES", "SE",
}

const (
	sourceEU1169Article24 = "https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A32011R1169"
	accessedOnRegistry    = "2026-08-26"
)

// Repository is a read-only, in-memory copy of the reviewed registry: rows
// change only through the documented legal/QA review process, never at
// runtime, so there is no database table behind it.
//
// Only use_by/best_before rows are indexed here because those are the only
// product.DateType values the current product schema models. The EAEU
// shelf-life vocabulary and the unverified CIS group from the shared
// registry apply to date types this MVP does not yet capture, so they are
// intentionally left out of the queryable index.
type Repository struct {
	rules map[string]map[product.DateType]Rule
}

func NewRepository() *Repository {
	rules := make(map[string]map[product.DateType]Rule, len(euCountryCodes))
	for _, countryCode := range euCountryCodes {
		rules[countryCode] = map[product.DateType]Rule{
			product.DateTypeUseBy: {
				RegulatorGroup: "eu_1169_2011",
				CountryCode:    countryCode,
				DateType:       product.DateTypeUseBy,
				Status:         StatusResearchRequired,
				SourceURL:      sourceEU1169Article24,
				AccessedOn:     accessedOnRegistry,
			},
			product.DateTypeBestBefore: {
				RegulatorGroup: "eu_1169_2011",
				CountryCode:    countryCode,
				DateType:       product.DateTypeBestBefore,
				Status:         StatusResearchRequired,
				SourceURL:      sourceEU1169Article24,
				AccessedOn:     accessedOnRegistry,
			},
		}
	}
	return &Repository{rules: rules}
}

// RuleFor returns the registry row for a country and date type. The second
// return value is false only when the country is entirely unlisted; an
// unlisted or research_required row is never invented as "no automation" —
// callers must still check Status before automating anything.
func (repository *Repository) RuleFor(countryCode string, dateType product.DateType) (Rule, bool) {
	byDateType, ok := repository.rules[countryCode]
	if !ok {
		return Rule{}, false
	}
	rule, ok := byDateType[dateType]
	return rule, ok
}

// RegulatorGroupForCountry exposes only the reviewed group association used by
// the registry. An unlisted country intentionally has no group instead of a
// guessed legal classification.
func (repository *Repository) RegulatorGroupForCountry(countryCode string) string {
	byDateType, ok := repository.rules[countryCode]
	if !ok {
		return ""
	}
	for _, rule := range byDateType {
		return rule.RegulatorGroup
	}
	return ""
}
