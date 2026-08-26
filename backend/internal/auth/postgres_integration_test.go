//go:build integration

package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"llm-homework/backend/internal/account"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresAuthenticationLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	repository := account.NewRepository(db)
	service := NewService(repository)
	localPart := "be003-" + uuid.NewString()
	emails := []string{localPart + "@Example.COM", strings.ToUpper(localPart) + "@example.com"}
	const password = "correct horse battery staple"

	type registrationResult struct {
		account Account
		err     error
	}
	start := make(chan struct{})
	results := make(chan registrationResult, len(emails))
	var workers sync.WaitGroup
	for _, email := range emails {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			registered, err := service.Register(ctx, email, password)
			results <- registrationResult{account: registered, err: err}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	var registered Account
	successes := 0
	duplicates := 0
	for result := range results {
		switch {
		case result.err == nil:
			successes++
			registered = result.account
		case errors.Is(result.err, ErrEmailTaken):
			duplicates++
		default:
			t.Fatalf("Register() concurrent result error = %v", result.err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("concurrent registration results: successes=%d duplicates=%d, want 1 and 1", successes, duplicates)
	}
	if registered.Email != strings.ToLower(localPart)+"@example.com" {
		t.Fatalf("registered email = %q, want normalized e-mail", registered.Email)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id = $1", registered.ID)
	})

	var storedPassword string
	if err := db.QueryRowContext(ctx, "SELECT password_hash FROM accounts WHERE id = $1", registered.ID).Scan(&storedPassword); err != nil {
		t.Fatalf("query stored password: %v", err)
	}
	if storedPassword == password || !strings.HasPrefix(storedPassword, "$argon2id$") {
		t.Fatalf("stored password is not a non-plaintext Argon2id encoding")
	}

	rawToken, loggedIn, err := service.Login(ctx, registered.Email, password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if loggedIn.ID != registered.ID {
		t.Fatalf("Login() account ID = %s, want %s", loggedIn.ID, registered.ID)
	}
	var storedTokenHash []byte
	if err := db.QueryRowContext(ctx, "SELECT token_hash FROM auth_sessions WHERE account_id = $1", registered.ID).Scan(&storedTokenHash); err != nil {
		t.Fatalf("query stored token hash: %v", err)
	}
	wantTokenHash := sha256.Sum256([]byte(rawToken))
	if !bytes.Equal(storedTokenHash, wantTokenHash[:]) || bytes.Equal(storedTokenHash, []byte(rawToken)) {
		t.Fatal("database does not contain only SHA-256(raw session token)")
	}

	if _, err := db.ExecContext(ctx, "UPDATE auth_sessions SET expires_at = $1 WHERE token_hash = $2", time.Now().Add(-time.Minute), storedTokenHash); err != nil {
		t.Fatalf("expire session: %v", err)
	}
	if _, err := service.AccountForSession(ctx, rawToken); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("AccountForSession(expired) error = %v, want ErrUnauthenticated", err)
	}

	if err := service.Logout(ctx, rawToken); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	var remainingSessions int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM auth_sessions WHERE account_id = $1", registered.ID).Scan(&remainingSessions); err != nil {
		t.Fatalf("count sessions after Logout(): %v", err)
	}
	if remainingSessions != 0 {
		t.Fatalf("sessions after Logout() = %d, want 0", remainingSessions)
	}
}
