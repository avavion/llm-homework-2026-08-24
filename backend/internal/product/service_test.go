package product

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memoryRepository struct {
	items map[uuid.UUID]Product
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{items: map[uuid.UUID]Product{}}
}

func (repository *memoryRepository) Create(_ context.Context, accountID uuid.UUID, input CreateInput) (Product, error) {
	now := time.Now()
	item := Product{
		ID:                    uuid.New(),
		AccountID:             accountID,
		Name:                  input.Name,
		DateType:              input.DateType,
		ExpiryDate:            input.ExpiryDate,
		Quantity:              input.Quantity,
		Unit:                  input.Unit,
		ProductGroup:          input.ProductGroup,
		StorageLocation:       input.StorageLocation,
		CountryCode:           input.CountryCode,
		AlertThresholdMinutes: input.AlertThresholdMinutes,
		LifecycleStatus:       LifecycleActive,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	repository.items[item.ID] = item
	return item, nil
}

func (repository *memoryRepository) Get(_ context.Context, accountID, productID uuid.UUID) (Product, error) {
	item, ok := repository.items[productID]
	if !ok || item.AccountID != accountID {
		return Product{}, ErrNotFound
	}
	return item, nil
}

func (repository *memoryRepository) List(_ context.Context, accountID uuid.UUID) ([]Product, error) {
	var results []Product
	for _, item := range repository.items {
		if item.AccountID == accountID {
			results = append(results, item)
		}
	}
	return results, nil
}

func (repository *memoryRepository) Complete(_ context.Context, accountID, productID uuid.UUID, status LifecycleStatus, completedAt time.Time) (Product, error) {
	item, ok := repository.items[productID]
	if !ok || item.AccountID != accountID || item.LifecycleStatus != LifecycleActive {
		return Product{}, ErrNotFound
	}
	item.LifecycleStatus = status
	item.CompletedAt = &completedAt
	repository.items[productID] = item
	return item, nil
}

func TestCreateRequiresNameDateTypeAndDate(t *testing.T) {
	service := NewService(newMemoryRepository())

	_, err := service.Create(context.Background(), uuid.New(), CreateInput{})
	if !errors.Is(err, ErrInvalidProduct) {
		t.Fatalf("err = %v, want ErrInvalidProduct", err)
	}
}

func TestCreateRejectsUnknownDateType(t *testing.T) {
	service := NewService(newMemoryRepository())

	_, err := service.Create(context.Background(), uuid.New(), CreateInput{
		Name: "Milk", DateType: "sell_by", ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if !errors.Is(err, ErrInvalidDateType) {
		t.Fatalf("err = %v, want ErrInvalidDateType", err)
	}
}

func TestCreateAllowsOptionalFields(t *testing.T) {
	service := NewService(newMemoryRepository())
	quantity := 2.0
	unit := "l"

	created, err := service.Create(context.Background(), uuid.New(), CreateInput{
		Name: "Milk", DateType: DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
		Quantity: &quantity, Unit: &unit,
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if created.LifecycleStatus != LifecycleActive {
		t.Fatalf("status = %v, want active", created.LifecycleStatus)
	}
}

func TestCreateRejectsThresholdBelowSixtyMinutes(t *testing.T) {
	service := NewService(newMemoryRepository())
	threshold := 59

	_, err := service.Create(context.Background(), uuid.New(), CreateInput{
		Name: "Milk", DateType: DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
		AlertThresholdMinutes: &threshold,
	})
	if !errors.Is(err, ErrThresholdTooSmall) {
		t.Fatalf("err = %v, want ErrThresholdTooSmall", err)
	}
}

func TestUsedAndDiscardedAreDistinctTerminalStates(t *testing.T) {
	service := NewService(newMemoryRepository())
	accountID := uuid.New()
	created, err := service.Create(context.Background(), accountID, CreateInput{
		Name: "Milk", DateType: DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create err = %v", err)
	}

	used, err := service.Complete(context.Background(), accountID, created.ID, LifecycleUsed)
	if err != nil {
		t.Fatalf("complete err = %v", err)
	}
	if used.LifecycleStatus != LifecycleUsed || used.CompletedAt == nil {
		t.Fatalf("used = %+v, want status used with CompletedAt set", used)
	}

	other, err := service.Create(context.Background(), accountID, CreateInput{
		Name: "Bread", DateType: DateTypeBestBefore, ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create err = %v", err)
	}
	discarded, err := service.Complete(context.Background(), accountID, other.ID, LifecycleDiscarded)
	if err != nil {
		t.Fatalf("complete err = %v", err)
	}
	if discarded.LifecycleStatus != LifecycleDiscarded || discarded.CompletedAt == nil {
		t.Fatalf("discarded = %+v, want status discarded with CompletedAt set", discarded)
	}
}

func TestCompleteRejectsRepeatedFinalAction(t *testing.T) {
	service := NewService(newMemoryRepository())
	accountID := uuid.New()
	created, err := service.Create(context.Background(), accountID, CreateInput{
		Name: "Milk", DateType: DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create err = %v", err)
	}

	if _, err := service.Complete(context.Background(), accountID, created.ID, LifecycleUsed); err != nil {
		t.Fatalf("first complete err = %v", err)
	}
	if _, err := service.Complete(context.Background(), accountID, created.ID, LifecycleDiscarded); !errors.Is(err, ErrAlreadyCompleted) {
		t.Fatalf("second complete err = %v, want ErrAlreadyCompleted", err)
	}
}

func TestForeignAccountCannotReadOrCompleteProduct(t *testing.T) {
	service := NewService(newMemoryRepository())
	owner := uuid.New()
	stranger := uuid.New()
	created, err := service.Create(context.Background(), owner, CreateInput{
		Name: "Milk", DateType: DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create err = %v", err)
	}

	if _, err := service.Get(context.Background(), stranger, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get err = %v, want ErrNotFound", err)
	}
	if _, err := service.Complete(context.Background(), stranger, created.ID, LifecycleUsed); !errors.Is(err, ErrNotFound) {
		t.Fatalf("complete err = %v, want ErrNotFound", err)
	}
}
