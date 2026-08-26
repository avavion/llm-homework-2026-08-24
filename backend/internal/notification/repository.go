package notification

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) ActiveWithCountry(ctx context.Context) ([]Candidate, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT
			p.id, p.account_id, p.name, p.date_type, p.expiry_date, p.quantity, p.unit,
			p.product_group, p.storage_location, p.country_code, p.alert_threshold_minutes,
			p.lifecycle_status, p.completed_at, p.created_at, p.updated_at,
			a.email_normalized
		FROM products p
		JOIN accounts a ON a.id = p.account_id
		WHERE p.lifecycle_status = 'active' AND p.country_code IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Candidate
	for rows.Next() {
		var candidate Candidate
		item := &candidate.Product
		if err := rows.Scan(
			&item.ID, &item.AccountID, &item.Name, &item.DateType, &item.ExpiryDate, &item.Quantity, &item.Unit,
			&item.ProductGroup, &item.StorageLocation, &item.CountryCode, &item.AlertThresholdMinutes,
			&item.LifecycleStatus, &item.CompletedAt, &item.CreatedAt, &item.UpdatedAt,
			&candidate.RecipientEmail,
		); err != nil {
			return nil, err
		}
		results = append(results, candidate)
	}
	return results, rows.Err()
}

func (repository *Repository) Claim(ctx context.Context, productID uuid.UUID, scheduledFor time.Time, channel string) (bool, error) {
	var id uuid.UUID
	err := repository.db.QueryRowContext(ctx, `
		INSERT INTO notification_deliveries (product_id, scheduled_for, channel)
		VALUES ($1, $2, $3)
		ON CONFLICT (product_id, scheduled_for, channel) DO NOTHING
		RETURNING id
	`, productID, scheduledFor, channel).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
