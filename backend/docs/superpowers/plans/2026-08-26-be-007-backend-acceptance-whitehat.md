# BE-007 Backend Acceptance and White-Hat Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce reproducible MVP acceptance evidence and a local white-hat report with no open Critical or High findings.

**Architecture:** Tests execute only against the local environment and map each acceptance scenario to BE-003 through BE-006. The QA report contains evidence and severity; one remediation package may be retested before a new blocker is escalated.

**Tech Stack:** Go `testing`, `go vet`, Docker Compose, Markdown.

**Spec:** `backend/docs/tasks/BE-007-backend-acceptance-whitehat.md`

## Global Constraints

- Depends on BE-001 through BE-006.
- Test local systems only; do not attack external systems.
- Publish findings as Critical, High, Medium, or Low with reproduction evidence.
- After one remediation package, escalate a new blocker instead of opening a third QA round.

---

### Task 1: Build acceptance and negative test matrix

**Files:**
- Create: `backend/docs/qa/BE-007-acceptance-matrix.md`
- Create: `backend/internal/qa/acceptance_test.go`
- Test: `backend/internal/qa/acceptance_test.go`

**Interfaces:**
- Consumes: auth, product, regulation, notification, and recognition HTTP contracts from BE-003 through BE-006.
- Produces: scenario IDs that link each task acceptance criterion to an automated or manual local check.

- [ ] **Step 1: Write acceptance scenarios first**

Create rows for registration case-variant rejection, invalid password, absent/tampered session, foreign product/draft ID, plain-password absence, status boundaries, duplicate notification, unverified rule, expired `use_by` recipe exclusion, `best_before` eligibility, reject, and duplicate approve.

- [ ] **Step 2: Write a failing cross-account API test**

```go
func TestAcceptanceForeignAccountCannotReadProduct(t *testing.T) {
    rr := callAs(secondAccount, http.MethodGet, "/v1/products/"+firstProductID.String(), nil)
    if rr.Code != http.StatusNotFound { t.Fatal(rr.Code) }
}
```

- [ ] **Step 3: Run the failing test**

Run: `cd backend && go test ./internal/qa -run TestAcceptanceForeignAccountCannotReadProduct -v`

Expected: FAIL until the dependent APIs are complete.

- [ ] **Step 4: Complete the matrix and run tests**

Run: `cd backend && go test ./internal/qa -v`

Expected: PASS with a result linked to every matrix row.

### Task 2: Execute local security review and publish evidence

**Files:**
- Create: `backend/docs/qa/BE-007-whitehat-report.md`
- Test: local Compose, Go test, and vet command output

**Interfaces:**
- Produces: report fields `finding_id`, severity, affected task/API, reproduction, expected result, actual result, evidence, and remediation status.

- [ ] **Step 1: Run static repository checks**

Run: `rg -n --glob '!sessions/**' 'password.*[:=].*["'\'' ].+|API[_-]?KEY|SECRET' backend`

Expected: any match is assessed; no real secret or plaintext password is accepted.

- [ ] **Step 2: Run full local verification**

Run: `cd backend && go test ./... && go vet ./... && docker compose up --build -d`

Expected: all tests and vet pass, and the local API starts.

- [ ] **Step 3: Publish the report**

Record command output or request/response evidence for every finding and explicitly state whether Critical or High findings remain open. If fixes are needed, request one consolidated remediation package; if a new blocker appears after its retest, mark it escalated to PM and product owner.

- [ ] **Step 4: Commit**

Run: `git add backend/docs/qa backend/internal/qa && git commit -m "test(backend): add acceptance whitehat evidence"`

## Self-review

- Coverage: all BE-003—BE-006 criteria, local-only security boundaries, severity evidence, and one-retest policy are covered.

