# BE-003 Authentication and Sessions — implementation report

## Status

Implementation commit `fad6b19` completed both Task 1 and Task 2 from
`backend/docs/superpowers/plans/2026-08-26-be-003-authentication-and-sessions.md`
and fix round 1 is ready for independent re-review. BE-004 and later backend
tasks were not started.

## Scope delivered

- PostgreSQL `accounts` and `auth_sessions` schema with normalized unique
  e-mail, UUID identifiers, cascading session deletion, SHA-256-length and
  expiry constraints/indexes.
- Argon2id password hashing with a fresh cryptographically random 16-byte salt,
  OWASP-aligned parameters from the approved MVP plan (`m=19456`, `t=2`,
  `p=1`), and a self-describing encoded value.
- `Register`, `Login`, `Logout`, and session resolution services behind an
  explicit repository interface.
- Opaque 32-byte random session token; only SHA-256(token) is passed to
  persistence.
- `POST /v1/auth/register`, `POST /v1/auth/login`,
  `POST /v1/auth/logout`, and `GET /v1/auth/session`.
- Cookie `session` with `Path=/`, `HttpOnly`, `Secure`, `SameSite=Lax`, and a
  30-day expiry. Login JSON does not contain the raw token; logout clears the
  cookie and deletes the server-side session.
- Reusable session middleware and `AccountFromContext(context.Context)` for
  downstream account-scoped handlers.
- Generic login failure for both unknown e-mail and wrong password; missing,
  invalid, expired, and logged-out sessions return the same unauthorized
  response.

## TDD evidence

### Task 1 RED

The first attempt was intentionally not accepted as valid RED because it
stopped at a not-yet-declared `google/uuid` test dependency. The test fake was
corrected without adding production code. The repeated command was:

```text
go test ./internal/auth -run TestRegisterNormalizesAndRejectsCaseVariant -v
```

It failed for the intended missing feature:

```text
internal/auth/service_test.go:11:2: package
llm-homework/backend/internal/account is not in std
FAIL llm-homework/backend/internal/auth [setup failed]
```

### Task 1 GREEN

```text
go test ./internal/auth -v
PASS
ok llm-homework/backend/internal/auth 0.813s
```

Five service/password/session tests passed: normalization and case-variant
duplicate rejection, password secrecy and unique salts, password verification,
SHA-256-only session persistence, generic invalid credentials, and logout.

### Task 2 RED

```text
go test ./internal/auth -run 'TestLoginSetsSafeSessionCookie|TestProtectedEndpoint' -v
```

It failed on the intended missing HTTP/middleware API:

```text
undefined: NewHandler
undefined: SessionCookieName
undefined: RequireSession
undefined: AccountFromContext
FAIL llm-homework/backend/internal/auth [build failed]
```

### Task 2 GREEN

```text
go test ./internal/auth -v
PASS
ok llm-homework/backend/internal/auth 1.049s
```

Eleven tests passed across service and HTTP behavior, including secure cookie
attributes, absence/tampering rejection, context propagation, generic public
login errors, session invalidation, and normalized registration response. The
subsequent `ServiceAPI` boundary refactor stayed green.

## Test database strategy

The BE-003 plan did not prescribe a test database strategy. The selected
minimal production-aligned strategy is:

1. Fast default Go tests use an in-memory implementation of the real auth
   repository interface. They exercise real Argon2id, token generation,
   service logic, middleware, JSON handlers, and cookie serialization without
   making `go test ./...` depend on Docker.
2. Repository SQL and migration behavior are verified through the real API and
   PostgreSQL 17.6 in an isolated Compose project named `be003-auth`. This tests
   the same pgx/database/sql adapter, migration binary, schema, and composition
   root used by the local service, rather than a different embedded database or
   SQL mock.

This avoids SQLite/PostgreSQL semantic drift and avoids adding a heavy database
test dependency. The isolated Docker stack and its disposable volume were
removed after the check.

The plan listed `000002_accounts_auth.sql`, but BE-001 selected
`golang-migrate`, which only discovers directional names. Therefore the change
uses `000002_accounts_auth.up.sql` and `000002_accounts_auth.down.sql`; the
verified migration output was:

```text
1/u init
2/u accounts_auth
```

## Verification evidence

Focused and full checks:

```text
go test ./...
ok llm-homework/backend/cmd/api
?  llm-homework/backend/internal/account [no test files]
ok llm-homework/backend/internal/auth
ok llm-homework/backend/internal/config
ok llm-homework/backend/internal/http

go vet ./...
exit 0
```

The real PostgreSQL/HTTP scenario produced:

```text
register=201 duplicate=409 login=200 session=200 logout=204
post_logout=401 db_account_checks=true,true,true
token_hash_matches_sha256=true remaining_sessions=0
```

`db_account_checks` means the stored e-mail equals `user@example.com`, the
stored password begins with the Argon2id encoding marker, and it does not equal
the submitted password. The E2E script did not print the raw cookie token or
its stored hash.

## Files changed

- `backend/migrations/000002_accounts_auth.up.sql`
- `backend/migrations/000002_accounts_auth.down.sql`
- `backend/internal/account/repository.go`
- `backend/internal/auth/password.go`
- `backend/internal/auth/service.go`
- `backend/internal/auth/service_test.go`
- `backend/internal/auth/http.go`
- `backend/internal/auth/middleware.go`
- `backend/internal/auth/http_test.go`
- `backend/internal/http/server.go`
- `backend/go.mod`
- `backend/go.sum`

No BE-002 or shared documentation was changed.

## Self-review and concerns

- The raw token necessarily exists transiently inside the service/HTTP process,
  but it crosses the HTTP boundary only as the protected cookie and is never
  persisted or returned in JSON.
- Session lookup checks expiry but does not proactively delete expired rows;
  cleanup is an operational concern outside BE-003 and does not permit expired
  authentication.
- SameSite=Lax is the required baseline. Broader CSRF controls, rate limiting,
  e-mail verification, and product-resource authorization are separate tasks.
- The registration conflict is intentionally explicit to satisfy the approved
  duplicate-e-mail acceptance criterion. Login errors remain generic and do
  not reveal account existence.
- Account isolation for future resources is enabled by a typed account in
  request context; product-specific 404/401 enforcement remains owned by
  BE-004, as stated by the plan.

## Fix round 1 — technical review findings

### Findings and root causes

1. **Cookie survived a failed server-side logout.** `logoutHandler` returned
   the `500` response before calling `http.SetCookie`, so the browser received
   no expired cookie when `DeleteSession` failed. The cookie is now expired
   before attempting repository deletion; the server still returns `500` to
   report that server-side invalidation was not confirmed.
2. **The PostgreSQL evidence was not reproducible from the repository.** The
   initial report recorded a manual isolated Compose flow, but no committed
   target rebuilt that evidence. A tagged integration test, Compose test
   service, `make test-integration`, and exact environment documentation now
   provide that path.
3. **Expiry and context identity had coverage gaps.** Expiry was enforced by
   the repository/fake but had no named service or HTTP test. The fake also
   assigned every account the zero UUID, so the middleware test could not
   detect an incorrect account ID. The fake now creates unique UUIDs and the
   context test asserts both ID and e-mail.

### Fix-round TDD evidence

Logout error-path RED before changing production code:

```text
go test ./internal/auth -run TestLogoutExpiresCookieWhenSessionDeletionFails -count=1 -v
http_test.go:178: cookie count = 0, want 1
FAIL
```

After moving cookie expiration ahead of `Service.Logout`:

```text
=== RUN   TestLogoutExpiresCookieWhenSessionDeletionFails
--- PASS: TestLogoutExpiresCookieWhenSessionDeletionFails (0.00s)
PASS
```

Integration workflow RED before adding Make/Compose orchestration:

```text
make: *** No rule to make target `test-integration'. Stop.
```

The expiry behavior already existed, so its new characterization tests passed
against production immediately. A mutation check temporarily removed only the
expiry predicate from the in-memory repository; both tests then failed on the
observable contract:

```text
TestSessionEndpointRejectsExpiredSession: status code = 200, want 401
TestAccountForSessionRejectsExpiredSession: error = <nil>, want ErrUnauthenticated
```

Restoring the predicate returned both tests to GREEN. No production mutation
was retained.

### Reproducible PostgreSQL integration architecture

`backend/internal/auth/postgres_integration_test.go` is opt-in with the
`integration` build tag. The Compose `auth-integration-test` service uses the
existing Dockerfile `build` stage, joins only the internal `database` network,
applies repository migrations with `/out/migrate`, then runs:

```text
go test -tags=integration -count=1 ./internal/auth
```

`make test-integration` assigns a dedicated Compose project and safe internal
test database credentials, then uses a shell trap to remove containers,
network, and volume on success or failure. PostgreSQL has no published host
port. The committed integration test covers:

- concurrent case-variant registration with exactly one success and one
  `ErrEmailTaken`;
- normalized e-mail and Argon2id/non-plaintext password persistence;
- exact SHA-256(raw token) persistence and absence of the raw token in DB;
- expired-session denial;
- logout deletion of the session row.

The target passed twice from a clean disposable volume. The final run was
started with an intentionally invalid external `DATABASE_URL`; Make's fixed
internal test URL overrode it, and the output confirmed:

```text
1/u init
2/u accounts_auth
ok llm-homework/backend/internal/auth 0.881s
```

The cleanup output confirmed removal of both test containers, the internal
network, and `llm-homework-backend-integration_postgres_data`.

The final pre-commit rerun on the complete tree also applied both migrations,
reported `ok llm-homework/backend/internal/auth 0.866s`, and removed the
disposable Compose resources.

### Technical-review metadata

The task register defines the review process but has no per-row status field.
Therefore only BE-003 task metadata was updated. At fix-round handoff it
recorded `technical_review: changes_required` and readiness for independent
re-review without self-assigning approval; the reviewer owned that transition.

### Independent re-review approval

Independent re-review completed on 2026-08-26 after verification of fix commit
`76125e5`: BE-003 passed the approval gate and task metadata is now
`technical_review: approved`.
