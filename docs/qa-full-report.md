# QA Full Report

## Общая информация

- Дата: 28.08.2026
- Ветка: `main`
- Статус: FAILED
- Область: React/Vite UI, доступные backend-контракты и regression-suite.
- Выполненные проверки:
  - `frontend`: `npm run lint` — passed; `npm run test` — 11/11; `npm run build` — passed.
  - `frontend`: `npm run test:e2e` — 8/8 passed на Chromium с viewport 320×740, 768×900 и 1440×960.
  - Backend runtime: `GET http://127.0.0.1:8080/healthz` — недоступен (connection refused). `go` и `docker` отсутствуют в окружении, поэтому `go test ./...`, Docker Compose и реальный API smoke не выполнены.
  - Контракты проверены по исходному коду и серверным unit-тестам: auth, products, recipes и product-drafts зарегистрированы; profile/country и notification-settings HTTP routes отсутствуют.

## Критические баги (блокируют релиз)

### QA-CRIT-01 — Ключевые сценарии страны и e-mail-напоминаний не реализованы

- Компоненты: `frontend/src/app.tsx`, `backend/internal/http/server.go`.
- Шаги воспроизведения:
  1. Войти в приложение и открыть «Настройки».
  2. Попробовать выбрать страну, группу регулятора или порог e-mail-напоминаний.
  3. Открыть регистрацию и проверить доступность выбора страны/языка.
- Фактический результат: экран содержит только blocked-state. В frontend нет формы или API-вызова; в backend не зарегистрированы profile/country или notification-settings endpoints.
- Ожидаемый результат: пользователь может сохранить страну/группу регулятора и e-mail-порог (не менее 60 минут), увидеть success/error feedback; регистрация собирает обязательные country/language данные.
- Влияние: не закрыты UF-01 и UF-08 из Design Requirements; без `country_code` сервер возвращает `research_required`, поэтому основной статусный сценарий нельзя завершить.

### QA-CRIT-02 — Английская локаль показывает смешанный русско-английский интерфейс

- Компоненты: `frontend/src/i18n.ts`, `frontend/src/ui.tsx`, `frontend/src/app.tsx`, `frontend/src/mock-api.ts`.
- Шаги воспроизведения:
  1. Открыть приложение в браузере с `navigator.language=en`.
  2. Войти и открыть список продуктов.
- Фактический результат: заголовки и навигация отображаются на английском, а статусы, типы дат, labels формы и fixture-названия — на русском. Это видно и в утверждённом mobile visual baseline: «Use first» рядом с «Срок истёк» и «Требует внимания».
- Ожидаемый результат: все пользовательские строки, включая статусы, типы дат, labels, ошибки и демонстрационные данные, соответствуют активной локали.
- Влияние: заявленная поддержка ru/en непригодна для англоязычного пользователя.

### QA-CRIT-03 — Bottom sheet не удерживает фокус и не имеет программной подписи

- Компонент: `frontend/src/ui.tsx`, `AddProductSheet`.
- Шаги воспроизведения:
  1. На mobile открыть «Добавить продукт».
  2. Нажимать `Tab` после последней ссылки sheet.
- Фактический результат: `role="dialog"` имеет `aria-modal`, но нет `aria-labelledby` и focus trap; фокус может перейти на фон. Открытие не переводит фокус на заголовок или безопасное действие.
- Ожидаемый результат: sheet подписан через `aria-labelledby`, фокус находится внутри до закрытия, при закрытии возвращается на trigger.
- Влияние: нарушены обязательные требования доступности для keyboard-only пользователей; текущий E2E покрывает только `Escape` и focus return, но не trap.

### QA-CRIT-04 — Контраст статусных и вторичных текстов ниже WCAG AA

- Компонент: `frontend/src/styles/tokens.scss`, `frontend/src/styles/global.scss`.
- Шаги воспроизведения:
  1. Открыть список продуктов и посмотреть 11–12px metadata/status badges.
  2. Рассчитать contrast token-пар для используемого текста и background.
- Фактический результат: вычисленные отношения контраста: muted `#758074` / `#fffdfa` = 4.06:1, active `#4d813e` / `#e6f4d1` = 4.03:1, attention `#a4671d` / `#fff1dc` = 4.16:1, danger `#b7473d` / `#fce7e4` = 4.44:1. Все ниже минимальных 4.5:1 для обычного текста.
- Ожидаемый результат: статусный и вторичный текст выполняет WCAG AA 4.5:1; это явно требовалось Design Requirements.
- Влияние: ухудшается читаемость самых важных предупреждений о сроках.

### QA-CRIT-05 — Реальный интеграционный прогон невозможен в текущем окружении

- Тип: блокер окружения, а не подтверждённый дефект production-кода.
- Шаги воспроизведения:
  1. Выполнить `curl http://127.0.0.1:8080/healthz`.
  2. Выполнить `go test ./...` из `backend` или поднять Docker Compose.
- Фактический результат: API не запущен; `curl` получает connection refused; команды `go` и `docker` не установлены.
- Ожидаемый результат: доступен запущенный backend с PostgreSQL и инструментами Go/Docker, чтобы проверить cookies, CORS, auth, products, drafts, recipes и ошибки по реальной сети.
- Влияние: пройдённые E2E работают только с development fixtures (`VITE_API_URL` пуст); readiness для реальной связки Frontend + Backend не подтверждён.

## Некритические баги (можно после релиза)

### QA-NONCRIT-01 — Успех создания продукта не показывает toast

- Шаги: создать валидный продукт вручную или approve photo draft.
- Факт: после redirect передаётся `state.notice`, но он нигде не отображается; `ToastRegion` не подключён к `AppShell`.
- Ожидается: компактный success-toast и управляемый focus target после обновления списка.

### QA-NONCRIT-02 — В auth-формах не выводятся client-side field errors

- Шаги: отправить login/register с пустым e-mail или паролем короче 8 символов.
- Факт: правила React Hook Form заданы, но `formState.errors` не отображаются и не связаны с control через `aria-invalid`/`aria-describedby`.
- Ожидается: inline error, доступное описание и сохранение введённых данных.

### QA-NONCRIT-03 — Destructive lifecycle actions выполняются без подтверждения

- Шаги: в списке или деталях нажать «Использован» / «Выброшен».
- Факт: запрос уходит сразу.
- Ожидается: confirm dialog с ясным последствием действия, как определено UF-05.

### QA-NONCRIT-04 — Карточки рецептов не содержат обязательного объяснения ингредиента и срока

- Шаги: открыть «Рецепты» при непустом inventory.
- Факт: UI отображает только title и число product IDs.
- Ожидается: «Использует [продукт] — [срок]» и предупреждение для eligible `best_before`; `expired use_by` не должен быть показан как причина.

### QA-NONCRIT-05 — Форма ручного добавления не содержит quantity/unit и не группирует optional details

- Шаги: открыть «Добавить продукт».
- Факт: доступны только name, date type/date, location и group; quantity/unit и раскрываемый блок «Детали хранения» отсутствуют.
- Ожидается: все поля Design Requirements, включая необязательные quantity/unit, с понятной иерархией.

## Замечания по UI/UX

- Адаптив inventory подтверждён скриншотами и E2E: на 320/768/1440 нет горизонтального overflow; mobile sheet закрывается по `Escape`, а `prefers-reduced-motion` отключает анимацию.
- Визуальное направление в коде не совпадает с утверждённым Premium Redesign: используются cream/olive tokens и `Newsreader`/`DM Sans`, тогда как документ определяет `#08111F/#F7F9FC/#20C997` и Onest/IBM Plex Mono; desktop rail также не dark. Это заметное, но не функциональное расхождение с Design Requirements.
- Mobile filters остаются inline chips, хотя спецификация задаёт bottom sheet. Relative date отсутствует: показывается только absolute ISO-date.
- Подтверждённый security regression: HTML-like name рендерится текстом, а не DOM-узлом; E2E passed. Route guard не раскрывает inventory без fixture session; E2E passed.
- В логе прогонов не зафиксировано ошибок приложения, но отдельный runtime-console capture на реальном backend не мог быть выполнен из-за отсутствия запущенного API.

## Рекомендации

1. До релиза реализовать и покрыть контрактами country/profile и notification settings; передавать `country_code`/alert threshold в create/approve flows.
2. Централизовать все строки в i18n, локализовать `StatusBadge`, date-type labels, form labels/errors и fixtures; добавить E2E для `ru` и `en`.
3. Исправить modal accessibility: `aria-labelledby`, initial focus, focus trap и тест `Tab`/`Shift+Tab`; привести контраст status/metadata tokens к >=4.5:1.
4. Запустить `make up` в окружении с Go и Docker, затем выполнить real-API E2E с `VITE_API_URL`, проверить secure cookie, CORS, 401/403/404/409 и network retry.
5. Закрыть UX-долг: success toast, auth field feedback, destructive confirmation, recipe explanation и optional product details. Затем обновить visual baselines после выравнивания с Premium Redesign.

## Повторный цикл — 28.08.2026

### Проверенные исправления

| Предыдущий пункт | Результат повторной проверки | Доказательство |
| --- | --- | --- |
| QA-CRIT-01: profile/settings отсутствовали | Исправлено на уровне кода и контрактов | `GET`/`PUT /v1/profile` и `GET`/`PUT /v1/notification-settings` зарегистрированы в `backend/internal/profile/http.go` и `server.go`; есть account-scoped unit-тесты и миграция `000006_account_profiles_settings.up.sql`; frontend вызывает оба контракта. |
| Country propagation | Исправлено на уровне клиента | `ProductForm` загружает profile и передаёт `countryCode` как в direct create, так и в draft approve; `toBackendProductPayload` сериализует `country_code`. |
| QA-CRIT-02: ru/en consistency | Исправлено в production-коде | Status labels, date labels, fixture names и form copy переведены в `i18n.ts`; source scan не выявил hard-coded Russian user copy вне i18n (кроме названий языков в selector и fallback API message). |
| QA-CRIT-03: add-product sheet a11y | Исправлено для add-product sheet | Add sheet использует `aria-labelledby`, initial focus на close button, Tab/Shift+Tab focus trap и return focus. Lifecycle confirmation проверен отдельно ниже. |
| Destructive action confirmation и success feedback | Исправлено | Lifecycle actions открывают confirmation dialog; `ToastRegion` подключён и обрабатывает navigation notice. |

### Выполненные проверки

- `npm run lint` — passed.
- `npm run test` — passed, 11/11.
- `npm run build` — passed.
- Стандартный `npm run test:e2e` не начал тесты: порт 4173 уже занят внешним Vite-process.
- Изолированный повтор Playwright на 4174 выявил regression в существующем E2E: suite жёстко ищет русские labels `E-mail`/`Пароль`, а текущий Chromium запускает английскую locale и корректно рендерит `Email`/`Password`. Два теста завершились по timeout; остальные не были надёжно выполнены. Это дефект тестовой автоматизации, не доказательство смешанной локализации продукта.
- Real API и Go/Docker smoke не выполнены: API на `127.0.0.1:8080` не запущен, Go toolchain и Docker отсутствуют. Этот факт остаётся environment blocker.

### Новые/неустранённые блокеры

#### QA-REPEAT-CRIT-01 — WCAG AA contrast по-прежнему не выполняется

- `tokens.scss` и `global.scss` не изменены в проблемной части. Контрасты обычного статусного/вторичного текста по-прежнему ниже 4.5:1: muted 4.06:1, active 4.03:1, attention 4.16:1, danger 4.44:1.
- Статус: не исправлено. Блокирует релиз, поскольку предупреждения о сроках являются ключевой информацией.

#### QA-REPEAT-CRIT-02 — Настройки e-mail-напоминаний не применяются к продуктам

- `PUT /v1/notification-settings` сохраняет account-scoped overrides в `account_notification_settings`, но `notification.Service` читает только `products.alert_threshold_minutes` и product-group defaults. Ни frontend, ни product API не передают сохранённую account setting в `alert_threshold_minutes` при create/approve.
- Worker корректно запускается в `backend/cmd/api/main.go`; проблема находится на границе persisted settings → product scheduling, а не в запуске worker.
- Ожидается: сохранённый user threshold применяется к вновь созданным (и согласованно к существующим) продуктам либо scheduler читает account settings.
- Статус: подтверждённый critical backend/integration defect. UI создаёт ложное впечатление работающих напоминаний.

#### QA-REPEAT-CRIT-03 — Confirmation dialog для lifecycle не удерживает фокус

- `LifecycleDialog` имеет `aria-labelledby`, initial focus и `Escape`, но отсутствует обработка `Tab`/`Shift+Tab`; фокус может перейти к фоновому интерфейсу.
- Ожидается: одинаковый focus-trap contract для всех modal dialogs, а не только для add-product sheet.
- Статус: critical a11y defect для destructive action.

#### QA-REPEAT-NONCRIT-01 — E2E suite не детерминирована по locale

- Existing `frontend/e2e/inventory.spec.ts` использует жёсткие русские labels/expectations, но не фиксирует browser locale. При `en` локали login page корректно содержит `Email`/`Password`, поэтому `getByLabel('E-mail')` и `getByLabel('Пароль')` never resolve.
- Ожидается: Playwright projects для `ru-RU` и `en-US`, либо locale-independent locators; добавить обе проверки в CI.

#### QA-REPEAT-NONCRIT-02 — User-facing names notification groups не локализованы

- Settings UI выводит raw protocol values (`refrigerated_perishable`, `fresh_produce` и т. п.), хотя для них есть локализованные copy keys в `i18n.ts`.
- Ожидается: отображать человекочитаемые локализованные названия, оставляя protocol value только внутри API.

## Итог повторного цикла

- Исправлено из исходного отчёта: 5 пунктов (profile/settings contracts, country propagation, локализация в production code, modal a11y, confirmation/toast).
- Не устранено: WCAG AA contrast.
- Новых/неустранённых критических: 3 (contrast, notification setting integration, lifecycle confirmation focus trap).
- Environment blocker: real API/Go/Docker smoke по-прежнему недоступен.
- Статус: FAILED — к релизу не готов до исправления QA-REPEAT-CRIT-01 и QA-REPEAT-CRIT-02, обновления E2E locale coverage и повторного доступного real-API smoke.

## Итог

- Исторический первый цикл: 10 пунктов (5 критических, включая 1 environment blocker; 5 некритических).
- Актуальный repeat-release gate: 3 критических дефекта приложения, 2 некритических замечания и 1 environment blocker.
- Статус: FAILED; детальные актуальные результаты приведены в разделе «Повторный цикл — 28.08.2026».

Релиз не рекомендован до устранения QA-CRIT-01…QA-CRIT-04 и проведения повторного real-API прогона после устранения QA-CRIT-05.
