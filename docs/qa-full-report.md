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

## Итог

- Всего багов: 10
- Критических: 5 (из них 1 блокер окружения, не дефект кода)
- Некритических: 5
- Статус: FAILED

Релиз не рекомендован до устранения QA-CRIT-01…QA-CRIT-04 и проведения повторного real-API прогона после устранения QA-CRIT-05.
