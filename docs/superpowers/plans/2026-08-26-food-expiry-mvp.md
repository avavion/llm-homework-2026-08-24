# Food Expiry MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Russian- and English-language web MVP that lets a home user record food products, receive e-mail reminders, and safely distinguish expiry-date types.

**Architecture:** The Go API owns identity, products, date-rule evaluation, draft recognition and notification scheduling; PostgreSQL persists normalized user and product data. The Next.js application renders the public indexable site and authenticated inventory UI, while photo OCR/LLM output remains a server-side draft until the user explicitly approves it. Date rules live in versioned configuration backed by researched regulatory evidence, not in application conditionals.

**Tech Stack:** Go, PostgreSQL, Go migrations, Go standard `testing` package, Next.js, TypeScript, React Testing Library, Playwright, e-mail provider adapter, OCR adapter, LLM API adapter.

**Spec:** `shared/docs/product-description.md`

## Global Constraints

- MVP is a web application in Russian and English; do not build a native mobile app or push notifications.
- Registration and login use e-mail plus password without e-mail verification; normalize e-mail to lower case, enforce uniqueness in PostgreSQL, and store only Argon2id password hashes.
- A product cannot be created from OCR/LLM output before explicit user `approve`; `reject` must leave no product record.
- Required product fields are name, date type and date; quantity, unit, group and storage location are optional.
- `use by` products after the confirmed regulatory expiry instant are excluded from recipes; `best before` products remain available with a warning.
- Country selection covers CIS and EU countries, but date-rule behavior is enabled only after primary-source research has documented the regulator group.
- E-mail is the only MVP notification channel; users can set an alert threshold down to one hour.
- Delivery, coupons, advertising, e-mail verification, shared households, external recipe catalogs and LLM recipe generation are out of scope.

---

### Task 1: Research and publish the date-rule registry

**Files:**
- Create: `shared/docs/regulatory-date-rules.md`
- Create: `shared/docs/product-group-alert-policy.md`
- Test: `shared/docs/regulatory-date-rules.md` review checklist

**Interfaces:**
- Produces: a country-to-regulator-group table, evidence URLs, source access dates, and a machine-readable rule contract: `group`, `date_type`, `expiry_interpretation`, `status_after_expiry`.
- Produces: initial product groups and default e-mail alert windows that the backend seeds in Task 3.

- [ ] **Step 1: Collect primary legal sources**

For every selected regulator group, record the regulator, countries, direct primary-source URL, access date, the legal label for the date, and whether the date concerns safety or quality. Use the EU sources already approved by the spec for `use by` / `best before`; do not infer an EAEU midnight rule without an authoritative source.

- [ ] **Step 2: Define the registry contract**

Write a table whose each enabled row contains exactly: `regulator_group`, ISO country codes, `date_type`, `expiry_timezone_source`, `expiry_instant_rule`, `post_expiry_status`, `recipe_eligibility`, evidence URL and evidence access date. Mark a row as `research_required` if any field is unconfirmed; such a row must not be enabled for automation.

- [ ] **Step 3: Define initial alert policy**

Write a separate table for `refrigerated_perishable`, `fresh_produce`, `frozen`, `shelf_stable`, and `other`. Each row states a default e-mail alert window, the minimum allowed user override of one hour, and the source/rationale. The document must make clear that defaults are convenience settings, not safety advice.

- [ ] **Step 4: Review the research gate**

Run: `rg -n "regulator_group|research_required|evidence URL|one hour|use by|best before" shared/docs/regulatory-date-rules.md shared/docs/product-group-alert-policy.md`

Expected: every enabled automation rule is evidence-backed; unverified rules are explicitly disabled.

### Task 2: Scaffold Go API, PostgreSQL and migration workflow

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/api/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/http/server.go`
- Create: `backend/migrations/000001_init.sql`
- Create: `backend/docker-compose.yml`
- Create: `backend/Makefile`
- Test: `backend/internal/http/server_test.go`

**Interfaces:**
- Produces: `GET /healthz` returning JSON `{"status":"ok"}`.
- Produces: PostgreSQL connectivity via `DATABASE_URL` and migration commands `make migrate-up`, `make test`, `make run`.

- [ ] **Step 1: Write the failing health endpoint test**

```go
func TestHealthzReturnsOK(t *testing.T) {
    request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    response := httptest.NewRecorder()
    NewServer(nil).ServeHTTP(response, request)
    if response.Code != http.StatusOK { t.Fatalf("got %d", response.Code) }
    if response.Body.String() != "{\"status\":\"ok\"}\n" { t.Fatal(response.Body.String()) }
}
```

- [ ] **Step 2: Run the failing test**

Run: `cd backend && go test ./internal/http -run TestHealthzReturnsOK -v`

Expected: FAIL because `NewServer` is not implemented.

- [ ] **Step 3: Implement the minimal service skeleton**

Create `NewServer(db *sql.DB) http.Handler`, register `GET /healthz`, load `DATABASE_URL` from the environment, and add a Docker Compose PostgreSQL service with a named volume. The initial SQL migration creates only `schema_migrations`; feature tables arrive in later migrations.

- [ ] **Step 4: Run focused and service checks**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

### Task 3: Implement account, country and product persistence

**Files:**
- Create: `backend/migrations/000002_accounts_products.sql`
- Create: `backend/internal/account/service.go`
- Create: `backend/internal/account/repository.go`
- Create: `backend/internal/auth/service.go`
- Create: `backend/internal/auth/password.go`
- Create: `backend/internal/auth/http.go`
- Create: `backend/internal/product/model.go`
- Create: `backend/internal/product/repository.go`
- Create: `backend/internal/regulation/repository.go`
- Test: `backend/internal/account/service_test.go`
- Test: `backend/internal/auth/service_test.go`
- Test: `backend/internal/product/repository_test.go`

**Interfaces:**
- Produces: `accounts(id, email_normalized, password_hash, country_code, regulator_group, locale, created_at)` with a unique index on `email_normalized`.
- Produces: `auth_sessions(id, account_id, token_hash, expires_at, created_at)` and `POST /v1/auth/register`, `POST /v1/auth/login`, `POST /v1/auth/logout`.
- Produces: `products(id, account_id, name, date_type, expiry_date, quantity, unit, product_group, storage_location, status, alert_threshold_minutes, created_at, completed_at)`.
- Produces: `Register(ctx, email, password, countryCode, locale)`, `Login(ctx, email, password)`, `Logout(ctx, token)`, and `CreateProduct(ctx, input)`.

- [ ] **Step 1: Write failing account normalization tests**

```go
func TestRegisterNormalizesEmailAndRejectsCaseVariant(t *testing.T) {
    first, _ := service.Register(ctx, "User@Example.COM", "correct horse battery staple", "RU", "ru")
    if first.EmailNormalized != "user@example.com" { t.Fatal(first.EmailNormalized) }
    _, err := service.Register(ctx, "user@example.com", "another password", "RU", "ru")
    if !errors.Is(err, ErrEmailTaken) { t.Fatal("expected ErrEmailTaken") }
}
```

- [ ] **Step 2: Run the failing account test**

Run: `cd backend && go test ./internal/auth -run TestRegisterNormalizesEmailAndRejectsCaseVariant -v`

Expected: FAIL because the authentication service and migration do not exist.

- [ ] **Step 3: Add migrations and repositories**

Use a lower-case database check plus a unique index to guarantee normalized uniqueness. Hash passwords with `golang.org/x/crypto/argon2.IDKey` (Argon2id), a cryptographically random per-password salt, memory `19456` KiB, time `2`, parallelism `1`, and a self-describing encoded hash. Store only the encoded hash. On successful login, create a cryptographically random opaque session token, store only its SHA-256 hash with an expiry, and send the raw token in an `HttpOnly`, `Secure`, `SameSite=Lax` cookie. Constrain `date_type` to `use_by` or `best_before`, require name and expiry date, and constrain `status` to `active`, `attention`, `expired`, `used`, or `discarded`. Store ISO country code and regulator group independently.

- [ ] **Step 4: Write and run product validation tests**

Test registration, login, logout, duplicate e-mail rejection, wrong-password rejection, and that database storage contains no plain-text password. Test that missing name, date type or expiry date is rejected; optional quantity, unit, group and storage persist; a product belongs only to its account.

Run: `cd backend && go test ./internal/account ./internal/product -v`

Expected: PASS.

### Task 4: Implement date status, alert scheduling and e-mail delivery

**Files:**
- Create: `backend/internal/regulation/evaluator.go`
- Create: `backend/internal/notification/service.go`
- Create: `backend/internal/notification/email.go`
- Create: `backend/internal/notification/worker.go`
- Create: `backend/internal/notification/templates/expiry_reminder.tmpl`
- Test: `backend/internal/regulation/evaluator_test.go`
- Test: `backend/internal/notification/service_test.go`

**Interfaces:**
- Produces: `Evaluate(product, rule, now) Status` and `NextNotificationAt(product, rule) time.Time`.
- Produces: `EmailSender.SendExpiryReminder(ctx, recipient, product)`; a development sender writes messages to logs.

- [ ] **Step 1: Write failing safety-rule tests**

```go
func TestUseByAfterExpiryIsExcludedFromRecipes(t *testing.T) {
    status := Evaluate(Product{DateType: UseBy, ExpiryDate: date}, Rule{ExpiryInstant: instant}, instant.Add(time.Second))
    if status != Expired { t.Fatal(status) }
    if EligibleForRecipes(status) { t.Fatal("expired use-by product is eligible") }
}

func TestBestBeforeAfterDateNeedsAttentionButStaysEligible(t *testing.T) {
    status := Evaluate(Product{DateType: BestBefore, ExpiryDate: date}, Rule{ExpiryInstant: instant}, instant.Add(time.Second))
    if status != Attention || !EligibleForRecipes(status) { t.Fatal(status) }
}
```

- [ ] **Step 2: Run the failing evaluator tests**

Run: `cd backend && go test ./internal/regulation -run 'TestUseBy|TestBestBefore' -v`

Expected: FAIL because evaluator functions do not exist.

- [ ] **Step 3: Implement rule-driven evaluation**

Load only enabled rules from the Task 1 registry seed. Reject automatic scheduling for a country whose rule is `research_required`; return a user-visible configuration state instead. Validate that alert thresholds are at least 60 minutes. Make the worker idempotent with a `notification_deliveries(product_id, scheduled_for, channel)` unique key.

- [ ] **Step 4: Run alert and e-mail tests**

Test threshold calculation, duplicate-delivery prevention, development e-mail content, and no schedule for an unverified rule.

Run: `cd backend && go test ./internal/regulation ./internal/notification -v`

Expected: PASS.

### Task 5: Add product API, lifecycle actions and recipe eligibility

**Files:**
- Create: `backend/internal/product/service.go`
- Create: `backend/internal/product/http.go`
- Create: `backend/internal/recipe/service.go`
- Create: `backend/internal/recipe/http.go`
- Test: `backend/internal/product/http_test.go`
- Test: `backend/internal/recipe/service_test.go`

**Interfaces:**
- Produces: `POST /v1/products`, `GET /v1/products`, `POST /v1/products/{id}/use`, and `POST /v1/products/{id}/discard`.
- Produces: `GET /v1/recipes?product_ids=...`, returning deterministic first-party recommendations only for eligible products.

- [ ] **Step 1: Write failing API contract tests**

Test that a create request with `name`, `date_type`, and `expiry_date` succeeds; a request without each required field returns `400`; another account receives `404` for the first account's product; `use` and `discard` set separate terminal statuses.

- [ ] **Step 2: Run the failing API tests**

Run: `cd backend && go test ./internal/product -run 'TestCreateProduct|TestCompleteProduct' -v`

Expected: FAIL because handlers do not exist.

- [ ] **Step 3: Implement minimal endpoints**

Authenticate requests by the opaque session cookie from Task 3, scope every query by account ID, and return JSON product DTOs with status, date type, expiry date, optional fields and notification threshold. Do not expose another account's products.

- [ ] **Step 4: Implement recipe filter and tests**

Seed a small first-party recommendation map keyed by product group. Ensure a recipe never includes an `expired` product and can include a `best_before` product marked `attention`.

Run: `cd backend && go test ./internal/product ./internal/recipe -v`

Expected: PASS.

### Task 6: Add draft photo recognition with explicit approval

**Files:**
- Create: `backend/internal/recognition/service.go`
- Create: `backend/internal/recognition/ocr.go`
- Create: `backend/internal/recognition/llm.go`
- Create: `backend/internal/recognition/http.go`
- Create: `backend/migrations/000003_product_drafts.sql`
- Test: `backend/internal/recognition/service_test.go`
- Test: `backend/internal/recognition/http_test.go`

**Interfaces:**
- Produces: `POST /v1/product-drafts/recognize`, `POST /v1/product-drafts/{id}/approve`, and `POST /v1/product-drafts/{id}/reject`.
- Produces: `ProductDraft` with extracted fields, source photo reference, recognition status and account ID.

- [ ] **Step 1: Write the failing approval test**

```go
func TestDraftDoesNotCreateProductUntilApproved(t *testing.T) {
    draft := service.CreateRecognizedDraft(ctx, accountID, recognition)
    if products.Count(ctx, accountID) != 0 { t.Fatal("product created before approval") }
    service.Approve(ctx, accountID, draft.ID, approvedFields)
    if products.Count(ctx, accountID) != 1 { t.Fatal("product missing after approval") }
}
```

- [ ] **Step 2: Run the failing test**

Run: `cd backend && go test ./internal/recognition -run TestDraftDoesNotCreateProductUntilApproved -v`

Expected: FAIL because draft service does not exist.

- [ ] **Step 3: Implement adapters and fallback**

Define `OCRClient.ExtractText(ctx, image)` and `LLMClient.ParseProduct(ctx, text, locale)`. Persist only a draft on success; on OCR, LLM or network failure return a structured failure response that lets the frontend open an empty manual form. `Reject` must mark the draft rejected and create no product.

- [ ] **Step 4: Run draft tests**

Test approval, rejection, account isolation, malformed OCR data and upstream failure fallback.

Run: `cd backend && go test ./internal/recognition -v`

Expected: PASS.

### Task 7: Scaffold Next.js and build core inventory flows

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/next.config.ts`
- Create: `frontend/app/[locale]/page.tsx`
- Create: `frontend/app/[locale]/inventory/page.tsx`
- Create: `frontend/components/ProductForm.tsx`
- Create: `frontend/components/ProductList.tsx`
- Create: `frontend/components/ProductStatusBadge.tsx`
- Create: `frontend/lib/api.ts`
- Create: `frontend/messages/ru.json`
- Create: `frontend/messages/en.json`
- Test: `frontend/components/ProductForm.test.tsx`
- Test: `frontend/components/ProductList.test.tsx`

**Interfaces:**
- Consumes: product API DTO `{id,name,date_type,expiry_date,status,alert_threshold_minutes}`.
- Produces: localized public home page and authenticated inventory with manual product creation and lifecycle actions.

- [ ] **Step 1: Write failing form tests**

Test that the form disables submit until `name`, `date_type`, and `expiry_date` are set, submits only those fields as required values, and permits empty optional fields.

- [ ] **Step 2: Run the failing component test**

Run: `cd frontend && npm test -- ProductForm.test.tsx`

Expected: FAIL because the Next.js project and component do not exist.

- [ ] **Step 3: Build manual inventory UI**

Create locale-aware routes for `ru` and `en`. Render status badges with separate accessible text for attention and expired products; provide buttons for used and discarded actions. Use server-rendered metadata and public home content so the marketing page remains indexable.

- [ ] **Step 4: Run component tests and type checks**

Run: `cd frontend && npm test -- ProductForm.test.tsx ProductList.test.tsx && npm run typecheck`

Expected: PASS.

### Task 8: Build photo draft review, settings and recipe UI

**Files:**
- Create: `frontend/components/PhotoRecognitionFlow.tsx`
- Create: `frontend/components/NotificationSettings.tsx`
- Create: `frontend/components/RecipeRecommendations.tsx`
- Create: `frontend/app/[locale]/settings/page.tsx`
- Test: `frontend/components/PhotoRecognitionFlow.test.tsx`
- Test: `frontend/components/NotificationSettings.test.tsx`
- Test: `frontend/e2e/mvp.spec.ts`

**Interfaces:**
- Consumes: draft API responses and product-status DTOs.
- Produces: upload/camera entry, editable recognized form, `approve`/`reject`, country selection, hourly-minimum alert settings and recipe recommendations.

- [ ] **Step 1: Write failing draft-review tests**

Test that a recognized draft populates editable fields, `approve` calls the approval endpoint, `reject` calls the rejection endpoint, and recognition failure shows the manual form without blocking the user.

- [ ] **Step 2: Run the failing draft component test**

Run: `cd frontend && npm test -- PhotoRecognitionFlow.test.tsx`

Expected: FAIL because the component does not exist.

- [ ] **Step 3: Implement photo and settings flows**

Use the browser camera only as an optional file-input capture source. Prevent an alert threshold below 60 minutes in both client validation and displayed help text. Render the country selector from backend-supported countries and show a non-automation notice for a `research_required` rule.

- [ ] **Step 4: Implement recipe presentation and end-to-end checks**

Render only backend-returned recommendations. Add Playwright flows: manually add a product, complete it as used, submit a photo draft and reject it, and confirm an expired `use by` item has no recipe action.

Run: `cd frontend && npm test && npm run typecheck && npx playwright test e2e/mvp.spec.ts`

Expected: PASS.

### Task 9: Validate the full MVP and operational handoff

**Files:**
- Create: `shared/docs/mvp-acceptance-checklist.md`
- Modify: `shared/docs/product-description.md`
- Test: `backend/...`, `frontend/e2e/mvp.spec.ts`

**Interfaces:**
- Produces: a repeatable acceptance checklist covering every MVP criterion and links to the verified regulator registry.

- [ ] **Step 1: Write acceptance scenarios**

Write scenarios for manual add, photo-draft approve/reject, e-mail threshold of one hour, `use by` recipe exclusion, `best before` warning, used/discarded lifecycle, e-mail normalization uniqueness, Russian/English UI, and no unsupported country automation.

- [ ] **Step 2: Run all backend checks**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS.

- [ ] **Step 3: Run all frontend checks**

Run: `cd frontend && npm test && npm run typecheck && npx playwright test`

Expected: PASS.

- [ ] **Step 4: Review scope and operational constraints**

Run: `rg -n "push|mobile|coupon|delivery|verification|external recipe|LLM recipe" shared/docs/mvp-acceptance-checklist.md shared/docs/product-description.md`

Expected: the checklist explicitly confirms these excluded features are not required for MVP acceptance.
