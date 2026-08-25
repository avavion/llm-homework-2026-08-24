# Session Accounting Hooks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish a shared repository guide and a Claude Code lifecycle hook that creates and completes department-specific session reports.

**Architecture:** A single standard-library Python command accepts Claude hook JSON on stdin or explicit CLI fields for other agent runtimes. It classifies the working directory as `backend`, `frontend`, `shared`, or the project root, writes an idempotent Markdown report in that scope's `sessions/` directory, and finalizes the same report at session end. `.claude/settings.json` invokes it for Claude Code; `AGENTS.md` makes the command the portable fallback for all other agents.

**Tech Stack:** Markdown, JSON, Python 3 standard library, Claude Code project settings.

**Spec:** Approved in chat on 2026-08-25: universal session accounting with Claude Code `SessionStart` / `SessionEnd` hooks and the same command for all agent runtimes.

## Global Constraints

- Session reports belong only in `sessions/`, `backend/sessions/`, `frontend/sessions/`, or `shared/sessions/`.
- A working directory under `backend/` maps to backend; under `frontend/` maps to frontend; locations under `shared/` map to shared.
- The project root maps to `sessions/` with department `root`; paths under `shared/` map to `shared/sessions/`.
- A repeated start event for the same `session_id` must reuse the existing report.
- No third-party package may be required by the hook.
- The hook must not block session start or end merely because report persistence fails.
- Existing report files must never be overwritten by a different session.

---

### Task 1: Document the workspace and cross-runtime protocol

**Files:**
- Create: `shared/docs/WORKSPACE_GUIDE.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Produces: the documented command `python3 scripts/session_hook.py start|end` for non-Claude runtimes.
- Produces: department ownership and report frontmatter contract used by the hook implementation.

- [ ] **Step 1: Write the workspace guide**

Document the roles of `backend/`, `frontend/`, and `shared/`; define ownership of each `docs/` and `sessions/` folder plus the root `sessions/` audit folder; declare the required frontmatter keys `hook`, `session_id`, `department`, `started_at`, and `completed_at`; list the mandatory report sections.

- [ ] **Step 2: Amend root agent instructions**

Replace the current shared-only session rule with department routing. Require non-Claude agents to run:

```sh
python3 scripts/session_hook.py start --cwd "$PWD" --session-id "<runtime-session-id>"
python3 scripts/session_hook.py end --cwd "$PWD" --session-id "<runtime-session-id>"
```

- [ ] **Step 3: Review documentation consistency**

Run: `rg -n "backend/sessions|frontend/sessions|shared/sessions|session_hook.py" AGENTS.md shared/docs/WORKSPACE_GUIDE.md`

Expected: all three departments and both portable commands are documented.

### Task 2: Implement the idempotent session hook command

**Files:**
- Create: `scripts/session_hook.py`
- Test: `tests/test_session_hook.py`

**Interfaces:**
- Consumes: `start` or `end`, a session ID, working directory, and optional JSON payload on stdin.
- Produces: one Markdown report named `session-YYYY-MM-DD-HHMMSS.md`; exit code `0` on both successful persistence and recoverable hook errors.

- [ ] **Step 1: Write failing tests for routing and start event**

Create tests using `unittest`, `tempfile.TemporaryDirectory`, and a temporary project tree. Invoke the script with `start --project-root <tmp> --cwd <tmp>/backend --session-id backend-1 --timestamp 2026-08-25T23:45:16+03:00`; assert that only `<tmp>/backend/sessions/session-2026-08-25-234516.md` exists and its frontmatter contains `hook: session.started`, `session_id: backend-1`, and `department: backend`. Add the same assertion for `<tmp>` mapping to `<tmp>/sessions/` with `department: root`.

- [ ] **Step 2: Run the routing test and confirm the expected failure**

Run: `python3 -m unittest tests.test_session_hook.SessionHookTest.test_start_routes_backend -v`

Expected: FAIL because `scripts/session_hook.py` does not exist.

- [ ] **Step 3: Implement the minimal command**

Implement `scripts/session_hook.py` with `argparse`, `json`, `pathlib`, and `datetime`. Read common Claude fields `session_id` and `cwd` from stdin JSON when CLI flags are absent. Route from the resolved working directory: the project root maps to `sessions/`; `backend/`, `frontend/`, and `shared/` map to their own `sessions/` folders. On `start`, search all four report directories for the session ID before creating the timestamped file; create a Markdown report with `hook: session.started`, empty `completed_at`, and sections for request, actions, sources, artifacts, conclusions, risks, self-criticism, and next step. On `end`, find the existing report by session ID and replace only the `hook` and `completed_at` frontmatter values with `session.completed` and the current ISO timestamp.

- [ ] **Step 4: Run the routing test and confirm it passes**

Run: `python3 -m unittest tests.test_session_hook.SessionHookTest.test_start_routes_backend -v`

Expected: PASS.

- [ ] **Step 5: Write failing tests for idempotency and completion**

Add tests that call `start` twice with the same session ID and assert that one report remains, then call `end` and assert that the same file contains `hook: session.completed` and a non-empty `completed_at`.

- [ ] **Step 6: Run the new tests and confirm expected failure**

Run: `python3 -m unittest tests.test_session_hook.SessionHookTest.test_start_is_idempotent tests.test_session_hook.SessionHookTest.test_end_finalizes_existing_report -v`

Expected: FAIL until idempotent lookup and frontmatter update are implemented.

- [ ] **Step 7: Implement idempotency and finalization**

Add session-ID lookup across all four `sessions/` directories, preserve the original report path during `end`, and keep every error non-blocking by emitting a concise message to stderr and returning `0`.

- [ ] **Step 8: Run all hook tests**

Run: `python3 -m unittest tests.test_session_hook -v`

Expected: PASS with routing, idempotency, and completion coverage.

### Task 3: Connect Claude Code lifecycle events

**Files:**
- Create: `.claude/settings.json`

**Interfaces:**
- Consumes: Claude Code `SessionStart` and `SessionEnd` JSON payloads on stdin.
- Produces: calls to `python3 "$CLAUDE_PROJECT_DIR/scripts/session_hook.py" start` and `python3 "$CLAUDE_PROJECT_DIR/scripts/session_hook.py" end`.

- [ ] **Step 1: Add project hook configuration**

Configure `SessionStart` with matcher `startup|resume` and `SessionEnd` without a matcher. Each command pipes Claude's JSON input directly to the script and passes its lifecycle action as the sole argument. Set a short command timeout so reporting cannot delay a session.

- [ ] **Step 2: Validate JSON and command references**

Run: `python3 -m json.tool .claude/settings.json >/dev/null && rg -n "SessionStart|SessionEnd|session_hook.py" .claude/settings.json`

Expected: valid JSON containing both events and the single shared script.

### Task 4: Perform end-to-end verification

**Files:**
- Verify: `scripts/session_hook.py`, `tests/test_session_hook.py`, `.claude/settings.json`, `AGENTS.md`, `shared/docs/WORKSPACE_GUIDE.md`

- [ ] **Step 1: Simulate backend, frontend, and shared starts**

Run the command four times with temporary project roots, fixed timestamps, and working directories in the root, backend, frontend, and shared scopes. Verify that each report is created in its matching `sessions/` directory.

- [ ] **Step 2: Run all automated checks**

Run: `python3 -m unittest tests.test_session_hook -v && python3 -m json.tool .claude/settings.json >/dev/null && git diff --check`

Expected: all tests pass, settings parse, and Git reports no whitespace errors.

- [ ] **Step 3: Review the final diff**

Run: `git status --short && git diff --check`

Expected: only the documented guide, portable hook, hook tests, Claude settings, and root instructions are changed.
