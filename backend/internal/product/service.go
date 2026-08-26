package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// MinAlertThresholdMinutes matches the product-group alert policy: a user
// threshold below one hour is rejected outright.
const MinAlertThresholdMinutes = 60

var (
	ErrInvalidProduct    = errors.New("name, date type, and date are required")
	ErrInvalidDateType   = errors.New("date type must be use_by or best_before")
	ErrThresholdTooSmall = errors.New("alert threshold must be at least 60 minutes")
	ErrAlreadyCompleted  = errors.New("product is already used or discarded")
	ErrInvalidStatus     = errors.New("status must be used or discarded")
)

// Store is the persistence boundary Service depends on, so tests can
// substitute an in-memory fake instead of a database.
type Store interface {
	Create(ctx context.Context, accountID uuid.UUID, input CreateInput) (Product, error)
	Get(ctx context.Context, accountID, productID uuid.UUID) (Product, error)
	List(ctx context.Context, accountID uuid.UUID) ([]Product, error)
	Complete(ctx context.Context, accountID, productID uuid.UUID, status LifecycleStatus, completedAt time.Time) (Product, error)
}

type Service struct {
	repository Store
	now        func() time.Time
}

func NewService(repository Store) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (service *Service) Create(ctx context.Context, accountID uuid.UUID, input CreateInput) (Product, error) {
	if err := ValidateCreate(input); err != nil {
		return Product{}, err
	}
	return service.repository.Create(ctx, accountID, input)
}

func (service *Service) Get(ctx context.Context, accountID, productID uuid.UUID) (Product, error) {
	return service.repository.Get(ctx, accountID, productID)
}

func (service *Service) List(ctx context.Context, accountID uuid.UUID) ([]Product, error) {
	return service.repository.List(ctx, accountID)
}

func (service *Service) Complete(ctx context.Context, accountID, productID uuid.UUID, status LifecycleStatus) (Product, error) {
	if status != LifecycleUsed && status != LifecycleDiscarded {
		return Product{}, ErrInvalidStatus
	}

	result, err := service.repository.Complete(ctx, accountID, productID, status, service.now())
	if errors.Is(err, ErrNotFound) {
		// Distinguish "missing" from "already completed" so handlers can
		// return the right conflict status instead of a bare 404.
		if _, getErr := service.repository.Get(ctx, accountID, productID); getErr == nil {
			return Product{}, ErrAlreadyCompleted
		}
		return Product{}, ErrNotFound
	}
	return result, err
}

// ValidateCreate applies the same field rules Create does, exported so other
// modules (e.g. draft approval) can validate before opening a transaction.
func ValidateCreate(input CreateInput) error {
	if input.Name == "" || input.DateType == "" || input.ExpiryDate.IsZero() {
		return ErrInvalidProduct
	}
	if !input.DateType.Valid() {
		return ErrInvalidDateType
	}
	if input.AlertThresholdMinutes != nil && *input.AlertThresholdMinutes < MinAlertThresholdMinutes {
		return ErrThresholdTooSmall
	}
	return nil
}
