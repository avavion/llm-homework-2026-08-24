# QA-отчёт: Front-End MVP

- **Дата:** 27.08.2026
- **Исполнитель:** Senior QA — Frontend
- **Вердикт:** `changes_required`
- **Область:** готовый Front-End MVP; проверка против [Design Requirements](../frontend/docs/design-requirements.md), [Premium Redesign](design-requirements-redesign.md) и задач FE-002—FE-007.

## Выполненные проверки

| Проверка | Результат | Evidence |
| --- | --- | --- |
| Production build | PASS | `cd frontend && npm run build` — Vite собрал 171 модуль без ошибок. |
| Unit tests | PASS, недостаточно для приёмки | `npm run test` — 3/3 теста прошли. Покрыты только shell продуктов, навигация и 404. |
| TypeScript / lint script | PASS | `npm run lint` — `tsc -b --noEmit` завершён с кодом 0. |
| Локальный smoke-стенд | PASS (HTTP) | Vite локально ответил `200` для `/products`; это не заменяет браузерную проверку рендера. |
| Playwright visual / adaptive / console | BLOCKED | В `frontend/package.json` нет `playwright`/`@playwright/test`; в проекте нет Playwright config, spec, baseline или trace. |
| Viewport 320 / 768 / 1440 | NOT EXECUTED | Браузерные тесты и screenshot-baseline отсутствуют; требования FE-007 не выполнимы воспроизводимо. |
| Реальный backend smoke | NOT EXECUTED | UI не использует `src/api.ts`, а доступный локальный API стенд не был предоставлен QA. |

## Critical / блокирующие баги

### C-01 — рабочий UI использует fake API, а не готовые API-клиенты

- **Severity:** Critical
- **Шаги воспроизведения:**
  1. Открыть `frontend/src/app.tsx`.
  2. Найти импорт и вызовы `productApi`/`draftApi`.
  3. Запустить приложение с `VITE_API_URL` на работающем backend.
- **Ожидаемый результат:** список и создание продукта используют `GET /v1/products`, `GET /v1/products/:id` и `POST /v1/products`; UI показывает loading, успех и локализованную ошибку API.
- **Фактический результат:** `app.tsx` импортирует `./mock-api` и вызывает только in-memory `productApi`/`draftApi`. `src/api.ts` существует, но не импортируется UI. Значение `VITE_API_URL` не влияет на пользовательский поток.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:5), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:15), [api.ts](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/api.ts:5).
- **Риск:** приложение демонстрирует данные, не принадлежащие сессии пользователя; созданные данные теряются после перезагрузки. FE-003/FE-004 нельзя принять как интеграционные.

### C-02 — аутентификация и защита маршрутов отсутствуют в UI

- **Severity:** Critical
- **Шаги воспроизведения:**
  1. Открыть `/products` в новом браузерном профиле без session cookie.
  2. Открыть `/login` или `/register`.
- **Ожидаемый результат:** неавторизованный пользователь перенаправляется на доступную форму входа; login/register/logout/session работают через cookie и защищённые маршруты не раскрывают данные.
- **Фактический результат:** `/products` всегда доступен; маршруты `/login` и `/register` попадают в catch-all 404. Методы auth в `api.ts` не подключены к экрану или route guard.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:23), [api.ts](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/api.ts:5).
- **Риск:** не выполнены FE-002 и требования DR UF-01; нет проверяемой boundary между public и private UI.

### C-03 — фото-черновик нарушает обязательный approve/reject flow

- **Severity:** Critical (данные / безопасностная логика)
- **Шаги воспроизведения:**
  1. Открыть `/products/new/photo` и выбрать изображение.
  2. После появления «Черновик — проверьте данные…» сохранить форму.
  3. Проверить вызов и состояние mock API.
- **Ожидаемый результат:** результат recognize заполняет редактируемые поля; продукт создаётся исключительно `POST /v1/product-drafts/:id/approve` после явного подтверждения. Reject не создаёт продукт.
- **Фактический результат:** результат `draftApi.recognize` отбрасывается. Вложенная `Form` вызывает `productApi.create`, не `draftApi.approve`; ID черновика не существует в UI. Кнопка reject расположена после формы, поэтому в этом flow пользователь может создать продукт без approve.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:21), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:22), [mock-api.ts](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/mock-api.ts:5).
- **Риск:** прямое нарушение FE-005 и DR UF-04: не проверенные данные распознавания могут стать продуктом, а reject-flow не гарантирует отсутствие созданной записи.

### C-04 — обязательные экраны и маршруты отсутствуют

- **Severity:** Critical
- **Шаги воспроизведения:**
  1. В навигации выбрать «Рецепты» или «Настройки».
  2. Открыть `/products/milk`.
- **Ожидаемый результат:** доступны states экранов recipes, settings и product details либо согласованный blocked-state без 404.
- **Фактический результат:** определены только `/products`, `/products/new`, `/products/new/photo`; все остальные маршруты рендерят 404, хотя ссылки на `/recipes` и `/settings` показаны в primary navigation.
- **Evidence:** [ui.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/ui.tsx:10), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:23).
- **Риск:** пользователь попадает в dead end; не выполнены DR раздел 3 и FE-006 mock UI.

### C-05 — нет error handling для запросов и мутаций; loading может остаться навсегда

- **Severity:** Critical
- **Шаги воспроизведения:**
  1. Смоделировать rejection `productApi.list`, `productApi.create`, `complete` или `draftApi.recognize`.
  2. Открыть список / отправить форму / завершить продукт / загрузить фото.
- **Ожидаемый результат:** plain-language alert и retry; форма не теряет значения; submit/action блокируется до завершения; ошибки сети не остаются в консоли как unhandled rejection.
- **Фактический результат:** promise chains не имеют `catch` или error-state. При ошибке list `items` остаётся `null` и skeleton не заканчивается. Submit, lifecycle и recognize не показывают ошибку; Photo не имеет submitting/progress state.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:15), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:18), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:21), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:22).
- **Риск:** не выполнен DR раздел 8; в реальном браузере это может приводить к console error и недоступному интерфейсу.

### C-06 — HTML-семантика фото-flow создаёт вложенные AppShell и `<main>`

- **Severity:** Critical (a11y / документная структура)
- **Шаги воспроизведения:**
  1. Открыть `/products/new/photo`.
  2. Выбрать файл и перейти в экран проверки черновика.
  3. Inspect accessibility tree / DOM.
- **Ожидаемый результат:** один AppShell, один landmark `<main>`, один `<h1>` страницы; draft form находится внутри этого main без повторного shell.
- **Фактический результат:** `Photo` уже обёрнут `Page`/`AppShell`, а `Form` внутри него снова обёрнут `Page`/`AppShell`; итог содержит вложенные `<main id="main-content">` и дублирующиеся skip links.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:9), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:21), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:22), [ui.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/ui.tsx:10).
- **Риск:** screen reader landmarks и переход skip-link неоднозначны; нарушается DR §7.

## Non-critical баги и риски

### N-01 — Playwright QA-контур отсутствует

- **Severity:** High / release-blocking acceptance gap
- **Шаги воспроизведения:** проверить `frontend/package.json` и дерево `frontend/` на Playwright config/specs.
- **Ожидаемый результат:** Chromium Playwright с 320/768/1440, visual snapshots, traces и командой запуска по FE-007.
- **Фактический результат:** нет зависимости, config, spec, snapshot-baseline и `test-results`.
- **Evidence:** [package.json](/Users/vilka/Works/llm-homework-2026-08-24/frontend/package.json:1); поиск не нашёл Playwright-файлов.

### N-02 — не выполнена требуемая адаптивная компоновка desktop rail

- **Severity:** Medium
- **Шаги воспроизведения:** открыть `/products` при ширине ≥1024px после появления browser-suite.
- **Ожидаемый результат:** согласно Redesign — dark rail 248px, main 8 columns и urgency rail 4 columns; согласно исходному DR — sidebar 240px.
- **Фактический результат:** CSS задаёт dark rail и nav шириной 132px. Это почти вдвое меньше спецификации, нарушает согласованную информационную иерархию.
- **Evidence:** [global.scss](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/styles/global.scss:1), [design-requirements-redesign.md](design-requirements-redesign.md).

### N-03 — keyboard focus и связанные с ошибками ARIA не реализованы полностью

- **Severity:** Medium
- **Шаги воспроизведения:**
  1. Перейти по форме клавишей Tab и отправить пустую форму.
  2. Проверить CSS `:focus-visible` и атрибуты поля с ошибкой.
- **Ожидаемый результат:** контрастный focus ring `--focus`; поле имеет `aria-invalid` и `aria-describedby` на текст ошибки; фокус после server error идёт на summary/первое поле.
- **Фактический результат:** в stylesheet нет правил `:focus`/`:focus-visible`; только `aria-invalid` задан программно, но ошибка не имеет id и не связана `aria-describedby`. Server error-flow отсутствует.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:21), [global.scss](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/styles/global.scss:1).

### N-04 — формы не соответствуют полной спецификации валидации и feedback

- **Severity:** Medium
- **Шаги воспроизведения:** открыть `/products/new`; оставить дату пустой или выполнить save; проверить поля и результат.
- **Ожидаемый результат:** форма разделяет обязательные/опциональные данные, подставляет/проверяет все разрешённые API поля, после успеха возвращает к обновлённому списку с toast и focus target.
- **Фактический результат:** только name/date проверяются на непустое значение; date не проверяется как дата, quantity/unit/country/alert threshold отсутствуют. Успех остаётся на форме в виде информационного alert, перехода, toast и обновления списка нет.
- **Evidence:** [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:20), [app.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.tsx:21).

### N-05 — visual states не покрыты целиком и не соответствуют Calm Ledger

- **Severity:** Medium
- **Шаги воспроизведения:** проверить стили `StatusBadge`, button, card и toast states.
- **Ожидаемый результат:** все компоненты имеют default/hover/active/disabled/loading/error; mint CTA, soft elevation, status rail и toast определены токенами.
- **Фактический результат:** нет CSS для hover/active/focus/loading/toast; нет классов `badge--used` и `badge--discarded`; локальные hex используются вместо токенов (`#fff`, `#aab7c7`, `#ffffff1f`, `#91a2b8`). Премиальное направление реализовано частично и без интерактивной обратной связи.
- **Evidence:** [global.scss](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/styles/global.scss:1), [ui.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/ui.tsx:7).

### N-06 — тесты не покрывают критичные пользовательские и негативные сценарии

- **Severity:** Medium
- **Шаги воспроизведения:** прочитать `frontend/src/app.test.tsx`.
- **Ожидаемый результат:** unit/integration coverage auth/guard, list loading-error-empty, POST validation, product lifecycle, draft approve/reject/double click, recipes/settings fixture и a11y keyboard smoke.
- **Фактический результат:** в одном файле только 3 smoke assertions; нет user interaction, API error, security или responsive coverage.
- **Evidence:** [app.test.tsx](/Users/vilka/Works/llm-homework-2026-08-24/frontend/src/app.test.tsx:1).

## Соответствие ключевым требованиям

| Категория | Статус | Вывод |
| --- | --- | --- |
| Inventory list / filter | Частично | На fixture данные фильтруются и сортируются по дате, но real API и error state не подключены. |
| Manual add | Частично | Есть RHF + Zod для двух обязательных полей, но только mock create и неполный feedback. |
| Photo draft | Fail | approve/reject safety invariant не соблюдён. |
| Auth / protected routes | Fail | API client есть, пользовательского flow нет. |
| Recipes / settings | Fail | Ссылки ведут в 404. |
| Loading / error / empty | Fail | Loading и empty частичны; error/retry отсутствуют. |
| Responsive 320/768/1440 | Not verified | Нет Playwright/browser artifacts; desktop CSS расходится с design spec. |
| a11y | Fail | Имеются semantic fundamentals, но нарушена структура draft-flow, focus и error association не сделаны. |
| Console errors | Not verified / risk | Browser console нельзя подтвердить без browser-suite; unhandled promise rejections вероятны при API errors. |
| Premium visual direction | Частично | Базовая палитра/rail/status accents есть, однако tokens и интерактивные states применены неполно. |

## Рекомендованный порядок исправлений

1. Исправить C-03: отдельный draft-state, редактирование result, только `approve`, безопасный reject и tests.
2. Исправить C-01 и C-05: внедрить real API adapter в UI, React Query loading/error/retry и локализованные ошибки.
3. Исправить C-02 и C-04: login/register/session/guard, product detail, recipes/settings либо честный blocked state.
4. Исправить C-06 и N-03: один shell/main, focus management, `:focus-visible`, `aria-describedby`.
5. Добавить Playwright acceptance suite и baseline (N-01), затем провести один разрешённый повторный QA-круг на 320/768/1440.

## Ограничения проверки

- QA не изменял код приложения и не запускал внешние системы.
- Локальный Vite server был доступен для HTTP smoke. Полный браузерный/Playwright прогон не выполнен, потому что tooling и tests отсутствуют в репозитории. Это не PASS для visual, responsive, keyboard или console acceptance.
