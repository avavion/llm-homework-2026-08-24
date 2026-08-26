package recognition

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

var (
	ErrEmptyImage             = errors.New("recognition: image is required")
	ErrRecognitionUnavailable = errors.New("recognition: recognition is unavailable, use the manual form")
)

// DraftRepository is the subset of Repository the service depends on.
type DraftRepository interface {
	Create(ctx context.Context, accountID uuid.UUID, fields DraftFields, rawText, sourceReference string) (ProductDraft, error)
	Get(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error)
	Reject(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error)
	Approve(ctx context.Context, accountID, draftID uuid.UUID, input product.CreateInput) (product.Product, error)
}

type Service struct {
	repository DraftRepository
	ocr        OCRClient
	llm        LLMClient
}

func NewService(repository DraftRepository, ocr OCRClient, llm LLMClient) *Service {
	return &Service{repository: repository, ocr: ocr, llm: llm}
}

// Recognize extracts draft fields from a photo and stores them as a pending
// draft. It never touches the products table: only Approve may create one.
// Any OCR/LLM failure collapses to ErrRecognitionUnavailable so provider
// details never reach the client.
func (service *Service) Recognize(ctx context.Context, accountID uuid.UUID, image []byte, locale, sourceReference string) (ProductDraft, error) {
	if len(image) == 0 {
		return ProductDraft{}, ErrEmptyImage
	}

	text, err := service.ocr.ExtractText(ctx, image)
	if err != nil {
		return ProductDraft{}, ErrRecognitionUnavailable
	}

	fields, err := service.llm.ParseProduct(ctx, text, locale)
	if err != nil {
		return ProductDraft{}, ErrRecognitionUnavailable
	}

	return service.repository.Create(ctx, accountID, fields, text, sourceReference)
}

func (service *Service) Get(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	return service.repository.Get(ctx, accountID, draftID)
}

// Approve validates the user-edited fields, then atomically creates one
// product and marks the draft approved.
func (service *Service) Approve(ctx context.Context, accountID, draftID uuid.UUID, edited product.CreateInput) (product.Product, error) {
	if err := product.ValidateCreate(edited); err != nil {
		return product.Product{}, err
	}

	return service.repository.Approve(ctx, accountID, draftID, edited)
}

func (service *Service) Reject(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	return service.repository.Reject(ctx, accountID, draftID)
}
