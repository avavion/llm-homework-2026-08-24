package profile

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Profile returns an empty profile when onboarding has not yet completed. This
// lets existing accounts complete the new required setup without receiving a
// resource-existence error that would leak implementation details to clients.
func (repository *Repository) Profile(ctx context.Context, accountID uuid.UUID) (Profile, error) {
	var result Profile
	err := repository.db.QueryRowContext(ctx, `
		SELECT country_code, language
		FROM account_profiles
		WHERE account_id = $1
	`, accountID).Scan(&result.CountryCode, &result.Language)
	if err == sql.ErrNoRows {
		return Profile{}, nil
	}
	return result, err
}

func (repository *Repository) SaveProfile(ctx context.Context, accountID uuid.UUID, input ProfileInput) (Profile, error) {
	var result Profile
	err := repository.db.QueryRowContext(ctx, `
		INSERT INTO account_profiles (account_id, country_code, language)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id) DO UPDATE SET
			country_code = EXCLUDED.country_code,
			language = EXCLUDED.language,
			updated_at = now()
		RETURNING country_code, language
	`, accountID, input.CountryCode, input.Language).Scan(&result.CountryCode, &result.Language)
	return result, err
}

func (repository *Repository) Settings(ctx context.Context, accountID uuid.UUID) (map[string]int, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT product_group, alert_threshold_minutes
		FROM account_notification_settings
		WHERE account_id = $1
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var group string
		var threshold int
		if err := rows.Scan(&group, &threshold); err != nil {
			return nil, err
		}
		result[group] = threshold
	}
	return result, rows.Err()
}

func (repository *Repository) SaveSetting(ctx context.Context, accountID uuid.UUID, productGroup string, thresholdMinutes int) error {
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO account_notification_settings (account_id, product_group, alert_threshold_minutes)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id, product_group) DO UPDATE SET
			alert_threshold_minutes = EXCLUDED.alert_threshold_minutes,
			updated_at = now()
	`, accountID, productGroup, thresholdMinutes)
	return err
}
