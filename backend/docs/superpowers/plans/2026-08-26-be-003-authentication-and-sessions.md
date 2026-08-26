# BE-003 Authentication and Sessions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide e-mail/password registration, login, logout, and account-isolated sessions without storing a readable password or session token.

**Architecture:** PostgreSQL stores lower-case e-mail and Argon2id encoded password hashes; an opaque random session token is sent only in a secure cookie while its SHA-256 hash is persisted. Every authenticated resource lookup takes the current account ID.

**Tech Stack:** Go, PostgreSQL, `golang.org/x/crypto/argon2`, `net/http`.

**Spec:** `backend/docs/tasks/BE-003-authentication-and-sessions.md`

## Global Constraints

- Depends on BE-001.
- Normalize e-mail before persistence and enforce its uniqueness.
- Cookie attributes are `HttpOnly`, `Secure`, and `SameSite=Lax`; a foreign resource returns 404 or 401 without data disclosure.

---

### Task 1: Persist accounts and sessions securely

**Files:**
- Create: `backend/migrations/000002_accounts_auth.sql`, `backend/internal/account/repository.go`, `backend/internal/auth/password.go`, `backend/internal/auth/service.go`
- Test: `backend/internal/auth/service_test.go`

**Interfaces:**
- Produces: `Register(ctx, email, password string) (Account, error)`, `Login(ctx, email, password string) (rawToken string, account Account, err error)`, and `Logout(ctx, rawToken string) error`.

- [ ] **Step 1: Write failing normalization and secrecy tests**

```go
func TestRegisterNormalizesAndRejectsCaseVariant(t *testing.T) {
    account, _ := service.Register(ctx, "User@Example.COM", "correct horse battery staple")
    if account.Email != "user@example.com" { t.Fatal(account.Email) }
    if _, err := service.Register(ctx, "user@example.com", "another password"); !errors.Is(err, ErrEmailTaken) { t.Fatal(err) }
}
```

Also assert the stored password column does not equal the submitted password.

- [ ] **Step 2: Run the failing test**

Run: `cd backend && go test ./internal/auth -run TestRegisterNormalizesAndRejectsCaseVariant -v`

Expected: FAIL because the migration and service do not exist.

- [ ] **Step 3: Implement schema and services**

Create `accounts` with a unique normalized e-mail column and `auth_sessions(account_id, token_hash, expires_at)`. Use a unique random salt with Argon2id for each password, store a self-describing encoded hash, generate a random opaque session token, and store only `sha256(rawToken)`.

- [ ] **Step 4: Run service tests**

Run: `cd backend && go test ./internal/auth -v`

Expected: PASS for duplicate e-mail, wrong password, hash verification, and token hash storage.

### Task 2: Expose session endpoints and enforce account context

**Files:**
- Create: `backend/internal/auth/http.go`, `backend/internal/auth/middleware.go`
- Modify: `backend/internal/http/server.go`
- Test: `backend/internal/auth/http_test.go`

**Interfaces:**
- Produces: `POST /v1/auth/register`, `POST /v1/auth/login`, `POST /v1/auth/logout`, `GET /v1/auth/session`, and `func AccountFromContext(context.Context) (Account, bool)`.

- [ ] **Step 1: Write failing cookie and unauthorized tests**

```go
func TestLoginSetsSafeSessionCookie(t *testing.T) {
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, loginRequest("user@example.com", "correct horse battery staple"))
    cookie := rr.Result().Cookies()[0]
    if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode { t.Fatal(cookie) }
}
```

Test that no or invalid cookie causes the protected endpoint to return `401`.

- [ ] **Step 2: Run the failing handler tests**

Run: `cd backend && go test ./internal/auth -run 'TestLoginSetsSafeSessionCookie|TestProtectedEndpoint' -v`

Expected: FAIL because handlers are unavailable.

- [ ] **Step 3: Implement handlers and middleware**

Decode JSON requests, use the service methods, set the raw token only in the secure cookie, clear it on logout, and attach the resolved account to request context. Return generic authentication errors without revealing whether an e-mail exists.

- [ ] **Step 4: Run backend checks and commit**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

Run: `git add backend/migrations backend/internal && git commit -m "feat(backend): add authentication sessions"`

## Self-review

- Coverage: normalization, uniqueness, Argon2id, opaque sessions, secure cookie, logout, and unauthenticated access are covered.
- Product-specific authorization remains owned by BE-004.

