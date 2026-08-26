package regulation

import (
	"errors"
	"time"

	"llm-homework/backend/internal/product"
)

// ErrAutomationUnavailable means the registry has no confirmed rule for this
// product's country and date type. Per the registry's evidence gate, no
// automatic status or notification schedule may be derived in this case.
var ErrAutomationUnavailable = errors.New("regulation: no confirmed rule for automation")

type Status string

const (
	StatusActive    Status = "active"
	StatusAttention Status = "attention"
	StatusExpired   Status = "expired"
)

// Evaluate computes a product's regulation-derived display status from a
// confirmed rule. It never inspects the product's country or date type to
// branch its logic — every distinction comes from the rule's own fields.
func Evaluate(item product.Product, rule Rule, now time.Time) (Status, error) {
	if rule.Status != StatusEnabled {
		return "", ErrAutomationUnavailable
	}

	expiryInstant, err := ExpiryInstant(item, rule)
	if err != nil {
		return "", err
	}

	if !now.After(expiryInstant) {
		return StatusActive, nil
	}
	if rule.PostExpiryStatus == PostExpiryExpired {
		return StatusExpired, nil
	}
	return StatusAttention, nil
}

// ExpiryInstant turns a product's calendar expiry date into the instant a
// confirmed rule designates, using only the rule's own timezone and
// instant-rule fields.
func ExpiryInstant(item product.Product, rule Rule) (time.Time, error) {
	location, err := time.LoadLocation(rule.ExpiryTimezone)
	if err != nil {
		return time.Time{}, err
	}

	local := item.ExpiryDate.In(location)
	switch rule.ExpiryInstantRule {
	case ExpiryAtStartOfDay:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location), nil
	case ExpiryAtEndOfDay:
		return time.Date(local.Year(), local.Month(), local.Day(), 23, 59, 59, 0, location), nil
	default:
		return time.Time{}, errors.New("regulation: rule has no expiry instant definition")
	}
}

// EligibleForRecipes reports whether a computed status still allows a
// product to appear in recipe suggestions.
func EligibleForRecipes(status Status) bool {
	return status != StatusExpired
}
