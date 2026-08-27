# Calm Ledger Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the MVP inventory UI into a responsive Calm Ledger experience that prioritises urgent products and makes lifecycle actions clear.

**Architecture:** Keep routing and the mock API unchanged. Refine the presentational layer in `app.tsx` and `ui.tsx`, then express the visual system through the existing SCSS token and global-style files. The inventory derives an urgency subset from the loaded product list, so it stays consistent with search, filter and completion mutations.

**Tech Stack:** React 19, React Router 7, TypeScript, Sass, React Testing Library, Vitest, Vite.

**Spec:** `docs/design-requirements-redesign.md`

## Global Constraints

- Preserve existing routes, Russian/English copy source, form validation and mock API contracts.
- Use SCSS custom properties instead of introducing a styling dependency.
- Keep semantic status text; colour or symbols must not be the only status signal.
- Support 320px, 768px and 1440px without horizontal overflow.
- Keep visible keyboard focus and honour `prefers-reduced-motion: reduce`.
- Do not overwrite unrelated uncommitted work.

---

### Task 1: Define Calm Ledger visual foundations

**Files:**
- Modify: `frontend/src/styles/tokens.scss`
- Modify: `frontend/src/styles/global.scss`
- Test: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: existing classes `.shell`, `.nav`, `.page`, `.card`, `.stack`, `.badge`, `.alert`, `.skeleton`.
- Produces: tokenised responsive presentation for all existing application screens.

- [ ] **Step 1: Write the failing visual-structure test**

```tsx
test('renders a labelled application navigation', () => {
  render(<MemoryRouter initialEntries={['/products']}><App /></MemoryRouter>)
  expect(screen.getByRole('navigation', { name: /навигация|navigation/i })).toBeInTheDocument()
})
```

- [ ] **Step 2: Run the test to verify the current label does not satisfy the required copy**

Run: `npm test -- --run src/app.test.tsx`

Expected: the new assertion fails until the navigation label is aligned with the application copy.

- [ ] **Step 3: Add token families and responsive primitives**

```scss
:root {
  --canvas: #eef2f7;
  --rail: #08111f;
  --accent: #20c997;
  --radius-panel: 20px;
  --shadow-raised: 0 12px 32px rgba(8, 17, 31, .10);
}

@media (max-width: 639px) {
  .nav { position: fixed; inset: auto 0 0; }
  .page { padding-bottom: 96px; }
}
```

Implement the complete responsive treatment in `global.scss`: desktop rail, mobile bottom bar, reserved page spacing, form controls, status tones, cards, focus states, and reduced-motion override.

- [ ] **Step 4: Run the focused test and static check**

Run: `npm test -- --run src/app.test.tsx && npm run lint`

Expected: PASS.

### Task 2: Recompose the app shell and reusable status components

**Files:**
- Modify: `frontend/src/ui.tsx`
- Modify: `frontend/src/app.test.tsx`
- Test: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: `Status`, `t`, React Router `NavLink`.
- Produces: `AppShell`, `StatusBadge`, `Alert`, `Skeleton`, and `EmptyState` with accessible visual hooks used by `app.tsx`.

- [ ] **Step 1: Write failing shell and status tests**

```tsx
test('shows named navigation destinations', () => {
  render(<MemoryRouter initialEntries={['/products']}><App /></MemoryRouter>)
  expect(screen.getByRole('link', { name: /мои продукты|my products/i })).toBeInTheDocument()
  expect(screen.getByRole('link', { name: /рецепты|recipes/i })).toBeInTheDocument()
})
```

- [ ] **Step 2: Run the focused test**

Run: `npm test -- --run src/app.test.tsx`

Expected: PASS after Task 1; test protects the component contract during refactor.

- [ ] **Step 3: Implement semantic shell regions and status affordances**

```tsx
export function StatusBadge({ status }: { status: Status }) {
  return <span className={`badge badge--${status}`}>
    <span className="badge__mark" aria-hidden="true">…</span>
    {labels[status]}
  </span>
}
```

Add a branded rail header, make the nav label localised, retain its native links, and give alert, empty, and skeleton states purpose-specific classes without changing their public props.

- [ ] **Step 4: Run all frontend tests**

Run: `npm test`

Expected: PASS.

### Task 3: Build the inventory action hierarchy and premium form layout

**Files:**
- Modify: `frontend/src/app.tsx`
- Modify: `frontend/src/app.test.tsx`
- Test: `frontend/src/app.test.tsx`

**Interfaces:**
- Consumes: `productApi.list(): Promise<Product[]>`, `productApi.complete(id, status)`, `StatusBadge`, `Alert`, `EmptyState`, `Skeleton`.
- Produces: an urgency rail derived by `items.filter(product => !['used', 'discarded'].includes(product.status)).slice(0, 3)` and structured inventory/form markup.

- [ ] **Step 1: Write a failing urgency test**

```tsx
test('shows an urgency rail with the most urgent products', async () => {
  render(<MemoryRouter initialEntries={['/products']}><App /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: /использовать первым|use first/i })).toBeInTheDocument()
  expect(screen.getByText('Рыба')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run the test to verify the new hierarchy is not yet exposed**

Run: `npm test -- --run src/app.test.tsx`

Expected: FAIL because the present urgency section has no product rail.

- [ ] **Step 3: Implement the inventory and form hierarchy**

```tsx
const urgent = shown.filter(product => !['used', 'discarded'].includes(product.status)).slice(0, 3)

<section className="urgency-rail" aria-labelledby="urgency-title">
  <h2 id="urgency-title">{t.urgency}</h2>
  <ul>{urgent.map(product => <UrgencyItem key={product.id} product={product} />)}</ul>
</section>
```

Create a named page header with count and add-product CTA. Replace visually flat product list rows with structured cards carrying name, storage, date type/date, semantic badge and lifecycle actions. Group form fields into `form-grid` / `form-actions` classes; preserve label/input associations, Zod errors, submit state, and routes.

- [ ] **Step 4: Run full verification**

Run: `npm test && npm run lint && npm run build`

Expected: all commands exit 0.

- [ ] **Step 5: Inspect responsive renderings**

Run: `npm run dev -- --host 127.0.0.1`

At 320px, 768px, and 1440px, verify no horizontal overflow; fixed mobile navigation does not cover content; all primary actions retain a 44px target; and desktop urgency rail stays beside the inventory.
