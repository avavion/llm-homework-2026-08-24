# Editorial Pantry Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the React interface as the approved Editorial Pantry design while preserving every route, API contract and product safety rule.

**Architecture:** Keep the application’s current React Router, React Query, React Hook Form and SCSS Modules-free global-SCSS architecture. Introduce the Editorial Pantry token layer first, then apply it through existing shared primitives and route components. The design is CSS-first: no image library, UI kit or API changes; login/register use a responsive split layout with an original CSS food-story surface.

**Tech Stack:** React 19, TypeScript, React Router 7, React Query 5, React Hook Form + Zod, SCSS, Vitest + Testing Library, Playwright.

**Spec:** `docs/design-requirements-editorial-pantry.md`

## Global Constraints

- Preserve all current routes and API contracts; do not change backend code.
- Use original CSS artwork/gradients for auth; do not copy image assets or code from Behance.
- Use semantic tokens only in SCSS; no raw colours inside route components.
- Keep normal text contrast at 4.5:1 or greater, controls at 44×44px or larger, and a visible `:focus-visible` treatment.
- Verify 375px, 768px, 1280px and 1440px. At 1280px the outer application shell is at least 1216px wide (including navigation); the main area must not use `max-width: 1240px`.
- Respect `prefers-reduced-motion`; use only transform/opacity for UI transitions.
- Do not stage or commit pre-existing unrelated worktree changes.

## File Structure

| File | Responsibility |
| --- | --- |
| `frontend/src/styles/tokens.scss` | Editorial Pantry colour, type, spacing, z-index and motion tokens. |
| `frontend/src/styles/global.scss` | Global layout, all route skins, responsive grid and accessibility behaviour. |
| `frontend/src/styles/grid-fixes.scss` | Remove superseded layout overrides after their rules move into the single global layout source. |
| `frontend/src/ui.tsx` | Shared navigation, status, feedback and add-product sheet semantics. |
| `frontend/src/app.tsx` | Route-level semantic hooks/classes for auth, inventory, forms, detail, recipe and settings views. |
| `frontend/src/app.test.tsx` | Unit/integration assertions for auth layout hooks and accessible shared UI. |
| `frontend/e2e/inventory.spec.ts` | Responsive, keyboard and visual regression coverage including 1280px. |
| `frontend/e2e/auth.spec.ts` | New auth-layout and form-accessibility browser coverage. |

---

### Task 1: Establish Editorial Pantry tokens and desktop grid

**Files:**
- Modify: `frontend/src/styles/tokens.scss`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/styles/grid-fixes.scss`
- Modify: `frontend/e2e/inventory.spec.ts`

**Interfaces:**
- Consumes: existing classes `.shell`, `.rail`, `.page`, `.dashboard`, `.table`.
- Produces: `--canvas`, `--surface`, `--ink`, `--muted`, `--brand`, `--brand-strong`, `--olive`, `--amber`, `--danger`, `--focus`, `--font-display`, `--font-body`, `--font-data`, `--space-*`, `--z-*`; an explicit `@media (min-width: 1280px)` grid.

- [ ] **Step 1: Write the failing 1280px layout test**

  Add this viewport to `frontend/e2e/inventory.spec.ts` before the existing 1440px case:

  ```ts
  { name: 'desktop-1280', width: 1280, height: 960 },
  ```

  Add an assertion in the same test after login:

  ```ts
  const shellWidth = await page.locator('.shell').evaluate((element) => element.getBoundingClientRect().width)
  expect(shellWidth).toBeGreaterThanOrEqual(1216)
  ```

- [ ] **Step 2: Run the test to verify it fails**

  Run: `npm run test:e2e -- --grep "desktop-1280"`

  Expected: FAIL because the existing outer shell does not keep the required desktop gutter while stretching to the 1280px layout.

- [ ] **Step 3: Replace the token layer**

  In `tokens.scss`, replace legacy mint aliases with the approved semantic tokens and keep backwards-compatible names only while `global.scss` is migrated:

  ```scss
  :root {
    --canvas: #fff8f0;
    --surface: #ffffff;
    --surface-soft: #fff1e6;
    --ink: #2a211b;
    --muted: #6d6259;
    --brand: #c94b35;
    --brand-strong: #9e3525;
    --olive: #66704a;
    --amber: #a65a00;
    --danger: #b42318;
    --focus: #2c5c8a;
    --font-display: 'Playfair Display', Georgia, serif;
    --font-body: 'DM Sans', system-ui, sans-serif;
    --font-data: ui-monospace, SFMono-Regular, Menlo, monospace;
    --space-1: 4px;
    --space-2: 8px;
    --space-3: 12px;
    --space-4: 16px;
    --space-6: 24px;
    --space-8: 32px;
    --z-nav: 10;
    --z-sheet: 20;
    --z-toast: 30;
  }
  ```

- [ ] **Step 4: Consolidate responsive layout rules**

  Move the active grid rules into `global.scss`, delete duplicate `.page`, `.inventory-layout` and `.product-card` overrides from `grid-fixes.scss`, and add:

  ```scss
  @media (min-width: 1280px) {
    .shell { width: min(100% - 64px, 1600px); margin-inline: auto; grid-template-columns: 232px minmax(0, 1fr); }
    .page { width: auto; max-width: none; margin-inline: 0; padding-inline: 32px; }
    .dashboard { min-width: 0; }
    .priority { grid-template-columns: 216px repeat(3, minmax(0, 1fr)); }
  }
  ```

- [ ] **Step 5: Run focused test and type checks**

  Run: `npm run test:e2e -- --grep "desktop-1280" && npm run lint`

  Expected: PASS; `.shell` is at least 1216px at 1280px, the page has no separate narrow maximum width, and TypeScript exits 0.

- [ ] **Step 6: Commit only task files**

  ```bash
  git add frontend/src/styles/tokens.scss frontend/src/styles/global.scss frontend/src/styles/grid-fixes.scss frontend/e2e/inventory.spec.ts
  git commit -m "feat: establish editorial pantry layout tokens"
  ```

### Task 2: Implement the accessible editorial auth split layout

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`
- Create: `frontend/e2e/auth.spec.ts`

**Interfaces:**
- Consumes: existing `Login`, `Register`, `PublicPage`, React Hook Form resolver and `api.auth` methods.
- Produces: `auth-layout`, `auth-story`, `auth-panel`, `auth-form`, `auth-error-summary` classes; auth controls retain labels, `autocomplete` attributes and current submit handlers.

- [ ] **Step 1: Write failing component tests**

  Add to `app.test.tsx`:

  ```tsx
  test('renders the sign-in screen as an accessible auth layout', async () => {
    await api.auth.logout()
    await renderApp('/login')
    expect(await screen.findByTestId('auth-story')).toBeInTheDocument()
    expect(screen.getByRole('main')).toHaveClass('auth-layout')
    expect(screen.getByLabel(/пароль|password/i)).toHaveAttribute('autocomplete', 'current-password')
  })
  ```

- [ ] **Step 2: Run the new test to verify it fails**

  Run: `npm run test -- app.test.tsx -t "accessible auth layout"`

  Expected: FAIL because `auth-story` and `auth-layout` do not exist.

- [ ] **Step 3: Add the structural auth classes without changing auth logic**

  Refactor `PublicPage` so it accepts an optional `variant`, and render login/register through:

  ```tsx
  <main id="main-content" className="auth-layout">
    <aside className="auth-story" data-testid="auth-story" aria-label={t.authStoryLabel}>
      <p className="auth-story__eyebrow">Pantry Ledger</p>
      <p className="auth-story__quote">{t.authStoryQuote}</p>
    </aside>
    <section className="auth-panel"><Page {...pageProps}>{children}</Page></section>
  </main>
  ```

  Add `authStoryLabel` and `authStoryQuote` to both locale dictionaries. Keep the photo/story surface decorative beyond its text, and do not introduce external images.

- [ ] **Step 4: Make authentication form semantics complete**

  Ensure email and password fields use these exact properties:

  ```tsx
  <input type="email" autoComplete="email" {...form.register('email')} />
  <input type="password" autoComplete={isRegister ? 'new-password' : 'current-password'} {...form.register('password')} />
  ```

  On an invalid submit, render the summary before fields only when there are two or more errors:

  ```tsx
  <div className="auth-error-summary" role="alert" tabIndex={-1}>
    <h2>{t.formErrorSummary}</h2>
    <a href="#email">{t.invalidEmail}</a>
    <a href="#password">{t.passwordMin}</a>
  </div>
  ```

- [ ] **Step 5: Style the split layout mobile-first**

  In `global.scss`, make `.auth-story` a brand-colour CSS composition with an abstract produce silhouette created by pseudo-elements; hide it below 1024px. Set `.auth-panel` to 420–480px, white surface, 32px desktop padding, and no decorative shadow on mobile. Keep form fields at 44px min-height.

- [ ] **Step 6: Add browser checks**

  In `auth.spec.ts`, add:

  ```ts
  test('sign-in keeps labels and the story panel on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 })
    await page.goto('/login')
    await expect(page.getByTestId('auth-story')).toBeVisible()
    await expect(page.getByLabel(/e-mail|email/i)).toBeVisible()
    await expect(page.getByLabel(/пароль|password/i)).toBeVisible()
    expect(await page.locator('html').evaluate((node) => node.scrollWidth <= node.clientWidth)).toBe(true)
  })
  ```

- [ ] **Step 7: Run tests and commit**

  Run: `npm run test -- app.test.tsx && npm run test:e2e -- --grep "sign-in keeps labels"`

  Expected: PASS.

  ```bash
  git add frontend/src/app.tsx frontend/src/styles/global.scss frontend/src/i18n.ts frontend/src/app.test.tsx frontend/e2e/auth.spec.ts
  git commit -m "feat: redesign accessible authentication screens"
  ```

### Task 3: Restyle shared navigation, feedback and lifecycle dialog

**Files:**
- Modify: `frontend/src/ui.tsx`
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: `AppShell`, `AddProductSheet`, `ToastRegion`, `StatusBadge`, `LifecycleDialog`.
- Produces: `nav` with label-and-icon destinations, editorial add sheet, a full Tab focus trap for lifecycle dialogs, and token-driven status tones.

- [ ] **Step 1: Write failing keyboard tests**

  Add an integration test that opens the lifecycle confirmation and asserts that Shift+Tab and Tab keep focus inside `[role="dialog"]`:

  ```tsx
  test('keeps keyboard focus inside the lifecycle dialog', async () => {
    await renderApp('/', true)
    fireEvent.click(await screen.findByRole('button', { name: /использован|used/i }))
    const dialog = screen.getByRole('dialog')
    const buttons = screen.getAllByRole('button', { hidden: false }).filter((button) => dialog.contains(button))
    buttons.at(-1)?.focus()
    fireEvent.keyDown(document, { key: 'Tab' })
    expect(buttons[0]).toHaveFocus()
  })
  ```

- [ ] **Step 2: Run it to verify the current dialog fails**

  Run: `npm run test -- app.test.tsx -t "lifecycle dialog"`

  Expected: FAIL because `LifecycleDialog` only handles Escape.

- [ ] **Step 3: Extract and reuse the focus trap**

  Add `trapDialogTabKey(event: KeyboardEvent, root: HTMLElement | null): void` in `ui.tsx`, using the existing `focusableSelector`. In `LifecycleDialog`, call this function for `Tab` in the document keydown listener. Do not alter mutation handling or focus return logic.

- [ ] **Step 4: Apply the visual system to shared components**

  Use classes already exposed by `ui.tsx` to create the following CSS treatments: cream rail with tomato active pill, 44px labelled mobile nav items, a tomato circular add action, olive/amber/danger status badges with icon and text, a 20px white sheet/dialog, and toast text that does not steal focus. Keep all icons from `lucide-react` and mark icons adjacent to visible text `aria-hidden`.

- [ ] **Step 5: Run checks and commit**

  Run: `npm run test -- app.test.tsx && npm run lint`

  Expected: PASS; the dialog Tab test passes and no TypeScript error is introduced.

  ```bash
  git add frontend/src/ui.tsx frontend/src/app.tsx frontend/src/styles/global.scss frontend/src/app.test.tsx
  git commit -m "feat: polish accessible editorial navigation and dialogs"
  ```

### Task 4: Redesign inventory and product-detail information hierarchy

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`
- Modify: `frontend/e2e/inventory.spec.ts`

**Interfaces:**
- Consumes: `ProductList`, `ProductDetail`, `LifecycleActions`, `StatusBadge`, `productKeys`.
- Produces: `inventory-header`, `priority`, `inventory-table`, `product-detail` semantic class hooks; no product shape or endpoint change.

- [ ] **Step 1: Write the inventory hierarchy test**

  Add:

  ```tsx
  test('keeps product safety information in each inventory row', async () => {
    await renderApp('/', true)
    const table = await screen.findByRole('table', { name: /мои продукты|my products/i })
    expect(table.getByText(/годен до|use by/i)).toBeInTheDocument()
    expect(table.getByText(/требует внимания|needs attention/i)).toBeInTheDocument()
  })
  ```

- [ ] **Step 2: Verify current test result**

  Run: `npm run test -- app.test.tsx -t "product safety information"`

  Expected: PASS or FAIL is acceptable here; record the result before changing layout. This is a characterization test that protects the non-negotiable safety copy during a visual rewrite.

- [ ] **Step 3: Add editorial information hooks**

  Keep existing visible strings and table roles. Add `inventory-header` to the dashboard hero, `priority-grid` to urgent cards, and `product-date` to the date cell. In `ProductDetail`, group status and date in a `section` with `aria-labelledby` rather than changing its API query or lifecycle actions.

- [ ] **Step 4: Implement responsive CSS**

  At 1280px, set the dashboard main track to `minmax(0, 1fr)` and priority cards to three equal columns. Between 640px and 1023px, collapse priority into readable rows. Below 640px, preserve name, date type and status in each product row; never hide date type. Use `font-variant-numeric: tabular-nums` on `.product-date time` and table date cells.

- [ ] **Step 5: Refresh visual coverage**

  Update screenshot expectations only after reviewing the new images. Add this visual case:

  ```ts
  test('inventory uses the wide editorial grid at 1280px', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 960 })
    await login(page)
    await expect(page.locator('.priority')).toBeVisible()
    await expect(page).toHaveScreenshot('inventory-desktop-1280.png', { fullPage: true })
  })
  ```

- [ ] **Step 6: Run tests and commit**

  Run: `npm run test -- app.test.tsx && npm run test:e2e -- --grep "inventory"`

  Expected: PASS after intentional snapshot update.

  ```bash
  git add frontend/src/app.tsx frontend/src/styles/global.scss frontend/src/app.test.tsx frontend/e2e/inventory.spec.ts frontend/e2e/inventory.spec.ts-snapshots
  git commit -m "feat: redesign inventory information hierarchy"
  ```

### Task 5: Redesign forms, drafts, recipes and settings while preserving feedback states

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/styles/global.scss`
- Modify: `frontend/src/app.test.tsx`
- Modify: `frontend/e2e/auth.spec.ts`

**Interfaces:**
- Consumes: `ProductForm`, `PhotoUpload`, `DraftReview`, `Recipes`, `Settings`, all current API calls and the existing Zod `productSchema`.
- Produces: `form-section`, `upload-card`, `draft-banner`, `recipe-grid`, `settings-group`, and an accessible form-error summary.

- [ ] **Step 1: Write failing form-error coverage**

  Add this test:

  ```tsx
  test('shows linked product form errors without losing entered values', async () => {
    await renderApp('/products/new', true)
    fireEvent.click(await screen.findByRole('button', { name: /сохранить продукт|save product/i }))
    const error = await screen.findByRole('alert')
    expect(error).toBeInTheDocument()
    expect(screen.getByLabel(/название|product name/i)).toHaveValue('')
  })
  ```

- [ ] **Step 2: Run it to establish form behaviour**

  Run: `npm run test -- app.test.tsx -t "linked product form errors"`

  Expected: FAIL if no focusable summary exists; use the failure to drive the summary implementation.

- [ ] **Step 3: Implement grouped form structure**

  Wrap mandatory product fields in `<fieldset className="form-section">` with a visible `<legend>`. Keep optional storage fields inside the existing `<details>`. Add a safety helper below date type and make every error id stable through the existing `errorId(field)` helper. On failed submit, focus the error summary if present, otherwise focus the first invalid field.

- [ ] **Step 4: Add route-specific visual treatments**

  Style the upload card as a generous dashed surface with a vector camera icon; style the draft page with a text-and-icon `draft-banner`; make recipes a responsive 1/2/3-column grid that retains the product-and-date safety explanation; split settings into visually independent `settings-group` cards. Do not use a photo as the only instruction or status signal.

- [ ] **Step 5: Add a browser empty/error state check**

  In `auth.spec.ts`, add a test that visits a protected form, submits it empty, and asserts a visible `role="alert"` and a fully visible focused element. Use `page.locator(':focus').boundingBox()` and compare its bounds to the viewport.

- [ ] **Step 6: Run checks and commit**

  Run: `npm run test && npm run lint && npm run test:e2e -- --grep "form|auth"`

  Expected: PASS; empty submit preserves field state and provides a recovery path.

  ```bash
  git add frontend/src/app.tsx frontend/src/styles/global.scss frontend/src/app.test.tsx frontend/e2e/auth.spec.ts
  git commit -m "feat: redesign editorial forms and supporting screens"
  ```

### Task 6: Complete visual QA, reduced-motion and regression verification

**Files:**
- Modify: `frontend/e2e/inventory.spec.ts`
- Modify: `frontend/e2e/auth.spec.ts`
- Modify: `frontend/e2e/inventory.spec.ts-snapshots/*`
- Create: `frontend/e2e/auth.spec.ts-snapshots/*` (generated by Playwright)
- Modify: `docs/design-requirements-editorial-pantry.md` (mark only verified checklist items)

**Interfaces:**
- Consumes: all completed route components, CSS tokens and existing Playwright `login(page)` helper.
- Produces: repeatable visual baselines at 375px, 768px, 1280px and 1440px and checked acceptance criteria backed by commands.

- [ ] **Step 1: Write missing viewport coverage**

  Define a shared array in each spec:

  ```ts
  const editorialViewports = [
    { name: 'mobile-375', width: 375, height: 812 },
    { name: 'tablet-768', width: 768, height: 900 },
    { name: 'desktop-1280', width: 1280, height: 960 },
    { name: 'desktop-1440', width: 1440, height: 960 },
  ]
  ```

- [ ] **Step 2: Add overflow and reduced-motion assertions**

  For each viewport, assert:

  ```ts
  expect(await page.locator('html').evaluate((node) => node.scrollWidth <= node.clientWidth)).toBe(true)
  ```

  Keep the existing reduced-motion sheet test and add a check that `transitionDuration` is `0s` for `.add-sheet` under reduced motion.

- [ ] **Step 3: Run visual tests and inspect all screenshots**

  Run: `npm run test:e2e -- --update-snapshots`

  Inspect the generated desktop-1280, mobile-375 and auth desktop snapshots. Reject any clipped labels, hidden focus, low-contrast text, unintended horizontal scroll, or an auth story surface shown on mobile.

- [ ] **Step 4: Run final non-visual verification**

  Run: `npm run test && npm run lint && npm run build`

  Expected: each command exits 0.

- [ ] **Step 5: Update only proven acceptance criteria and commit**

  Mark the corresponding verified checklist lines in `docs/design-requirements-editorial-pantry.md`; do not mark manual contrast checks unless they were actually inspected.

  ```bash
  git add frontend/e2e frontend/src/styles docs/design-requirements-editorial-pantry.md
  git commit -m "test: verify editorial pantry redesign"
  ```

## Self-Review

### Spec coverage

| Specification area | Implemented by |
| --- | --- |
| Editorial palette, type, spacing, motion | Task 1 |
| Desktop ≥1280 wide layout | Tasks 1 and 4 |
| Auth split layout and own visuals | Task 2 |
| Navigation, status, dialogs, toasts | Task 3 |
| Inventory, date semantics and product detail | Task 4 |
| Manual/photo forms, drafts, recipes, settings | Task 5 |
| Responsive, a11y, Playwright visual checks | Task 6 |

### Placeholder scan

The plan contains no deferred work markers or vague implementation-only steps; each test, command and interface is concrete.

### Type consistency

The plan only consumes existing public component names (`AppShell`, `ProductForm`, `ProductList`, `ProductDetail`, `api.auth`) and defines the new helper signature `trapDialogTabKey(event, root)`. It does not add or rename API types, endpoints or product fields.
