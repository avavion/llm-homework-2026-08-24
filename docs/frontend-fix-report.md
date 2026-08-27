# Отчёт Frontend: исправление Critical замечаний QA

- **Дата:** 27.08.2026
- **Исполнитель:** Senior Frontend Developer
- **Основание:** [QA-отчёт](qa-report.md)
- **Статус:** готово к повторной QA-проверке

## Исправлено

| QA ID | Исправление | Проверяемый результат |
| --- | --- | --- |
| C-01 | Рабочий UI использует `api.ts`: `GET /v1/products`, `GET /v1/products/:id`, `POST /v1/products` и lifecycle-запросы. | При `VITE_API_URL` используются HTTP/cookie-запросы; без переменной в development подключаются только явно названные fixtures. В production без переменной применяется same-origin HTTP, не fixtures. |
| C-02 | Добавлены `/login`, `/register`, восстановление сессии, logout и route guard. | Неавторизованный пользователь перенаправляется на вход до загрузки приватных данных; cookie передаётся через `credentials: include`. Country/profile намеренно не реализованы: контракта нет. |
| C-03 | Фото распознаётся в отдельный server/fixture draft, затем открывается `/product-drafts/:id`. | Продукт создаёт исключительно `POST .../:id/approve`; reject не создаёт продукт. Есть regression-test этого инварианта. |
| C-04 | Восстановлены `/products/:id`, `/recipes` и `/settings`. | Detail и recipes имеют loading/error/retry/empty-состояния. Settings честно показывает blocked-state для отсутствующих notification/profile API. |
| C-05 | Добавлены React Query состояния и локализованные сообщения об ошибке. | List/detail/recipes/draft имеют retry; формы, lifecycle, recognize, approve/reject блокируют повторный клик и показывают inline error. |
| C-06 | `AppShell` применяется только в защищённом layout. | Фото-загрузка и review используют один защищённый shell и один landmark `main`; review не вкладывает форму в новый shell. |

## Верификация

- `cd frontend && npm run lint` — PASS
- `cd frontend && npm run test` — PASS (4 tests)
- `cd frontend && npm run build` — PASS
- `git diff --check` — PASS

## Открытые ограничения

- Playwright acceptance suite (N-01) не добавлялся в этот пакет Critical fixes; его должен выполнить отдельный этап FE-007.
- API country/profile, notification settings и PATCH/DELETE products отсутствуют на backend и не эмулируются как production-возможности.
- Реальный backend smoke требует запущенный сервис и корректный `VITE_API_URL`; в рамках этой сессии внешний стенд не предоставлен.
