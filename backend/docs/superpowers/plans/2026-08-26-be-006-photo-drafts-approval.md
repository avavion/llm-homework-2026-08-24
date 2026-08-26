# BE-006 Photo Drafts and Approval Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create editable OCR/LLM product drafts and persist a product only after the owning user approves one.

**Architecture:** OCR and LLM are replaceable interfaces whose output is stored in an account-scoped `product_drafts` record. Approval validates user-supplied fields through the BE-004 product service; rejection never creates a product and adapter failure leaves manual entry available.

**Tech Stack:** Go, PostgreSQL, HTTP multipart upload, Go `testing`.

**Spec:** `backend/docs/tasks/BE-006-photo-drafts-approval.md`

## Global Constraints

- Depends on BE-003 and BE-004.
- Recognition must not write to `products`; only `approve` may do so.
- Adapter errors must not disclose secrets and must not block the manual form.

---

### Task 1: Model draft recognition and persistence

**Files:**
- Create: `backend/migrations/000005_product_drafts.sql`, `backend/internal/recognition/ocr.go`, `backend/internal/recognition/llm.go`, `backend/internal/recognition/repository.go`, `backend/internal/recognition/service.go`
- Test: `backend/internal/recognition/service_test.go`

**Interfaces:**
- Produces: `OCRClient.ExtractText(ctx context.Context, image []byte) (string, error)`, `LLMClient.ParseProduct(ctx context.Context, text, locale string) (DraftFields, error)`, and `Recognize(ctx, accountID uuid.UUID, image []byte) (ProductDraft, error)`.

- [ ] **Step 1: Write the failing no-product-before-approval test**

```go
func TestRecognizeCreatesDraftNotProduct(t *testing.T) {
    _, _ = service.Recognize(ctx, accountID, image)
    if products.Count(ctx, accountID) != 0 { t.Fatal("product created before approval") }
}
```

- [ ] **Step 2: Run the failing test**

Run: `cd backend && go test ./internal/recognition -run TestRecognizeCreatesDraftNotProduct -v`

Expected: FAIL because draft persistence is absent.

- [ ] **Step 3: Implement adapter boundary and draft storage**

Store account ID, recognized fields, source reference, recognition status, and timestamps in `product_drafts`. Convert OCR/LLM failures to a stable response code that tells the client to use manual entry; never log request credentials or raw provider secrets.

- [ ] **Step 4: Run focused tests**

Run: `cd backend && go test ./internal/recognition -v`

Expected: PASS for successful draft and adapter failure fallback.

### Task 2: Implement approve/reject HTTP workflow

**Files:**
- Create: `backend/internal/recognition/http.go`
- Modify: `backend/internal/http/server.go`
- Test: `backend/internal/recognition/http_test.go`

**Interfaces:**
- Produces: `POST /v1/product-drafts/recognize`, `POST /v1/product-drafts/{id}/approve`, and `POST /v1/product-drafts/{id}/reject`.

- [ ] **Step 1: Write failing approval, rejection, and foreign-owner tests**

```go
func TestApproveCreatesOneProductAndRejectCreatesNone(t *testing.T) {
    approved := approveAs(owner, draft.ID, editedFields)
    if approved.Code != http.StatusCreated || products.Count(ctx, owner.ID) != 1 { t.Fatal(approved.Code) }
    rejected := rejectAs(owner, anotherDraft.ID)
    if rejected.Code != http.StatusNoContent || products.Count(ctx, owner.ID) != 1 { t.Fatal(rejected.Code) }
}
```

Test a foreign account receives `404` and a second approval cannot create another product.

- [ ] **Step 2: Run the failing workflow tests**

Run: `cd backend && go test ./internal/recognition -run 'TestApprove|TestForeign' -v`

Expected: FAIL because handlers are absent.

- [ ] **Step 3: Implement transactions and handlers**

Scope draft lookup by account ID. On approval, validate edited fields with BE-004 and atomically create one product plus mark the draft approved; on rejection, mark it rejected only.

- [ ] **Step 4: Run checks and commit**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

Run: `git add backend/migrations backend/internal/recognition backend/internal/http/server.go && git commit -m "feat(backend): add photo drafts approval"`

## Self-review

- Coverage: draft-only recognition, editable approval, rejection, double approval, account isolation, and failure fallback are covered.

