# Universal Session Logging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` for task-by-task execution.

**Goal:** Make the dialogue session format mandatory and verifiable for root, backend, frontend, and shared agents.

**Architecture:** One existing hook creates routed reports; three local instruction files require its use. A standard-library validator scans only new-format reports and checks visible metadata, dialogue blocks, mandatory sections, and final verdict fields.

**Tech Stack:** Python 3 standard library, Markdown, unittest.

**Spec:** `docs/superpowers/specs/2026-08-26-universal-session-logging-design.md`

## Global Constraints

- Do not rewrite historical reports.
- New reports use the template emitted by `scripts/session_hook.py`.
- The validator exits non-zero on invalid new-format reports.
- The hook itself remains non-blocking.

---

### Task 1: Test and implement report validation

**Files:**
- Create: `scripts/validate_session_reports.py`
- Modify: `tests/test_session_hook.py`

- [ ] Write a failing subprocess test that creates valid new-format reports in all four session folders and expects validator exit code `0`.
- [ ] Write a failing test with an empty new-format report and expect non-zero exit plus its path.
- [ ] Implement the standard-library validator and run `python3 -m unittest tests.test_session_hook`.

### Task 2: Make the protocol local to each department

**Files:**
- Create: `backend/AGENTS.md`, `frontend/AGENTS.md`, `shared/AGENTS.md`
- Modify: `AGENTS.md`, `shared/docs/WORKSPACE_GUIDE.md`

- [ ] Require start hook with `--agent`, actual dialogue blocks, final verdict, and `python3 scripts/validate_session_reports.py --project-root .` before handoff.
- [ ] Document which directory each local agent owns and that legacy reports are archival.
- [ ] Run `rg` to confirm all four agent instructions require the validator.

### Task 3: Verify the repository contract

**Files:**
- Verify: hook, validator, four `AGENTS.md`, guide, tests

- [ ] Run the full unittest module.
- [ ] Run the validator against the current repository.
- [ ] Parse Claude settings and run `git diff --check`.
