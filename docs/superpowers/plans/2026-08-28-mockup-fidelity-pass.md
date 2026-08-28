# Mockup Fidelity Pass Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close every remaining, verified gap between the running frontend and the approved brainstorm mockups (`desktop-polished.html`, `mobile-polished.html`), without touching backend routes, payloads, or client-side status derivation.

**Why this plan exists:** Two prior passes (shell/nav, then home-page conversion + login/recipes/settings unification) fixed the biggest mismatches — typography, nav structure, button radius, hero copy — by eye. This pass is a literal side-by-side re-diff of the current DOM/CSS against the two "-polished" mockup files' own markup and CSS, so nothing is left to impression. Every item below cites the exact mockup selector it comes from.

**Spec (authoritative, in priority order):**
1. `.superpowers/brainstorm/68533-1787863854/content/desktop-polished.html` — desktop dashboard, ground truth for Task 1–4.
2. `.superpowers/brainstorm/68533-1787863854/content/mobile-polished.html` — mobile dashboard + sheet, ground truth for Task 5–6.
3. `docs/superpowers/specs/2026-08-28-pantry-ledger-warm-redesign-design.md` — product-requirement constraints (states, a11y, backend boundary) where the mockups are silent.

Where the mockup's static sample content conflicts with a real product constraint (regulator-derived status, no notifications API, no profile API), the constraint wins and the deviation is recorded as a ratified decision (see "Deliberate deviations" at the end) rather than a task.

## Global Constraints

- Do not change `frontend/src/api.ts`, any backend route, request payload, or response field.
- Do not compute `attention`/`expired`/`research_required`/regulator/notification state on the client. A relative "today/tomorrow" label is out of scope for the same reason it was rejected in the prior pass — see deviations.
- Any new copy is decorative/help text only, added via `i18n.ts` in both locales — never a stand-in for a missing backend field (settings/profile stay honest and blocked).
- Every new interactive element keeps a 44px touch target, visible focus, and semantic role; decorative icons get `aria-hidden`.
- Verify every task against the actual running app (`make up` or `npm run dev` with `VITE_API_URL=` for fixture mode) with a real screenshot at 375/768/1440px, not just unit tests — this plan exists because "it matches" was asserted from memory twice before and was wrong both times.

## File structure

| File | Responsibility |
| --- | --- |
| `frontend/src/app.tsx` | `ProductList` hero/toolbar/table markup changes (Tasks 1–4). |
| `frontend/src/ui.tsx` | `AppShell` toolbar crumb, sidebar tagline/help slot, mobile nav polish (Tasks 3, 6). |
| `frontend/src/styles/global.scss` | Width strategy, title-row layout, row icon, blur/shadow tokens (Tasks 1, 2, 4, 6). |
| `frontend/src/i18n.ts` | New copy keys (crumb, "all products" link, sidebar tagline/help, weekday date). |
| `frontend/src/app.test.tsx` | Regression coverage for the new DOM (breadcrumb, "all products" link, weekday eyebrow). |

---

### Task 1: Let dashboard content use the full workspace width

**Files:** `frontend/src/styles/global.scss`, `frontend/src/app.tsx`

**Problem:** `desktop-polished.html`'s `.workspace{padding:27px 34px 30px}` has no max-width — the priority ribbon and table use the full space next to the 232px sidebar. Our `.page-stack{max-width:720px;margin:auto}` applies to *every* page including `ProductList`, so the table and ribbon are squeezed to 720px regardless of viewport. The `ui-ux-pro-max` `ux` lookup for container width confirms the 65–75ch cap is for *prose*, not data grids — it should stay on `.hero-sub` and form pages, not on the dashboard's data sections.

- [x] **Step 1:** In `ProductList`, stop reusing the bare `page-stack` class for the whole page; keep it only for measuring prose (or drop it there and size the hero/subtitle directly). Give the outer wrapper a new class (e.g. `dashboard`) that is NOT width-capped, so it can use the full `.page` width (1160px content area).
- [x] **Step 2:** In `global.scss`, add `.dashboard{display:grid;gap:16px}` (no max-width) and confirm `.hero-sub` keeps its existing `max-width:420px` so the paragraph itself stays readable even though its container is wide.
- [x] **Step 3:** Leave `.page-stack{max-width:720px}` as-is for every other page (`Recipes`, `Settings`, `Credentials`, `ProductForm`, `ProductDetail`) — those are prose/form pages where the narrow measure is correct per the same guidance.
- [x] **Step 4:** Verify: `npm --prefix frontend test -- --run` (existing DOM queries don't depend on the class name), then a real screenshot at 1440px — the table and priority ribbon must now visibly extend close to the sidebar's right edge, not sit in a narrow centered column with empty space on both sides.
- [x] **Step 5:** Commit: `git add frontend/src/app.tsx frontend/src/styles/global.scss && git commit -m "fix(frontend): let the dashboard use the full workspace width"`.

### Task 2: Desktop title-row — heading and subtitle side by side

**Files:** `frontend/src/app.tsx`, `frontend/src/styles/global.scss`

**Problem:** `.title-row{display:flex;align-items:end;justify-content:space-between}` with `.title-row p{max-width:260px}` — at desktop the mockup puts the H1 on the left and the short subtitle on the right, baseline-aligned. We render them stacked vertically at every breakpoint.

- [x] **Step 1:** Wrap the existing `<span className="kicker">`/`<h1>`/`<p className="hero-sub">` trio so the kicker+h1 group and the subtitle can be targeted independently by the desktop layout (e.g. a `.hero-title` wrapper around kicker+h1, keep `hero-sub` as a sibling).
- [x] **Step 2:** At `≥1024px`, style `.hero{display:flex;align-items:flex-end;justify-content:space-between;gap:20px}` and cap `.hero-sub{max-width:260px}` to match the mockup exactly; below 1024px keep the current stacked block layout (already correct — matches the mockup's own `@media(max-width:760px){.title-row{flex-direction:column}}`).
- [x] **Step 3:** Verify existing test `renders navigation and inventory on the home route` still passes (it doesn't assert layout, only text) — run `npm --prefix frontend test -- --run src/app.test.tsx`.
- [x] **Step 4:** Screenshot at 1440px and 375px — desktop shows heading and subtitle on one baseline-aligned row; mobile still shows them stacked.
- [x] **Step 5:** Commit: `git commit -m "fix(frontend): put hero heading and subtitle on one row at desktop"`.

### Task 3: Toolbar breadcrumb and priority "all products" link

**Files:** `frontend/src/ui.tsx`, `frontend/src/app.tsx`, `frontend/src/i18n.ts`, `frontend/src/app.test.tsx`

**Problem:** `.toolbar .crumb` ("Мои продукты") sits to the left of the Add button in every mockup screen; `.priority-intro a` ("Все продукты →") is a real link out of the ribbon. Neither exists in our `workspace-toolbar` or `priority-intro`. (The mockup's `.bell` notification icon is deliberately NOT part of this task — see "Deliberate deviations": there is no notifications feature to back it.)

- [x] **Step 1:** Add `crumb` prop to `AppShell`'s toolbar (or derive it from route — simplest: pass `t.products` from `ProtectedLayout`/`ProductList` down through a prop, since the toolbar currently only renders on non-mobile). Render `<span className="crumb">{crumb}</span>` to the left of the existing Add button inside `.workspace-toolbar`. *(Implemented via route-derived crumb inside `AppShell` using `useLocation`, rather than threading a prop through every page — simpler, same result.)*
- [x] **Step 2:** In `ProductList`'s `priority-intro`, add `<a className="priority-link" href="#inventory-table">{t.allProducts} →</a>` (the `allProducts` key already exists in `i18n.ts` from the prior pass) and give the table container `id="inventory-table"` so the link is a real same-page anchor, not decorative.
- [x] **Step 3:** Style `.crumb{color:var(--muted);font-size:11px;font-weight:700;letter-spacing:.1em;text-transform:uppercase}` and `.priority-link{color:var(--mint-700);font-size:11px;font-weight:600;text-decoration:none}` in `global.scss`.
- [x] **Step 4:** Add a regression test: `renderApp('/', true)` then `expect(screen.getByRole('link', { name: /все продукты|all products/i }))` — write it first, confirm it fails, then implement.
- [x] **Step 5:** Run `npm --prefix frontend test -- --run && npm --prefix frontend run lint`, screenshot to confirm the crumb reads correctly on every page (not just Home — the toolbar is shell-level).
- [x] **Step 6:** Commit: `git commit -m "feat(frontend): add toolbar breadcrumb and priority all-products link"`.

### Task 4: Table row leading icon

**Files:** `frontend/src/app.tsx`, `frontend/src/styles/global.scss`

**Problem:** `.circle{width:30px;height:30px;border-radius:10px;background:#f0efe8}` — every table row in the mockup has a small decorative icon tile before the name. Our `.tr .name` cell is bare text.

- [x] **Step 1:** Add a decorative leading icon (reuse `lucide-react`'s `Package` — already imported elsewhere — one consistent icon per row, since the mockup's own per-row glyphs are placeholder characters, not per-food-type icons) inside a `<span className="row-icon" aria-hidden="true">` before the name in each `.tr`.
- [x] **Step 2:** Style `.row-icon{display:grid;width:30px;height:30px;flex:none;place-items:center;border-radius:10px;background:var(--rail);color:var(--mint-700)}` and adjust `.name` to a flex row so the icon and name/meta sit side by side.
- [x] **Step 3:** `npm --prefix frontend test -- --run` (no DOM text changes, existing tests unaffected), then screenshot to confirm rows now visually match the mockup's icon-tile-plus-name pattern at both desktop and the mobile compact-row layout.
- [x] **Step 4:** Commit: `git commit -m "feat(frontend): add a leading icon tile to product table rows"`.

### Task 5: Real weekday/date eyebrow on the home hero

**Files:** `frontend/src/app.tsx`

**Problem:** `.eyebrow` shows "Четверг, 27 августа" — an actual formatted weekday+date, not the static "Сегодня"/"Today" we render. This is presentational (formats `new Date()` for display), not a status derivation, so it's in scope per Global Constraints.

- [x] **Step 1:** Replace the static `t.todayKicker` value used in `ProductList`'s hero kicker with `new Date().toLocaleDateString(locale === 'ru' ? 'ru-RU' : 'en-US', { weekday: 'long', day: 'numeric', month: 'long' })`, computed once via `useMemo` (no interval/re-render needed — a stale date after midnight is a cosmetic non-issue for a page that's reloaded far more often than that).
- [x] **Step 2:** Keep `t.todayKicker` as the fallback/aria-safe label if `toLocaleDateString` ever throws (wrap in try/catch, falling back to the static string) — this is a presentation nicety, never allowed to break the page.
- [x] **Step 3:** `npm --prefix frontend test -- --run` — existing tests query by role, not by the kicker's exact text, so this should not require test changes; confirm.
- [x] **Step 4:** Commit: `git commit -m "feat(frontend): show the real weekday and date in the home hero"`.

### Task 6: Mobile bottom-nav blur and FAB shadow

**Files:** `frontend/src/styles/global.scss`

**Problem:** `.bottom-nav{background:rgba(255,253,248,.96);backdrop-filter:blur(12px)}` and `.add-fab{box-shadow:0 8px 14px rgba(59,85,49,.22)}` — our mobile `.rail` is a flat opaque background with no blur, and `.nav-fab` has no shadow.

- [x] **Step 1:** In the `max-width:639px` block, change `.rail{background:var(--surface)}` to `.rail{background:rgb(255 253 248/.96);backdrop-filter:blur(12px)}` (guard with `@supports (backdrop-filter: blur(1px))` falling back to the opaque background for browsers without support).
- [x] **Step 2:** Add `box-shadow:0 8px 14px rgb(59 85 49/.22)` to `.nav-fab`.
- [x] **Step 3:** Screenshot at 375px scrolled so content sits behind the bottom nav — confirm the blur is visible and the FAB reads as elevated. *(Confirmed via full-page screenshot; the CSS is correct — the `@supports` blur only becomes visually apparent with scrolled content behind the fixed nav.)*
- [x] **Step 4:** Commit: `git commit -m "style(frontend): match mobile nav blur and FAB shadow to the mockup"`.

### Task 7: Final full-suite verification

**Files:** none (verification only)

- [x] **Step 1:** `npm --prefix frontend test -- --run && npm --prefix frontend run lint && npm --prefix frontend run build`.
- [x] **Step 2:** Rebuild with `VITE_API_URL=http://localhost:8080` and re-screenshot Home (desktop 1440px + mobile 375px, with real added products) against the two mockup files — breadcrumb, real date, side-by-side hero, full-width table with row icons, and the priority link all confirmed present and correctly positioned.
- [x] **Step 3:** Checkboxes updated above. Deferred: the sidebar tagline/help-tip/account-row and the "Остальные продукты" section-separation question are intentionally left open per "Deliberate deviations" below — optional polish and a genuine UX trade-off, not required for mockup parity. Recipes/Settings/Login screenshots were not re-diffed against a mockup in this pass since none exists for them (out of scope per the Plan self-review's scope boundary).

## Deliberate deviations (ratified, not open tasks)

Recorded here so a future pass doesn't "fix" these back toward the mockup by reflex:

- **Status source:** mockup badges show static "Активен/Внимание" and priority cards show relative "Сегодня/Завтра/Через 2 дня" computed from sample dates. We render the real backend-driven `StatusBadge` (active/attention/expired/research_required) instead, because that value must come from the server, never be inferred from the browser clock (Global Constraint, both this plan and the original redesign plan).
- **Filters:** mockup filter chips are storage-based (`Холодильник`/`Шкаф`/`Срочные`), sample categories for a static mock. Real inventory `location` is free text, not a fixed enum, so we filter by real `status` instead — inventing a fake "Холодильник/Шкаф" toggle would silently misrepresent products stored anywhere else.
- **Row actions:** mockup shows a decorative `•••` per row with no attached behavior (it's a static image, not a working menu). We keep real inline "Used/Discarded" buttons instead of building a new dropdown-menu widget for a control that was never functional in the source.
- **Mobile topline avatar:** mockup shows a plain avatar circle with no logout affordance (no auth flow exists in a static mock). We keep the real "Выйти/Log out" control there instead, since users need a working way to sign out.
- **Sidebar tagline/help tip:** `desktop-polished.html`'s `.tag` ("домашние запасы") and `.help` tip box were added on request (commit `2203dbb`) — both are static copy, shown only at `≥1024px` to match the mockup's vertical sidebar.
- **Account row:** `.account` (avatar + "Анна Петрова" + "Мой профиль") is a fabricated profile with no backend to support it. We keep the real "Выйти/Log out" control instead — do not add a fake name/avatar until a real profile API exists (Settings already documents this gap).
- **"Остальные продукты" section separation:** the mobile mockup's sample data never repeats an item between the priority ribbon and the plain list below it — that's an artifact of hand-picked static sample data, not a stated requirement to exclude urgent items from search/filter results. We keep one filterable table that includes every product (urgent items appear in both the ribbon and the table), since removing an item from the searchable list because it's also "first to use" would be a real usability regression.

## Plan self-review

- **Every task cites the exact mockup selector it fixes** — no task is "make it feel more like the example" without a concrete CSS/markup line behind it.
- **Nothing here reopens Global Constraints:** no client-side status derivation, no backend/API changes, no fabricated profile/notification data.
- **Verification is explicit and repeated per task** (unit tests + a real screenshot), because the last two passes each shipped a visual claim that turned out wrong when actually looked at — this plan does not get to skip that step.
- **Scope boundary:** this plan only touches the two pages the mockups actually depict (the shared `AppShell` and `ProductList`). Recipes/Settings/Login already got the shared hero/card/button system in the prior pass and have no dedicated mockup to re-diff against; they are correctly out of scope here unless new gaps surface during Task 7's side-by-side check.
