// Package regulation implements the registry-driven date rules described in
// shared/docs/regulatory-date-rules.md. Automation is only allowed for a rule
// whose Status is Enabled; every other rule leaves status and scheduling to
// the user, per the registry's evidence gate.
package regulation

import "llm-homework/backend/internal/product"

type RuleStatus string

const (
	StatusEnabled          RuleStatus = "enabled"
	StatusResearchRequired RuleStatus = "research_required"
)

// PostExpiryStatus is the regulator-confirmed display status a product takes
// on once its expiry instant has passed.
type PostExpiryStatus string

const (
	PostExpiryExpired   PostExpiryStatus = "expired"
	PostExpiryAttention PostExpiryStatus = "attention"
)

// ExpiryInstantRule names how a confirmed rule turns a calendar date into an
// instant. It is data, supplied per rule row, never chosen per country in code.
type ExpiryInstantRule string

const (
	ExpiryAtStartOfDay ExpiryInstantRule = "start_of_day"
	ExpiryAtEndOfDay   ExpiryInstantRule = "end_of_day"
)

// Rule is one row of the regulatory registry, scoped to a single country and
// date type.
type Rule struct {
	RegulatorGroup    string
	CountryCode       string
	DateType          product.DateType
	Status            RuleStatus
	ExpiryTimezone    string // IANA zone name; required when Status == StatusEnabled
	ExpiryInstantRule ExpiryInstantRule
	PostExpiryStatus  PostExpiryStatus
	SourceURL         string
	AccessedOn        string
}
