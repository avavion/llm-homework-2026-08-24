package recognition

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

type memoryRepository struct {
	drafts   map[uuid.UUID]ProductDraft
	products map[uuid.UUID]product.Product
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{drafts: map[uuid.UUID]ProductDraft{}, products: map[uuid.UUID]product.Product{}}
}

func (repository *memoryRepository) Create(_ context.Context, accountID uuid.UUID, fields DraftFields, rawText, sourceReference string) (ProductDraft, error) {
	draft := ProductDraft{
		ID: uuid.New(), AccountID: accountID, Fields: fields, RawText: rawText,
		SourceReference: sourceReference, Status: DraftPending, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repository.drafts[draft.ID] = draft
	return draft, nil
}

func (repository *memoryRepository) Get(_ context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	draft, ok := repository.drafts[draftID]
	if !ok || draft.AccountID != accountID {
		return ProductDraft{}, ErrNotFound
	}
	return draft, nil
}

func (repository *memoryRepository) Reject(_ context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	draft, ok := repository.drafts[draftID]
	if !ok || draft.AccountID != accountID {
		return ProductDraft{}, ErrNotFound
	}
	if draft.Status != DraftPending {
		return ProductDraft{}, ErrAlreadyDecided
	}
	draft.Status = DraftRejected
	repository.drafts[draftID] = draft
	return draft, nil
}

func (repository *memoryRepository) Approve(_ context.Context, accountID, draftID uuid.UUID, input product.CreateInput) (product.Product, error) {
	draft, ok := repository.drafts[draftID]
	if !ok || draft.AccountID != accountID {
		return product.Product{}, ErrNotFound
	}
	if draft.Status != DraftPending {
		return product.Product{}, ErrAlreadyDecided
	}

	created := product.Product{
		ID: uuid.New(), AccountID: accountID, Name: input.Name, DateType: input.DateType,
		ExpiryDate: input.ExpiryDate, LifecycleStatus: product.LifecycleActive,
	}
	repository.products[created.ID] = created

	draft.Status = DraftApproved
	draft.ApprovedProductID = &created.ID
	repository.drafts[draftID] = draft
	return created, nil
}

func (repository *memoryRepository) productCount(accountID uuid.UUID) int {
	count := 0
	for _, item := range repository.products {
		if item.AccountID == accountID {
			count++
		}
	}
	return count
}

type fakeOCR struct {
	text string
	err  error
}

func (ocr fakeOCR) ExtractText(context.Context, []byte) (string, error) {
	return ocr.text, ocr.err
}

type fakeLLM struct {
	fields DraftFields
	err    error
}

func (llm fakeLLM) ParseProduct(context.Context, string, string) (DraftFields, error) {
	return llm.fields, llm.err
}

func TestRecognizeCreatesDraftNotProduct(t *testing.T) {
	repository := newMemoryRepository()
	name := "Milk"
	service := NewService(repository, fakeOCR{text: "MILK best before 2026-03-01"}, fakeLLM{fields: DraftFields{Name: &name}})
	accountID := uuid.New()

	draft, err := service.Recognize(context.Background(), accountID, []byte("image-bytes"), "en", "photo.jpg")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if draft.Status != DraftPending {
		t.Fatalf("status = %v, want pending", draft.Status)
	}
	if repository.productCount(accountID) != 0 {
		t.Fatal("product created before approval")
	}
}

func TestRecognizeFailsClosedWithoutLeakingProviderDetails(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository, fakeOCR{err: errors.New("provider api key rejected: sk-secret-123")}, fakeLLM{})

	_, err := service.Recognize(context.Background(), uuid.New(), []byte("image-bytes"), "en", "photo.jpg")
	if !errors.Is(err, ErrRecognitionUnavailable) {
		t.Fatalf("err = %v, want ErrRecognitionUnavailable", err)
	}
}

func TestRecognizeRejectsEmptyImage(t *testing.T) {
	service := NewService(newMemoryRepository(), fakeOCR{}, fakeLLM{})

	_, err := service.Recognize(context.Background(), uuid.New(), nil, "en", "photo.jpg")
	if !errors.Is(err, ErrEmptyImage) {
		t.Fatalf("err = %v, want ErrEmptyImage", err)
	}
}

func TestApproveCreatesOneProductAndRejectCreatesNone(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository, fakeOCR{}, fakeLLM{})
	owner := uuid.New()

	draft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo.jpg")
	created, err := service.Approve(context.Background(), owner, draft.ID, product.CreateInput{
		Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("approve err = %v", err)
	}
	if repository.productCount(owner) != 1 || created.Name != "Milk" {
		t.Fatalf("created = %+v, count = %d", created, repository.productCount(owner))
	}

	anotherDraft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo2.jpg")
	if _, err := service.Reject(context.Background(), owner, anotherDraft.ID); err != nil {
		t.Fatalf("reject err = %v", err)
	}
	if repository.productCount(owner) != 1 {
		t.Fatalf("count = %d, want 1 (reject must not create a product)", repository.productCount(owner))
	}
}

func TestForeignAccountCannotDecideOnDraft(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository, fakeOCR{}, fakeLLM{})
	owner := uuid.New()
	stranger := uuid.New()

	draft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo.jpg")
	if _, err := service.Approve(context.Background(), stranger, draft.ID, product.CreateInput{
		Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour),
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDoubleApprovalIsRejected(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository, fakeOCR{}, fakeLLM{})
	owner := uuid.New()

	draft, _ := repository.Create(context.Background(), owner, DraftFields{}, "text", "photo.jpg")
	input := product.CreateInput{Name: "Milk", DateType: product.DateTypeUseBy, ExpiryDate: time.Now().Add(24 * time.Hour)}

	if _, err := service.Approve(context.Background(), owner, draft.ID, input); err != nil {
		t.Fatalf("first approve err = %v", err)
	}
	if _, err := service.Approve(context.Background(), owner, draft.ID, input); !errors.Is(err, ErrAlreadyDecided) {
		t.Fatalf("second approve err = %v, want ErrAlreadyDecided", err)
	}
	if repository.productCount(owner) != 1 {
		t.Fatalf("count = %d, want 1", repository.productCount(owner))
	}
}
