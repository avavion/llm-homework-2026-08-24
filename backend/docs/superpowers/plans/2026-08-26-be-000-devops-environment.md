# BE-000 DevOps Environment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide a reproducible local Go API and PostgreSQL environment with health checks and documented cross-platform builds.

**Architecture:** Docker Compose runs the API and an internal PostgreSQL service on a named volume. Environment values are supplied through an ignored local `.env`; the tracked example contains no credentials.

**Tech Stack:** Docker Compose, Go, PostgreSQL, Make.

**Spec:** `backend/docs/tasks/BE-000-devops-environment.md`

## Global Constraints

- Do not add production deployment, CI/CD, cloud resources, real e-mail/OCR/LLM credentials, or business logic.
- Pin image versions; do not publish the PostgreSQL port unless the reason is documented.
- The environment must work on Linux, macOS, and Windows without committing secrets.

---

### Task 1: Define the secure local runtime

**Files:**
- Create: `backend/docker-compose.yml`, `backend/Dockerfile`, `backend/.env.example`, `backend/.gitignore`
- Test: Docker Compose configuration check

**Interfaces:**
- Produces: services `api` and `postgres`, API health endpoint `GET /healthz`, and named volume `postgres_data`.

- [ ] **Step 1: Write the environment contract**

Record `DATABASE_URL`, `API_PORT`, `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD` as variable names in `.env.example`; use only non-secret example values.

- [ ] **Step 2: Implement Compose services**

Configure the API to depend on a PostgreSQL health check and configure PostgreSQL with an internal network and `postgres_data`. Add `.env` to `.gitignore`.

- [ ] **Step 3: Validate configuration before startup**

Run: `cd backend && docker compose config`

Expected: the resolved configuration contains `api`, `postgres`, and `postgres_data`, and contains no real credential.

### Task 2: Add operator commands and documentation

**Files:**
- Create: `backend/Makefile`, `backend/docs/environment.md`
- Test: Make target dry runs and documented command review

**Interfaces:**
- Produces: targets `up`, `down`, `logs`, `migrate-up`, `test`, `build-linux-amd64`, `build-darwin-amd64`, `build-darwin-arm64`, and `build-windows-amd64`.

- [ ] **Step 1: Add deterministic Make targets**

Use `docker compose up --build -d` for `up`, `docker compose down -v` for `down`, and Go environment variables `GOOS` and `GOARCH` for each cross-build target.

- [ ] **Step 2: Document supported workflow**

State exact Go and PostgreSQL versions, startup and shutdown commands, migration command, build artifacts, and the limitation that the environment is local only.

- [ ] **Step 3: Verify lifecycle**

Run: `cd backend && make up && curl --fail http://localhost:${API_PORT:-8080}/healthz && make down`

Expected: health returns success and shutdown removes the named volume.

- [ ] **Step 4: Commit**

Run: `git add backend/docker-compose.yml backend/Dockerfile backend/.env.example backend/.gitignore backend/Makefile backend/docs/environment.md && git commit -m "chore(backend): add local devops environment"`

## Self-review

- Coverage: Compose, health, secrets, volume, Make targets, cross-builds, and environment documentation are covered.
- No automatic production deployment or secret storage is introduced.

