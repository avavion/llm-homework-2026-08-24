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
	sourceTRCU0222011     = "https://eec.eaeunion.org/upload/medialibrary/9db/TrTsPishevkaMarkirovka.pdf"
	accessedOnRegistry    = "2026-08-26"
)

// eaeuTimezoneByCountry is a demo-mode assumption, not a regulator-confirmed
// timezone: TR CU 022/2011 section 4.7 requires an hour on short-shelf-life
// labels but names no single timezone for the group. Each entry is that
// country's principal civil IANA zone, picked as a product decision to unblock
// a working demo — see "Демо-режим" in shared/docs/regulatory-date-rules.md.
var eaeuTimezoneByCountry = map[string]string{
	"AM": "Asia/Yerevan",
	"BY": "Europe/Minsk",
	"KZ": "Asia/Almaty",
	"KG": "Asia/Bishkek",
	"RU": "Europe/Moscow",
}

// Repository is a read-only, in-memory copy of the reviewed registry: EU rows
// change only through the documented legal/QA review process, never at
// runtime, so there is no database table behind them. The AM/BY/KZ/KG/RU rows
// are the one documented exception — a product decision to enable a demo
// without that review; see "Демо-режим" in
// shared/docs/regulatory-date-rules.md for exactly what is and isn't
// verified there.
//
// Only use_by/best_before rows are indexed here because those are the only
// product.DateType values the current product schema models. The unverified
// wider-CIS group from the shared registry (AZ, MD, TJ, TM, UA, UZ) has no
// confirmed date_type mapping at all — not even a demo assumption — so it is
// intentionally left out of the queryable index.
type Repository struct {
	rules map[string]map[product.DateType]Rule
}

func NewRepository() *Repository {
	rules := make(map[string]map[product.DateType]Rule, len(euCountryCodes)+len(eaeuTimezoneByCountry))
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
	// Demo-mode assumption for AM/BY/KZ/KG/RU: enabled on a product decision
	// to unblock a working demo, not because TR CU 022/2011 clears the
	// evidence gate (it names no timezone/instant rule and does not map
	// shelf-life labels onto use_by/best_before). See "Демо-режим" in
	// shared/docs/regulatory-date-rules.md before treating this as legal
	// clearance.
	for countryCode, timezone := range eaeuTimezoneByCountry {
		rules[countryCode] = map[product.DateType]Rule{
			product.DateTypeUseBy: {
				RegulatorGroup:    "eaeu_tr_ts_022_2011",
				CountryCode:       countryCode,
				DateType:          product.DateTypeUseBy,
				Status:            StatusEnabled,
				ExpiryTimezone:    timezone,
				ExpiryInstantRule: ExpiryAtEndOfDay,
				PostExpiryStatus:  PostExpiryExpired,
				SourceURL:         sourceTRCU0222011,
				AccessedOn:        accessedOnRegistry,
			},
			product.DateTypeBestBefore: {
				RegulatorGroup:    "eaeu_tr_ts_022_2011",
				CountryCode:       countryCode,
				DateType:          product.DateTypeBestBefore,
				Status:            StatusEnabled,
				ExpiryTimezone:    timezone,
				ExpiryInstantRule: ExpiryAtEndOfDay,
				PostExpiryStatus:  PostExpiryAttention,
				SourceURL:         sourceTRCU0222011,
				AccessedOn:        accessedOnRegistry,
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
