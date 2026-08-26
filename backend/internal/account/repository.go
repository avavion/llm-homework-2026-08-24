package account

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmailTaken = errors.New("email is already registered")
	ErrNotFound   = errors.New("account not found")
)

type Account struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type Record struct {
	Account
	PasswordHash string
}

type Session struct {
	AccountID uuid.UUID
	TokenHash []byte
	ExpiresAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) CreateAccount(ctx context.Context, email, passwordHash string) (Account, error) {
	var result Account
	err := repository.db.QueryRowContext(ctx, `
		INSERT INTO accounts (email_normalized, password_hash)
		VALUES ($1, $2)
		RETURNING id, email_normalized
	`, email, passwordHash).Scan(&result.ID, &result.Email)
	if isUniqueViolation(err) {
		return Account{}, ErrEmailTaken
	}
	return result, err
}

func (repository *Repository) AccountByEmail(ctx context.Context, email string) (Record, error) {
	var result Record
	err := repository.db.QueryRowContext(ctx, `
		SELECT id, email_normalized, password_hash
		FROM accounts
		WHERE email_normalized = $1
	`, email).Scan(&result.ID, &result.Email, &result.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	return result, err
}

func (repository *Repository) CreateSession(ctx context.Context, session Session) error {
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (account_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, session.AccountID, session.TokenHash, session.ExpiresAt)
	return err
}

func (repository *Repository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := repository.db.ExecContext(ctx, `
		DELETE FROM auth_sessions
		WHERE token_hash = $1
	`, tokenHash)
	return err
}

func (repository *Repository) AccountBySession(ctx context.Context, tokenHash []byte, now time.Time) (Account, error) {
	var result Account
	err := repository.db.QueryRowContext(ctx, `
		SELECT accounts.id, accounts.email_normalized
		FROM auth_sessions
		JOIN accounts ON accounts.id = auth_sessions.account_id
		WHERE auth_sessions.token_hash = $1
		  AND auth_sessions.expires_at > $2
	`, tokenHash, now).Scan(&result.ID, &result.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	return result, err
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}
