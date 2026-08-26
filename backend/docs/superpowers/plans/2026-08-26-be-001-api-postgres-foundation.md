# BE-001 API and PostgreSQL Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a testable Go/PostgreSQL service whose health endpoint returns HTTP 200 and `{"status":"ok"}`.

**Architecture:** A small `cmd/api` composition root loads `DATABASE_URL`, opens the database, and passes it to an HTTP router. SQL migrations are ordered files; the initial migration reserves only migration bookkeeping.

**Tech Stack:** Go, `net/http`, PostgreSQL, SQL migrations, Docker Compose, Make.

**Spec:** `backend/docs/tasks/BE-001-api-postgres-foundation.md`

## Global Constraints

- Do not add accounts, products, e-mail, or OCR in this task.
- `make run`, `make migrate-up`, `go test ./...`, and `go vet ./...` must succeed.

---

### Task 1: Establish module, configuration, and health router

**Files:**
- Create: `backend/go.mod`, `backend/cmd/api/main.go`, `backend/internal/config/config.go`, `backend/internal/http/server.go`
- Test: `backend/internal/http/server_test.go`

**Interfaces:**
- Produces: `func NewServer(db *sql.DB) http.Handler` and `GET /healthz`.

- [ ] **Step 1: Write the failing health test**

```go
func TestHealthzReturnsOK(t *testing.T) {
    w := httptest.NewRecorder()
    NewServer(nil).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
    if w.Code != http.StatusOK || w.Body.String() != "{\"status\":\"ok\"}\n" { t.Fatal(w.Code, w.Body.String()) }
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd backend && go test ./internal/http -run TestHealthzReturnsOK -v`

Expected: FAIL because `NewServer` is unavailable.

- [ ] **Step 3: Implement the smallest router and configuration**

```go
func NewServer(db *sql.DB) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Content-Type", "application/json"); _, _ = w.Write([]byte("{\"status\":\"ok\"}\n")) })
    return mux
}
```

Load a non-empty `DATABASE_URL` before opening the database in `main`.

- [ ] **Step 4: Run focused verification**

Run: `cd backend && go test ./internal/http -run TestHealthzReturnsOK -v`

Expected: PASS.

### Task 2: Add migration and developer workflow

**Files:**
- Create: `backend/migrations/000001_init.up.sql`, `backend/migrations/000001_init.down.sql`
- Modify: `backend/docker-compose.yml`, `backend/Makefile`
- Test: full Go checks

**Interfaces:**
- Produces: a migration workflow and a PostgreSQL connection used by the API composition root.

- [ ] **Step 1: Add the initial migration**

Create the `golang-migrate` pair `000001_init.up.sql` and `000001_init.down.sql`. The up migration creates only the migration bookkeeping required by the selected command; the down migration reverses only that change. Feature tables belong to later BE tasks.

- [ ] **Step 2: Wire the service command paths**

Ensure `make run` starts `cmd/api`, `make migrate-up` applies `migrations/`, and Compose passes `DATABASE_URL` without exposing a secret in Git.

- [ ] **Step 3: Run repository checks**

Run: `cd backend && go test ./... && go vet ./... && make migrate-up`

Expected: PASS against the local PostgreSQL service.

- [ ] **Step 4: Commit**

Run: `git add backend/go.mod backend/go.sum backend/cmd/api backend/internal backend/migrations/000001_init.up.sql backend/migrations/000001_init.down.sql backend/docker-compose.yml backend/Makefile && git commit -m "feat(backend): add API and postgres foundation"`

## Self-review

- Coverage: health response, configuration, connection setup, migration command, and verification commands are covered.
- Account and product schemas are intentionally deferred.
