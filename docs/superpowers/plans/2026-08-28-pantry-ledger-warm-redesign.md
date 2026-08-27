# Pantry Ledger Warm Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the approved warm, home-like Pantry Ledger interface on desktop, tablet and mobile without changing existing backend routes or request payloads.

**Architecture:** Keep React Router, React Query, React Hook Form, Zod, fixture mode and every existing HTTP path. Split visual primitives and the responsive shell out of the large route file only where the resulting component has a single responsibility. The client must render server-provided data as-is; backend display semantics are an explicit dependency, never inferred from the browser clock.

**Tech Stack:** React 19, TypeScript, React Router 7, TanStack Query 5, React Hook Form 7, Zod 4, Sass, Vitest, Testing Library and Playwright.

**Spec:** `docs/superpowers/specs/2026-08-28-pantry-ledger-warm-redesign-design.md`

## Global Constraints

- Preserve every existing route, cookie session behavior, API URL, HTTP method, request field and backend response field unless a separately approved backend contract changes it.
- Do not compute `attention`, `expired`, `research_required`, recipe eligibility, regulator group or notification rules on the client.
- The production UI is light-only; remove the user-selectable light/dark/system setting and its local-storage write. This is a product-scope change from the current UI, not a backend change.
- Use inline SVG with a single stroke language; do not add an icon package, webfont dependency or styling framework.
- Keep visible focus, named controls, semantic status text, 44px touch targets and `prefers-reduced-motion: reduce` throughout.
- Preserve all loading, empty, failure, draft, lifecycle and `research_required` states from the product requirements.
- Do not edit `backend/` implementation or migrations as part of this plan. Backend work starts only after its owner accepts the contract request below.

---

## File structure

| File | Responsibility |
| --- | --- |
| `shared/docs/backend-requests/2026-08-28-product-display-contract.md` | Precise request to backend for missing display semantics; no assumed implementation. |
| `frontend/src/api.ts` | Typed adapters only; maps the approved response contract without changing request payloads. |
| `frontend/src/mock-api.ts` | Fixture parity with the approved display-status contract for UI tests and local review. |
| `frontend/src/ui.tsx` | SVG icon primitives, semantic badges, alerts, toast region, responsive navigation and bottom add sheet. |
| `frontend/src/app.tsx` | Route-level composition, form/dialog state and existing mutations. |
| `frontend/src/styles/tokens.scss` | Warm light-only semantic design tokens and motion tokens. |
| `frontend/src/styles/global.scss` | Desktop rail, tablet top nav, mobile bottom nav/sheet, component states and responsive layout. |
| `frontend/src/app.test.tsx` | Behavioural and accessible-component tests. |
| `frontend/e2e/inventory.spec.ts` | Viewport, navigation, sheet, reduced-motion and no-overflow checks. |

### Task 1: Freeze the backend boundary and open the required contract request

**Files:**
- Create: `shared/docs/backend-requests/2026-08-28-product-display-contract.md`
- Modify: `frontend/src/api.ts` only after the request has an accepted response contract.
- Test: `frontend/src/api.test.ts` (create)

**Interfaces:**
- Consumes: current `GET /v1/products` and `GET /v1/products/{id}` `PublicProduct` from `backend/internal/product/http.go`.
- Produces: an explicit backend-owner decision before client status mapping changes.

- [ ] **Step 1: Write the backend request exactly as follows.**

```md
# Запрос к Backend: display semantics продукта для Frontend

## Контекст

Frontend не может безопасно выводить `attention`, `expired` или `research_required` из даты в браузере: момент истечения и пригодность рецепта зависят от подтверждённого регуляторного правила.

## Текущее расхождение

`GET /v1/products` и `GET /v1/products/{id}` возвращают `status`, равный только lifecycle (`active`, `used`, `discarded`). В UI уже предусмотрены display states `attention`, `expired` и `research_required`, но текущий удалённый DTO их не содержит.

## Запрашиваемое дополнение (обратно совместимое)

Добавьте в оба product response поле `display_status` со значениями `active | attention | expired | used | discarded | research_required`.

- Backend определяет значение через действующий реестр regulation rules, дату, date type и lifecycle.
- `used` и `discarded` имеют приоритет над display-статусом даты.
- Для `research_required` backend не делает медицинский или юридический вывод и не обещает автоматическое напоминание.
- Существующее `status` lifecycle-поле и все текущие request payload остаются без изменений.

## Примеры

```json
{ "id": "…", "date_type": "use_by", "status": "active", "display_status": "expired" }
{ "id": "…", "date_type": "best_before", "status": "active", "display_status": "attention" }
{ "id": "…", "date_type": "use_by", "status": "active", "display_status": "research_required" }
```

## Критерий готовности

Опубликованы JSON-примеры и contract/integration tests для list и get; 401/403/404 не раскрывают данные другого аккаунта.
```

- [ ] **Step 2: Do not modify `ApiProduct.status` yet; record the request URL/owner outcome in the frontend task register or PR description.**

The request is required because `backend/internal/product/ToPublicProduct` currently serializes `LifecycleStatus` into `status`, while the browser must not calculate regulation-derived state.

- [ ] **Step 3: After backend approval, write a failing adapter test.**

```ts
import { expect, test } from 'vitest'
import { fromApiProduct } from './api'

test('uses server display_status instead of deriving a deadline status', () => {
  expect(fromApiProduct({
    id: 'spinach', name: 'Spinach', date_type: 'best_before',
    expiry_date: '2026-08-25T00:00:00Z', status: 'active',
    display_status: 'attention',
  }).status).toBe('attention')
})
```

- [ ] **Step 4: Run the focused test before changing the adapter.**

Run: `npm --prefix frontend test -- --run src/api.test.ts`

Expected: FAIL until `display_status` is part of the accepted DTO adapter.

- [ ] **Step 5: Map only the approved optional field and retain a safe compatibility fallback.**

```ts
type ApiProduct = {
  // existing fields
  status: 'active' | 'used' | 'discarded'
  display_status?: Product['status']
}

export const fromApiProduct = (item: ApiProduct): Product => ({
  // existing mappings
  status: item.display_status ?? item.status,
})
```

The fallback may only preserve the lifecycle value; it must not derive a date result. Add the same optional field to fixture records only after the backend values are agreed.

- [ ] **Step 6: Run adapter and full unit tests.**

Run: `npm --prefix frontend test -- --run src/api.test.ts src/app.test.tsx`

Expected: PASS.

- [ ] **Step 7: Commit the contract request and adapter only if the backend owner accepts the contract.**

```bash
git add shared/docs/backend-requests/2026-08-28-product-display-contract.md frontend/src/api.ts frontend/src/api.test.ts frontend/src/mock-api.ts
git commit -m "docs: request product display-status contract"
```

### Task 2: Establish the light-only warm foundation

**Files:**
- Modify: `frontend/src/styles/tokens.scss`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`
- Modify: `frontend/e2e/inventory.spec.ts`

**Interfaces:**
- Consumes: existing CSS class names and the `data-theme` behavior in `app.tsx`.
- Produces: semantic CSS custom properties and responsive base styles used by every screen.

- [ ] **Step 1: Replace the dark-theme test with a light-only test.**

```tsx
test('does not offer a theme picker in settings', async () => {
  await renderApp('/settings', true)
  expect(await screen.findByRole('heading', { name: /настройки|settings/i })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /тёмная|dark/i })).not.toBeInTheDocument()
})
```

- [ ] **Step 2: Run the targeted test.**

Run: `npm --prefix frontend test -- --run src/app.test.tsx`

Expected: FAIL because the current Settings screen exposes the theme picker.

- [ ] **Step 3: Define the approved semantic token set and motion tokens in `tokens.scss`.**

```scss
:root {
  color-scheme: light;
  --canvas: #f8f6ef;
  --surface: #fffdfa;
  --rail: #f0f1e9;
  --ink: #273129;
  --muted: #758074;
  --brand: #506d48;
  --brand-soft: #dfead5;
  --attention: #a4671d;
  --attention-soft: #fff1dc;
  --danger: #b7473d;
  --danger-soft: #fce7e4;
  --focus: #385e9d;
  --motion-fast: 160ms;
  --motion-sheet: 220ms;
}
```

Remove `data-theme=dark` and system-dark overrides. Do not hard-code these values in components.

- [ ] **Step 4: Build the global baseline.**

Use `body { background: var(--canvas); color: var(--ink); }`, 16px base type, a consistent focus outline, 44px minimum controls and a shared `@media (prefers-reduced-motion: reduce)` block that removes non-essential animation and transition durations.

- [ ] **Step 5: Update the Playwright assertion.**

Replace the dark-theme scenario with a 320px assertion that the page has no horizontal overflow and the visible primary CTA is not covered by the fixed bottom region.

- [ ] **Step 6: Verify the foundation.**

Run: `npm --prefix frontend test && npm --prefix frontend run lint && npm --prefix frontend run build`

Expected: PASS.

- [ ] **Step 7: Commit the isolated foundation.**

```bash
git add frontend/src/styles/tokens.scss frontend/src/styles/global.scss frontend/src/app.test.tsx frontend/e2e/inventory.spec.ts
git commit -m "feat(frontend): add warm light design foundation"
```

### Task 3: Rebuild the accessible responsive application shell

**Files:**
- Modify: `frontend/src/ui.tsx`
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/app.test.tsx`
- Modify: `frontend/src/styles/global.scss`

**Interfaces:**
- Consumes: `AppShell`, `NavLink`, `t`, authenticated route layout.
- Produces: desktop rail at ≥1024px, compact tablet navigation at 640–1023px, mobile navigation + `AddProductSheet` at <640px.

- [ ] **Step 1: Write failing behavioural tests for nav and add sheet.**

```tsx
test('opens and closes the mobile add-product sheet', async () => {
  await renderApp('/products', true)
  fireEvent.click(await screen.findByRole('button', { name: /добавить продукт|add product/i }))
  expect(screen.getByRole('dialog', { name: /добавить продукт|add product/i })).toBeInTheDocument()
  fireEvent.keyDown(document, { key: 'Escape' })
  expect(screen.queryByRole('dialog', { name: /добавить продукт|add product/i })).not.toBeInTheDocument()
})
```

- [ ] **Step 2: Run the focused test.**

Run: `npm --prefix frontend test -- --run src/app.test.tsx`

Expected: FAIL because no sheet component exists.

- [ ] **Step 3: Add focused visual primitives in `ui.tsx`.**

Create `Icon`, `IconButton`, `AppShell`, `AddProductSheet`, `ToastRegion`, `StatusBadge`, `Alert`, `Skeleton` and `EmptyState`. Use inline SVG paths, not Unicode/emoji glyphs. `AddProductSheet` accepts `{ open: boolean; onClose(): void }`, renders `role="dialog" aria-modal="true"`, restores focus to its trigger and links to `/products/new` and `/products/new/photo`.

- [ ] **Step 4: Wire `ProtectedLayout` to own sheet state.**

Keep all route links and logout behavior. Render the desktop add button and mobile FAB from the same `openAddSheet` callback; do not change any backend request or navigation destination.

- [ ] **Step 5: Implement breakpoint styling.**

At desktop show the rail and full labels. At tablet show a top nav. At mobile show exactly three named destinations and a central, accessible FAB. Give sheets a backdrop, close button, 220ms transform/opacity animation and no layout-shifting animation.

- [ ] **Step 6: Run test, lint and build.**

Run: `npm --prefix frontend test -- --run src/app.test.tsx && npm --prefix frontend run lint && npm --prefix frontend run build`

Expected: PASS.

- [ ] **Step 7: Commit the shell.**

```bash
git add frontend/src/ui.tsx frontend/src/app.tsx frontend/src/app.test.tsx frontend/src/styles/global.scss
git commit -m "feat(frontend): add responsive pantry shell"
```

### Task 4: Recompose inventory, lifecycle and forms without changing mutations

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: `api.products.list/create/get/complete`, `api.drafts.*`, `ProductInput`, `StatusBadge` and existing query keys.
- Produces: approved hierarchy and interaction states with unchanged client API calls.

- [ ] **Step 1: Write the failing inventory hierarchy test.**

```tsx
test('renders at most three products in the use-first rail', async () => {
  await renderApp('/products', true)
  const rail = await screen.findByRole('region', { name: /использовать первым|use first/i })
  expect(within(rail).getAllByRole('listitem')).toHaveLength(3)
})
```

Import `within` from Testing Library. Ensure fixture data contains at least three non-terminal products before running this test.

- [ ] **Step 2: Run the test.**

Run: `npm --prefix frontend test -- --run src/app.test.tsx`

Expected: FAIL until the rail has an accessible region label and stable fixture count.

- [ ] **Step 3: Implement the desktop/tablet/mobile inventory layout.**

Keep the existing `shown` and `urgent` query-derived lists. Render urgency before inventory in DOM and visually position it by grid at desktop. Replace symbolic status marks with named SVG-supported badges; retain name, storage, date type, absolute date, status and lifecycle actions. Keep search/filter logic client-side and do not add backend query parameters.

- [ ] **Step 4: Add lifecycle confirmation before mutation.**

Use a focus-managed dialog whose props are `{ product: Product; action: 'used' | 'discarded'; onConfirm(): void; onClose(): void }`. The dialog names the product, has a text consequence, disables its confirm button while `complete.isPending`, retains error text on failure and invalidates the same product query keys on success.

- [ ] **Step 5: Reformat the product, photo and draft forms.**

Preserve `productSchema`, the existing `POST /v1/products` payload mapping and explicit draft approve/reject. Move optional storage fields into a native `<details>` section labelled «Детали хранения» without removing labels. Keep `aria-describedby`, inline errors, manual fallback after recognition failure and no product creation before approve.

- [ ] **Step 6: Add and run form/draft regression tests.**

```tsx
test('keeps manual entry available after photo recognition fails', async () => {
  await renderApp('/products/new/photo', true)
  expect(await screen.findByRole('link', { name: /ввести вручную|enter manually/i })).toBeVisible()
})
```

Run: `npm --prefix frontend test -- --run src/app.test.tsx`

Expected: PASS, including the existing no-create-before-approve test.

- [ ] **Step 7: Commit the inventory and form changes.**

```bash
git add frontend/src/app.tsx frontend/src/styles/global.scss frontend/src/app.test.tsx frontend/src/mock-api.ts
git commit -m "feat(frontend): recompose pantry inventory flows"
```

### Task 5: Apply the system to recipes, settings, public screens and feedback

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/ui.tsx`
- Modify: `frontend/src/i18n.ts`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: existing recipes response `{ title, product_ids }[]`, existing blocked settings state and all route components.
- Produces: coherent presentation without inventing profile/settings endpoints.

- [ ] **Step 1: Write failing tests for a recipe empty state and a blocked settings explanation.**

```tsx
test('keeps settings honest when profile APIs are unavailable', async () => {
  await renderApp('/settings', true)
  expect(await screen.findByText(/после согласования API|after the API is agreed/i)).toBeInTheDocument()
})
```

- [ ] **Step 2: Run the focused test.**

Run: `npm --prefix frontend test -- --run src/app.test.tsx`

Expected: PASS before styling; the test becomes the safeguard against fabricating settings functionality.

- [ ] **Step 3: Apply reusable page surfaces.**

Give recipe rows an explanatory product-count line based only on `product_ids`, render a single-CTA empty state, and preserve recipe output from the backend. Keep the current settings API-blocked wording and `research_required` alert; do not add local country, regulator or notification data forms until a backend profile/settings contract exists.

- [ ] **Step 4: Add toast feedback.**

Route successful create/approve/lifecycle completion through `ToastRegion`; use `role="status" aria-live="polite"`, a dismiss control and no focus steal. Errors remain contextual `Alert` messages.

- [ ] **Step 5: Refine login/register and not-found.**

Use the same warm public-page shell, labelled controls, clear errors, submit loading state and no private navigation.

- [ ] **Step 6: Verify full unit and type suites.**

Run: `npm --prefix frontend test && npm --prefix frontend run lint && npm --prefix frontend run build`

Expected: PASS.

- [ ] **Step 7: Commit cross-screen consistency.**

```bash
git add frontend/src/app.tsx frontend/src/ui.tsx frontend/src/i18n.ts frontend/src/styles/global.scss frontend/src/app.test.tsx
git commit -m "feat(frontend): unify pantry screens and feedback"
```

### Task 6: Validate responsive visual quality, motion and backend compatibility

**Files:**
- Modify: `frontend/e2e/inventory.spec.ts`
- Create: `frontend/e2e/accessibility.spec.ts`
- Create: `frontend/tests/visual/warm-redesign.spec.ts-snapshots/` through Playwright snapshot approval
- Modify: `frontend/docs/tasks/FE-007-qa-whitehat-and-visual-acceptance.md` only to record the executed command and baseline location.

**Interfaces:**
- Consumes: fixture-mode app, current Playwright config and the approved server display-status response when remote integration is available.
- Produces: reproducible visual and keyboard evidence; no backend writes.

- [ ] **Step 1: Add desktop, tablet and mobile screenshot assertions.**

```ts
await expect(page).toHaveScreenshot(`products-${viewport.name}.png`, {
  fullPage: true,
  maxDiffPixelRatio: 0.01,
})
```

Log in via fixture mode as in the existing spec. Cover 320×740, 768×900 and 1440×960.

- [ ] **Step 2: Add sheet keyboard and reduced-motion checks.**

```ts
await page.emulateMedia({ reducedMotion: 'reduce' })
await page.getByRole('button', { name: /добавить продукт|add product/i }).click()
await expect(page.getByRole('dialog')).toBeVisible()
await page.keyboard.press('Escape')
await expect(page.getByRole('dialog')).toBeHidden()
```

- [ ] **Step 3: Add a semantic status check guarded by the backend contract.**

For fixture mode assert visible status text for active, attention and expired fixtures. For remote mode, run only after the backend request accepts `display_status`; assert that the UI displays the server value and does not infer it from the date.

- [ ] **Step 4: Run and inspect frontend verification.**

Run: `npm --prefix frontend run test:e2e`

Expected: PASS; snapshots only change when explicitly reviewed. Inspect `frontend/test-results/` on failure.

- [ ] **Step 5: Run final non-visual verification.**

Run: `git diff --check && npm --prefix frontend test && npm --prefix frontend run lint && npm --prefix frontend run build`

Expected: every command exits 0.

- [ ] **Step 6: Commit the QA evidence.**

```bash
git add frontend/e2e frontend/tests/visual frontend/docs/tasks/FE-007-qa-whitehat-and-visual-acceptance.md
git commit -m "test(frontend): cover warm redesign acceptance"
```

## Plan self-review

- **Spec coverage:** Tasks 2–5 cover the visual system, responsive structures, all route types, states and accessibility. Task 6 covers responsive, motion and visual acceptance.
- **Backend safety:** Task 1 blocks regulation-derived UI state on an accepted backend display-status contract and keeps all existing HTTP paths/payloads unchanged. Settings remain transparently blocked because no profile/settings API is available.
- **Animation coverage:** Tasks 2, 3 and 6 cover micro-feedback, sheet motion, crossfade, reduced motion and interaction cancellation.
- **Gaps intentionally excluded:** Country/profile and notification settings cannot be implemented safely without a new backend contract; the request identifies this separate scope rather than inventing it in the client.
