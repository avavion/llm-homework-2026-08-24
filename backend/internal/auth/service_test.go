package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"llm-homework/backend/internal/account"
)

func TestRegisterNormalizesAndRejectsCaseVariant(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	password := "correct horse battery staple"

	registered, err := service.Register(context.Background(), "User@Example.COM", password)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if registered.Email != "user@example.com" {
		t.Fatalf("registered email = %q, want %q", registered.Email, "user@example.com")
	}
	if repository.accounts[registered.Email].PasswordHash == password {
		t.Fatal("stored password hash contains the submitted password unchanged")
	}

	if _, err := service.Register(context.Background(), "user@example.com", "another password"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("second Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestPasswordHashUsesUniqueSaltAndVerifies(t *testing.T) {
	const password = "correct horse battery staple"

	first, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword(first) error = %v", err)
	}
	second, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword(second) error = %v", err)
	}

	if first == second {
		t.Fatal("two hashes for the same password are equal; want unique salts")
	}
	if first == password || second == password {
		t.Fatal("encoded password hash equals the submitted password")
	}
	if ok, err := verifyPassword(password, first); err != nil || !ok {
		t.Fatalf("verifyPassword(correct) = %v, %v; want true, nil", ok, err)
	}
	if ok, err := verifyPassword("wrong password", first); err != nil || ok {
		t.Fatalf("verifyPassword(wrong) = %v, %v; want false, nil", ok, err)
	}
}

func TestLoginStoresOnlySHA256TokenHash(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	ctx := context.Background()

	registered, err := service.Register(ctx, "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	rawToken, loggedIn, err := service.Login(ctx, "USER@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rawToken == "" {
		t.Fatal("Login() raw token is empty")
	}
	if loggedIn.ID != registered.ID {
		t.Fatalf("Login() account ID = %s, want %s", loggedIn.ID, registered.ID)
	}

	stored := repository.sessions[0]
	wantHash := sha256.Sum256([]byte(rawToken))
	if !bytes.Equal(stored.TokenHash, wantHash[:]) {
		t.Fatalf("stored token hash = %x, want SHA-256(raw token) %x", stored.TokenHash, wantHash)
	}
	if bytes.Equal(stored.TokenHash, []byte(rawToken)) {
		t.Fatal("repository stored the raw session token")
	}
	if !stored.ExpiresAt.After(time.Now()) {
		t.Fatalf("session expiry = %s, want a future instant", stored.ExpiresAt)
	}
}

func TestLoginReturnsSameGenericErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	ctx := context.Background()

	if _, err := service.Register(ctx, "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, _, wrongPasswordErr := service.Login(ctx, "user@example.com", "wrong password")
	_, _, unknownEmailErr := service.Login(ctx, "missing@example.com", "wrong password")
	if !errors.Is(wrongPasswordErr, ErrInvalidCredentials) {
		t.Fatalf("wrong-password error = %v, want ErrInvalidCredentials", wrongPasswordErr)
	}
	if !errors.Is(unknownEmailErr, ErrInvalidCredentials) {
		t.Fatalf("unknown-email error = %v, want ErrInvalidCredentials", unknownEmailErr)
	}
	if wrongPasswordErr.Error() != unknownEmailErr.Error() {
		t.Fatalf("authentication errors differ: %q vs %q", wrongPasswordErr, unknownEmailErr)
	}
}

func TestLogoutDeletesSessionByTokenHash(t *testing.T) {
	repository := newMemoryRepository()
	service := NewService(repository)
	ctx := context.Background()

	if _, err := service.Register(ctx, "user@example.com", "correct horse battery staple"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	rawToken, _, err := service.Login(ctx, "user@example.com", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if err := service.Logout(ctx, rawToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if len(repository.sessions) != 0 {
		t.Fatalf("session count after Logout() = %d, want 0", len(repository.sessions))
	}
}

type memoryRepository struct {
	accounts map[string]account.Record
	sessions []account.Session
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		accounts: make(map[string]account.Record),
	}
}

func (repository *memoryRepository) CreateAccount(_ context.Context, email, passwordHash string) (account.Account, error) {
	if _, exists := repository.accounts[email]; exists {
		return account.Account{}, account.ErrEmailTaken
	}
	record := account.Record{
		Account:      account.Account{Email: email},
		PasswordHash: passwordHash,
	}
	repository.accounts[email] = record
	return record.Account, nil
}

func (repository *memoryRepository) AccountByEmail(_ context.Context, email string) (account.Record, error) {
	record, exists := repository.accounts[email]
	if !exists {
		return account.Record{}, account.ErrNotFound
	}
	return record, nil
}

func (repository *memoryRepository) CreateSession(_ context.Context, session account.Session) error {
	session.TokenHash = append([]byte(nil), session.TokenHash...)
	repository.sessions = append(repository.sessions, session)
	return nil
}

func (repository *memoryRepository) DeleteSession(_ context.Context, tokenHash []byte) error {
	for index, session := range repository.sessions {
		if bytes.Equal(session.TokenHash, tokenHash) {
			repository.sessions = append(repository.sessions[:index], repository.sessions[index+1:]...)
			break
		}
	}
	return nil
}

func (repository *memoryRepository) AccountBySession(_ context.Context, tokenHash []byte, now time.Time) (account.Account, error) {
	for _, session := range repository.sessions {
		if bytes.Equal(session.TokenHash, tokenHash) && session.ExpiresAt.After(now) {
			for _, record := range repository.accounts {
				if record.ID == session.AccountID {
					return record.Account, nil
				}
			}
		}
	}
	return account.Account{}, account.ErrNotFound
}
