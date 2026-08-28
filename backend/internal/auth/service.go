package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"llm-homework/backend/internal/account"
)

const (
	sessionTokenBytes = 32
	sessionLifetime   = 30 * 24 * time.Hour
)

var (
	ErrEmailTaken          = errors.New("email is already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRegistration = errors.New("email and password are required")
	ErrUnauthenticated     = errors.New("authentication required")
)

type Account = account.Account

type Repository interface {
	CreateAccount(ctx context.Context, email, passwordHash string) (account.Account, error)
	AccountByEmail(ctx context.Context, email string) (account.Record, error)
	CreateSession(ctx context.Context, session account.Session) error
	DeleteSession(ctx context.Context, tokenHash []byte) error
	AccountBySession(ctx context.Context, tokenHash []byte, now time.Time) (account.Account, error)
}

// ProfileInitializer gives a fresh account its default profile. The country
// selector is not built yet (shared/docs decisions), so every new account
// starts on the EAEU default rather than being blocked until they visit
// Settings; ProfileInitializer alone decides what that default is.
type ProfileInitializer interface {
	InitializeProfile(ctx context.Context, accountID uuid.UUID) error
}

type Service struct {
	repository Repository
	profiles   ProfileInitializer
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// WithProfileInitializer wires a default-profile bootstrapper into
// registration. It returns the service to allow chaining at construction.
func (service *Service) WithProfileInitializer(profiles ProfileInitializer) *Service {
	service.profiles = profiles
	return service
}

func (service *Service) Register(ctx context.Context, email, password string) (Account, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return Account{}, ErrInvalidRegistration
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return Account{}, err
	}
	result, err := service.repository.CreateAccount(ctx, email, passwordHash)
	if errors.Is(err, account.ErrEmailTaken) {
		return Account{}, ErrEmailTaken
	}
	if err != nil {
		return Account{}, err
	}

	if service.profiles != nil {
		if err := service.profiles.InitializeProfile(ctx, result.ID); err != nil {
			return Account{}, err
		}
	}

	return result, nil
}

func (service *Service) Login(ctx context.Context, email, password string) (string, Account, error) {
	record, err := service.repository.AccountByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, account.ErrNotFound) {
		// Perform the same expensive operation as a password check so an unknown
		// e-mail does not get a cheap, externally distinguishable path.
		_, _ = hashPassword(password)
		return "", Account{}, ErrInvalidCredentials
	}
	if err != nil {
		return "", Account{}, err
	}

	valid, err := verifyPassword(password, record.PasswordHash)
	if err != nil {
		return "", Account{}, fmt.Errorf("verify stored password hash: %w", err)
	}
	if !valid {
		return "", Account{}, ErrInvalidCredentials
	}

	rawToken, err := newSessionToken()
	if err != nil {
		return "", Account{}, err
	}
	if err := service.repository.CreateSession(ctx, account.Session{
		AccountID: record.ID,
		TokenHash: hashSessionToken(rawToken),
		ExpiresAt: time.Now().Add(sessionLifetime),
	}); err != nil {
		return "", Account{}, err
	}

	return rawToken, record.Account, nil
}

func (service *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return service.repository.DeleteSession(ctx, hashSessionToken(rawToken))
}

func (service *Service) AccountForSession(ctx context.Context, rawToken string) (Account, error) {
	if rawToken == "" {
		return Account{}, ErrUnauthenticated
	}
	result, err := service.repository.AccountBySession(ctx, hashSessionToken(rawToken), time.Now())
	if errors.Is(err, account.ErrNotFound) {
		return Account{}, ErrUnauthenticated
	}
	return result, err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newSessionToken() (string, error) {
	value := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashSessionToken(rawToken string) []byte {
	hash := sha256.Sum256([]byte(rawToken))
	return hash[:]
}
