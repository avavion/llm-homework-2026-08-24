package product

import (
	"time"

	"github.com/google/uuid"
)

type DateType string

const (
	DateTypeUseBy      DateType = "use_by"
	DateTypeBestBefore DateType = "best_before"
)

func (dateType DateType) Valid() bool {
	return dateType == DateTypeUseBy || dateType == DateTypeBestBefore
}

// LifecycleStatus tracks the user-driven lifecycle of a product. Regulation-derived
// display states such as "attention" or "expired" are computed at read time by the
// regulation package and are never stored here.
type LifecycleStatus string

const (
	LifecycleActive    LifecycleStatus = "active"
	LifecycleUsed      LifecycleStatus = "used"
	LifecycleDiscarded LifecycleStatus = "discarded"
)

func (status LifecycleStatus) Terminal() bool {
	return status == LifecycleUsed || status == LifecycleDiscarded
}

type Product struct {
	ID                    uuid.UUID
	AccountID             uuid.UUID
	Name                  string
	DateType              DateType
	ExpiryDate            time.Time
	Quantity              *float64
	Unit                  *string
	ProductGroup          *string
	StorageLocation       *string
	CountryCode           *string
	AlertThresholdMinutes *int
	LifecycleStatus       LifecycleStatus
	CompletedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateInput struct {
	Name                  string
	DateType              DateType
	ExpiryDate            time.Time
	Quantity              *float64
	Unit                  *string
	ProductGroup          *string
	StorageLocation       *string
	CountryCode           *string
	AlertThresholdMinutes *int
}
