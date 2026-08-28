package regulation

import (
	"testing"

	"llm-homework/backend/internal/product"
)

func TestEAEUCountriesAreEnabledAsADemoAssumption(t *testing.T) {
	repository := NewRepository()

	for _, countryCode := range []string{"AM", "BY", "KZ", "KG", "RU"} {
		for _, dateType := range []product.DateType{product.DateTypeUseBy, product.DateTypeBestBefore} {
			rule, ok := repository.RuleFor(countryCode, dateType)
			if !ok {
				t.Fatalf("RuleFor(%s, %s) missing", countryCode, dateType)
			}
			if rule.Status != StatusEnabled {
				t.Fatalf("RuleFor(%s, %s).Status = %v, want enabled", countryCode, dateType, rule.Status)
			}
			if rule.ExpiryTimezone == "" || rule.ExpiryInstantRule == "" {
				t.Fatalf("RuleFor(%s, %s) = %+v, want a usable timezone and instant rule", countryCode, dateType, rule)
			}
		}
	}
}

func TestWiderCISGroupIsNotIndexed(t *testing.T) {
	repository := NewRepository()

	for _, countryCode := range []string{"AZ", "MD", "TJ", "TM", "UA", "UZ"} {
		if _, ok := repository.RuleFor(countryCode, product.DateTypeUseBy); ok {
			t.Fatalf("RuleFor(%s, use_by) unexpectedly found a rule; that group has no confirmed date_type mapping, demo or otherwise", countryCode)
		}
	}
}
