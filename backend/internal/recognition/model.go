package recognition

import (
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

type DraftStatus string

const (
	DraftPending  DraftStatus = "pending"
	DraftApproved DraftStatus = "approved"
	DraftRejected DraftStatus = "rejected"
)

// DraftFields are the recognized values a user reviews and edits before
// approval. Every field is optional: OCR/LLM output is a suggestion, never a
// substitute for the user's own confirmation.
type DraftFields struct {
	Name            *string
	DateType        *product.DateType
	ExpiryDate      *time.Time
	Quantity        *float64
	Unit            *string
	ProductGroup    *string
	StorageLocation *string
	CountryCode     *string
}

type ProductDraft struct {
	ID                uuid.UUID
	AccountID         uuid.UUID
	SourceReference   string
	RawText           string
	Fields            DraftFields
	Status            DraftStatus
	ApprovedProductID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
