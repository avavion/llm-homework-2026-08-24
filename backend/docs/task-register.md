# Реестр backend-задач MVP

Источник: `shared/docs/product-description.md`. Перед разработкой каждой задачи обязателен статус `technical_review: approved` от Senior Backend Developer. Всего допускается два круга: исходная проверка и одна повторная после исправлений.

| ID | Задача | Исполнитель | Зависимости | Timebox | План |
| --- | --- | --- | --- | --- | --- |
| BE-000 | DevOps-среда Go + PostgreSQL | System Admin / DevOps | — | 2 дня | [план](superpowers/plans/2026-08-26-be-000-devops-environment.md) |
| BE-001 | Каркас Go API и PostgreSQL | Backend Developer | — | 2 дня | [план](superpowers/plans/2026-08-26-be-001-api-postgres-foundation.md) |
| BE-002 | Реестр регуляторных правил | PM + исследователь | — | 3 дня | [план](superpowers/plans/2026-08-26-be-002-regulatory-rules-registry.md) |
| BE-003 | Регистрация и сессии | Backend Developer | BE-001 | 3 дня | [план](superpowers/plans/2026-08-26-be-003-authentication-and-sessions.md) |
| BE-004 | Продукты и жизненный цикл | Backend Developer | BE-001, BE-003 | 3 дня | [план](superpowers/plans/2026-08-26-be-004-products-and-lifecycle.md) |
| BE-005 | Сроки, e-mail, рецепты | Backend Developer | BE-002, BE-004 | 4 дня | [план](superpowers/plans/2026-08-26-be-005-expiry-notifications-recipes.md) |
| BE-006 | OCR/LLM-черновики | Backend Developer | BE-003, BE-004 | 4 дня | [план](superpowers/plans/2026-08-26-be-006-photo-drafts-approval.md) |
| BE-007 | QA и white-hat приёмка | QA | BE-001—BE-006 | 3 дня | [план](superpowers/plans/2026-08-26-be-007-backend-acceptance-whitehat.md) |

## Процесс

PM создаёт `draft` → Backend Developer даёт `approved` или замечания → исполнитель прикладывает тесты → QA даёт `passed` или `changes_required` → один пакет исправлений и общий повторный контроль. Новый блокер эскалируется, а не запускает третий круг.
