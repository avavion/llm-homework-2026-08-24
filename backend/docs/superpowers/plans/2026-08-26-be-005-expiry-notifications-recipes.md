# BE-005 Expiry Notifications and Recipes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Calculate rule-backed product status, deliver idempotent e-mail reminders, and exclude expired `use_by` products from recipes.

**Architecture:** A registry-backed evaluator returns status and recipe eligibility from a confirmed rule; no status is inferred for `research_required`. A scheduler persists delivery idempotency and delegates e-mail to an adapter; recipes filter only evaluator-approved products.

**Tech Stack:** Go, PostgreSQL, `time`, e-mail adapter, Go `testing`.

**Spec:** `backend/docs/tasks/BE-005-expiry-notifications-recipes.md`

## Global Constraints

- Depends on BE-002 and BE-004.
- Alert thresholds below 60 minutes are rejected.
- `use_by` after expiry is `expired` and recipe-ineligible; `best_before` after expiry is `attention` and remains eligible.
- A `research_required` rule receives neither automatic status nor schedule.

---

### Task 1: Implement registry-driven status and recipe eligibility

**Files:**
- Create: `backend/internal/regulation/evaluator.go`, `backend/internal/regulation/repository.go`, `backend/internal/recipe/service.go`
- Test: `backend/internal/regulation/evaluator_test.go`, `backend/internal/recipe/service_test.go`

**Interfaces:**
- Produces: `Evaluate(product Product, rule Rule, now time.Time) (Status, error)` and `EligibleForRecipes(status Status) bool`.

- [ ] **Step 1: Write failing date-type tests**

```go
func TestUseByAfterExpiryIsExcluded(t *testing.T) {
    status, _ := Evaluate(useByProduct, confirmedRule, expiry.Add(time.Second))
    if status != StatusExpired || EligibleForRecipes(status) { t.Fatal(status) }
}
func TestBestBeforeAfterExpiryNeedsAttention(t *testing.T) {
    status, _ := Evaluate(bestBeforeProduct, confirmedRule, expiry.Add(time.Second))
    if status != StatusAttention || !EligibleForRecipes(status) { t.Fatal(status) }
}
```

- [ ] **Step 2: Run the failing evaluator tests**

Run: `cd backend && go test ./internal/regulation -run 'TestUseBy|TestBestBefore' -v`

Expected: FAIL because evaluator functions are absent.

- [ ] **Step 3: Implement confirmed-rule evaluation**

Load only enabled registry rows. Return a distinct non-automation error for `research_required`; calculate the expiry instant solely from the stored rule and never from a hard-coded country condition.

- [ ] **Step 4: Run evaluator and recipe tests**

Run: `cd backend && go test ./internal/regulation ./internal/recipe -v`

Expected: PASS.

### Task 2: Schedule idempotent e-mail notification

**Files:**
- Create: `backend/migrations/000004_notification_deliveries.sql`, `backend/internal/notification/service.go`, `backend/internal/notification/email.go`, `backend/internal/notification/worker.go`
- Test: `backend/internal/notification/service_test.go`

**Interfaces:**
- Produces: `NextNotificationAt(product Product, threshold time.Duration) (time.Time, error)` and `EmailSender.SendExpiryReminder(ctx context.Context, recipient string, product Product) error`.

- [ ] **Step 1: Write failing threshold and idempotency tests**

```go
func TestSchedulerDoesNotSendSameReminderTwice(t *testing.T) {
    _ = service.DeliverDue(ctx, now)
    _ = service.DeliverDue(ctx, now)
    if sender.Calls != 1 { t.Fatal(sender.Calls) }
}
```

Also assert 59 minutes returns `ErrThresholdTooSmall` and an unverified rule produces no delivery.

- [ ] **Step 2: Run the failing tests**

Run: `cd backend && go test ./internal/notification -run 'TestScheduler|TestThreshold' -v`

Expected: FAIL because notification service is absent.

- [ ] **Step 3: Implement delivery log and development sender**

Add a unique key on `(product_id, scheduled_for, channel)`; insert it before calling the sender, and use a development sender that records safe message content without credentials.

- [ ] **Step 4: Run full checks and commit**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

Run: `git add backend/migrations backend/internal/regulation backend/internal/notification backend/internal/recipe && git commit -m "feat(backend): add expiry notifications"`

## Self-review

- Coverage: rule gate, both date types, 60-minute minimum, idempotency, e-mail failure boundary, and recipe eligibility are covered.

