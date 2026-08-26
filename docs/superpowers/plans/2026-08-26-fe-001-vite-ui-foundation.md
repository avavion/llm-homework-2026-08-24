# FE-001 Vite UI Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a runnable Vite React/TypeScript foundation and an internal accessible UI library that renders every MVP screen shell from the Design Requirements without production API integration.

**Architecture:** `frontend/src/ui` holds token-driven primitives; `frontend/src/features` holds screen-specific compositions; `frontend/src/app` owns router, providers and shell. Screen fixtures intentionally model loading, empty, error and safety states so later API tasks replace only data boundaries.

**Tech Stack:** Vite, React, TypeScript, React Router, TanStack React Query, SCSS Modules, Vitest, React Testing Library.

**Spec:** `frontend/docs/design-requirements.md`, `frontend/docs/tasks/FE-001-ux-system-and-responsive-spec.md`

## Global Constraints

- Use internal project components only; do not add `@hkgtlb/ui`.
- Follow Design Requirements tokens, semantic markup, keyboard focus and breakpoints 320px, 768px and 1440px.
- Do not connect production API, calculate food-safety status in the browser, add push, shopping, family sharing, barcode or external recipes.
- All visible copy is localized through the project i18n dictionary; both ru and en are provided.

---

### Task 1: Scaffold the Vite application

**Files:**
- Create: `frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig*.json`, `frontend/index.html`
- Create: `frontend/src/main.tsx`, `frontend/src/app/App.tsx`
- Test: `frontend/src/app/App.test.tsx`

**Interfaces:**
- Produces `App`, mounted by `main.tsx`, and package scripts `dev`, `build`, `test`, `lint`.

- [x] Write a failing render test that expects the skip link and primary page heading.
- [x] Configure Vite, TypeScript, Vitest and React Testing Library.
- [x] Implement the minimal provider tree and `App` that makes the render test pass.
- [x] Run `npm run test` and `npm run build`.

### Task 2: Create tokens and UI primitives

**Files:**
- Create: `frontend/src/styles/tokens.scss`, `frontend/src/styles/global.scss`
- Create: `frontend/src/ui/{AppShell,Button,Alert,EmptyState,Skeleton,StatusBadge}.tsx` and matching `*.module.scss`
- Test: `frontend/src/ui/StatusBadge.test.tsx`, `frontend/src/ui/AppShell.test.tsx`

**Interfaces:**
- Produces typed primitives used by page compositions. `StatusBadge` accepts only `active | attention | expired | used | discarded | research_required`.

- [x] Write failing tests for accessible status text and skip-link target.
- [x] Add semantic color, spacing, type and breakpoint tokens from the Design Requirements.
- [x] Implement primitives with visible focus and plain-text content rendering.
- [x] Run focused tests and build.

### Task 3: Compose routed screen shells and state fixtures

**Files:**
- Create: `frontend/src/app/routes.tsx`, `frontend/src/app/i18n.ts`
- Create: `frontend/src/features/{auth,products,drafts,recipes,settings,not-found}/` page modules and styles
- Test: `frontend/src/app/routes.test.tsx`

**Interfaces:**
- Produces public and private screen shells for every Design Requirements route. API tasks replace fixtures behind feature boundaries.

- [x] Write failing route tests for the products shell and 404; remaining routes are covered by rendered shells pending their API tasks.
- [x] Add ru/en-ready copy boundary and fixture-driven loading, empty, error and `research_required` states.
- [x] Implement page shells, navigation and responsive composition with no production requests.
- [x] Run the full unit test suite and production build.

### Task 4: Verify the FE-001 deliverable

**Files:**
- Modify: `frontend/docs/tasks/FE-001-ux-system-and-responsive-spec.md`
- Test: all FE-001 unit tests and Vite production build

- [x] Check each FE-001 acceptance condition against the rendered shell.
- [x] Mark FE-001 complete only after the checks pass.
- [x] Record the commands and result in the session report.
