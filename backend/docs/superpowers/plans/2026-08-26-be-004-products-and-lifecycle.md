# BE-004 Products and Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an authenticated user create, read, and complete only their own products.

**Architecture:** The product repository scopes every query by `account_id`; a service validates fields and guards terminal state transitions. HTTP handlers map JSON DTOs to service input and hide foreign records.

**Tech Stack:** Go, PostgreSQL, `net/http`, Go `testing`.

**Spec:** `backend/docs/tasks/BE-004-products-and-lifecycle.md`

## Global Constraints

- Depends on BE-001 and BE-003.
- Required fields are name, date type, and date; date types are `use_by` and `best_before`.
- Valid terminal statuses are distinct `used` and `discarded`; other accounts cannot observe a product.

---

### Task 1: Define product schema and stateful service

**Files:**
- Create: `backend/migrations/000003_products.sql`, `backend/internal/product/model.go`, `backend/internal/product/repository.go`, `backend/internal/product/service.go`
- Test: `backend/internal/product/service_test.go`

**Interfaces:**
- Produces: `Create(ctx, accountID uuid.UUID, input CreateInput) (Product, error)`, `Get(ctx, accountID, productID uuid.UUID) (Product, error)`, and `Complete(ctx, accountID, productID uuid.UUID, status Status) (Product, error)`.

- [ ] **Step 1: Write failing validation and transition tests**

```go
func TestCreateRequiresNameDateTypeAndDate(t *testing.T) {
    _, err := service.Create(ctx, accountID, CreateInput{})
    if !errors.Is(err, ErrInvalidProduct) { t.Fatal(err) }
}
func TestUsedAndDiscardedAreDistinctTerminalStates(t *testing.T) {
    used, _ := service.Complete(ctx, accountID, productID, StatusUsed)
    if used.Status != StatusUsed || used.CompletedAt == nil { t.Fatal(used) }
}
```

- [ ] **Step 2: Run failing tests**

Run: `cd backend && go test ./internal/product -run 'TestCreateRequires|TestUsedAndDiscarded' -v`

Expected: FAIL because product service is absent.

- [ ] **Step 3: Implement schema and service**

Persist name, `date_type`, `expiry_date`, optional quantity/unit/group/storage, status, and `completed_at`. Reject a second completion and any status other than `used` or `discarded` in `Complete`.

- [ ] **Step 4: Verify service behavior**

Run: `cd backend && go test ./internal/product -v`

Expected: PASS for required fields, allowed date types, optional fields, and terminal actions.

### Task 2: Add account-scoped product API

**Files:**
- Create: `backend/internal/product/http.go`
- Modify: `backend/internal/http/server.go`
- Test: `backend/internal/product/http_test.go`

**Interfaces:**
- Produces: `POST /v1/products`, `GET /v1/products`, `GET /v1/products/{id}`, `POST /v1/products/{id}/use`, and `POST /v1/products/{id}/discard`.

- [ ] **Step 1: Write failing isolation test**

```go
func TestForeignProductIsNotVisible(t *testing.T) {
    rr := performAs(secondAccount, http.MethodGet, "/v1/products/"+firstProductID.String(), nil)
    if rr.Code != http.StatusNotFound { t.Fatal(rr.Code) }
}
```

- [ ] **Step 2: Run the failing handler test**

Run: `cd backend && go test ./internal/product -run TestForeignProductIsNotVisible -v`

Expected: FAIL because handlers are absent.

- [ ] **Step 3: Implement handlers**

Read the account from BE-003 middleware, validate request JSON, pass its ID to every service call, return `400` for invalid input and `404` for a missing or foreign product.

- [ ] **Step 4: Run checks and commit**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

Run: `git add backend/migrations backend/internal/product backend/internal/http/server.go && git commit -m "feat(backend): add products lifecycle"`

## Self-review

- Coverage: create/list/detail, required fields, account isolation, use/discard, and repeat-completion rejection are covered.

