package product

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("product not found")

// dbtx is satisfied by both *sql.DB and *sql.Tx so the repository can run
// standalone or as part of a caller-managed transaction (see WithTx).
type dbtx interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Repository struct {
	db dbtx
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// WithTx returns a repository bound to the given transaction, letting callers
// (e.g. the recognition draft-approval flow) create a product atomically
// alongside other writes.
func WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: tx}
}

func (repository *Repository) Create(ctx context.Context, accountID uuid.UUID, input CreateInput) (Product, error) {
	var result Product
	err := repository.db.QueryRowContext(ctx, `
		INSERT INTO products (
			account_id, name, date_type, expiry_date, quantity, unit,
			product_group, storage_location, country_code, alert_threshold_minutes,
			lifecycle_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id, account_id, name, date_type, expiry_date, quantity, unit,
			product_group, storage_location, country_code, alert_threshold_minutes,
			lifecycle_status, completed_at, created_at, updated_at
	`,
		accountID, input.Name, string(input.DateType), input.ExpiryDate, input.Quantity, input.Unit,
		input.ProductGroup, input.StorageLocation, input.CountryCode, input.AlertThresholdMinutes,
		string(LifecycleActive),
	).Scan(scanTargets(&result)...)
	return result, err
}

func (repository *Repository) Get(ctx context.Context, accountID, productID uuid.UUID) (Product, error) {
	var result Product
	err := repository.db.QueryRowContext(ctx, `
		SELECT
			id, account_id, name, date_type, expiry_date, quantity, unit,
			product_group, storage_location, country_code, alert_threshold_minutes,
			lifecycle_status, completed_at, created_at, updated_at
		FROM products
		WHERE id = $1 AND account_id = $2
	`, productID, accountID).Scan(scanTargets(&result)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return result, err
}

func (repository *Repository) List(ctx context.Context, accountID uuid.UUID) ([]Product, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT
			id, account_id, name, date_type, expiry_date, quantity, unit,
			product_group, storage_location, country_code, alert_threshold_minutes,
			lifecycle_status, completed_at, created_at, updated_at
		FROM products
		WHERE account_id = $1
		ORDER BY created_at DESC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Product
	for rows.Next() {
		var item Product
		if err := rows.Scan(scanTargets(&item)...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

// Complete transitions a product to a terminal lifecycle status. It only
// updates rows that are still active, so a repeated call on an already
// completed product affects zero rows and the caller can report a conflict.
func (repository *Repository) Complete(ctx context.Context, accountID, productID uuid.UUID, status LifecycleStatus, completedAt time.Time) (Product, error) {
	var result Product
	err := repository.db.QueryRowContext(ctx, `
		UPDATE products
		SET lifecycle_status = $1, completed_at = $2, updated_at = now()
		WHERE id = $3 AND account_id = $4 AND lifecycle_status = $5
		RETURNING
			id, account_id, name, date_type, expiry_date, quantity, unit,
			product_group, storage_location, country_code, alert_threshold_minutes,
			lifecycle_status, completed_at, created_at, updated_at
	`, string(status), completedAt, productID, accountID, string(LifecycleActive)).Scan(scanTargets(&result)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return result, err
}

func scanTargets(product *Product) []any {
	return []any{
		&product.ID, &product.AccountID, &product.Name, &product.DateType, &product.ExpiryDate,
		&product.Quantity, &product.Unit, &product.ProductGroup, &product.StorageLocation,
		&product.CountryCode, &product.AlertThresholdMinutes, &product.LifecycleStatus,
		&product.CompletedAt, &product.CreatedAt, &product.UpdatedAt,
	}
}
