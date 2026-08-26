package recognition

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"llm-homework/backend/internal/product"
)

var (
	ErrNotFound       = errors.New("recognition: draft not found")
	ErrAlreadyDecided = errors.New("recognition: draft has already been approved or rejected")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, accountID uuid.UUID, fields DraftFields, rawText, sourceReference string) (ProductDraft, error) {
	var result ProductDraft
	err := repository.db.QueryRowContext(ctx, `
		INSERT INTO product_drafts (
			account_id, source_reference, raw_text, name, date_type, expiry_date,
			quantity, unit, product_group, storage_location, country_code, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'pending')
		RETURNING
			id, account_id, source_reference, raw_text, name, date_type, expiry_date,
			quantity, unit, product_group, storage_location, country_code,
			status, approved_product_id, created_at, updated_at
	`,
		accountID, sourceReference, rawText, fields.Name, fields.DateType, fields.ExpiryDate,
		fields.Quantity, fields.Unit, fields.ProductGroup, fields.StorageLocation, fields.CountryCode,
	).Scan(scanTargets(&result)...)
	return result, err
}

func (repository *Repository) Get(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	var result ProductDraft
	err := repository.db.QueryRowContext(ctx, `
		SELECT
			id, account_id, source_reference, raw_text, name, date_type, expiry_date,
			quantity, unit, product_group, storage_location, country_code,
			status, approved_product_id, created_at, updated_at
		FROM product_drafts
		WHERE id = $1 AND account_id = $2
	`, draftID, accountID).Scan(scanTargets(&result)...)
	if errors.Is(err, sql.ErrNoRows) {
		return ProductDraft{}, ErrNotFound
	}
	return result, err
}

// Reject marks a pending draft rejected. It only updates a row that is still
// pending, so a second decision on the same draft affects zero rows.
func (repository *Repository) Reject(ctx context.Context, accountID, draftID uuid.UUID) (ProductDraft, error) {
	var result ProductDraft
	err := repository.db.QueryRowContext(ctx, `
		UPDATE product_drafts
		SET status = 'rejected', updated_at = now()
		WHERE id = $1 AND account_id = $2 AND status = 'pending'
		RETURNING
			id, account_id, source_reference, raw_text, name, date_type, expiry_date,
			quantity, unit, product_group, storage_location, country_code,
			status, approved_product_id, created_at, updated_at
	`, draftID, accountID).Scan(scanTargets(&result)...)
	if errors.Is(err, sql.ErrNoRows) {
		if _, getErr := repository.Get(ctx, accountID, draftID); getErr != nil {
			return ProductDraft{}, ErrNotFound
		}
		return ProductDraft{}, ErrAlreadyDecided
	}
	return result, err
}

// Approve locks the draft row, inserts the product on the same transaction,
// and marks the draft approved — all atomically, so a concurrent second
// approval never creates a second product.
func (repository *Repository) Approve(ctx context.Context, accountID, draftID uuid.UUID, input product.CreateInput) (product.Product, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return product.Product{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT status FROM product_drafts WHERE id = $1 AND account_id = $2 FOR UPDATE
	`, draftID, accountID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return product.Product{}, ErrNotFound
	}
	if err != nil {
		return product.Product{}, err
	}
	if DraftStatus(status) != DraftPending {
		return product.Product{}, ErrAlreadyDecided
	}

	created, err := product.WithTx(tx).Create(ctx, accountID, input)
	if err != nil {
		return product.Product{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE product_drafts SET status = 'approved', approved_product_id = $1, updated_at = now()
		WHERE id = $2
	`, created.ID, draftID); err != nil {
		return product.Product{}, err
	}

	if err := tx.Commit(); err != nil {
		return product.Product{}, err
	}
	return created, nil
}

func scanTargets(draft *ProductDraft) []any {
	return []any{
		&draft.ID, &draft.AccountID, &draft.SourceReference, &draft.RawText,
		&draft.Fields.Name, &draft.Fields.DateType, &draft.Fields.ExpiryDate,
		&draft.Fields.Quantity, &draft.Fields.Unit, &draft.Fields.ProductGroup,
		&draft.Fields.StorageLocation, &draft.Fields.CountryCode,
		&draft.Status, &draft.ApprovedProductID, &draft.CreatedAt, &draft.UpdatedAt,
	}
}
