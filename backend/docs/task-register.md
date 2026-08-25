# Реестр backend-задач MVP

Источник: `shared/docs/product-description.md`. Перед разработкой каждой задачи обязателен статус `technical_review: approved` от Senior Backend Developer. Всего допускается два круга: исходная проверка и одна повторная после исправлений.

| ID | Задача | Исполнитель | Зависимости | Timebox |
| --- | --- | --- | --- | --- |
| BE-000 | DevOps-среда Go + PostgreSQL | System Admin / DevOps | — | 2 дня |
| BE-001 | Каркас Go API и PostgreSQL | Backend Developer | — | 2 дня |
| BE-002 | Реестр регуляторных правил | PM + исследователь | — | 3 дня |
| BE-003 | Регистрация и сессии | Backend Developer | BE-001 | 3 дня |
| BE-004 | Продукты и жизненный цикл | Backend Developer | BE-001, BE-003 | 3 дня |
| BE-005 | Сроки, e-mail, рецепты | Backend Developer | BE-002, BE-004 | 4 дня |
| BE-006 | OCR/LLM-черновики | Backend Developer | BE-003, BE-004 | 4 дня |
| BE-007 | QA и white-hat приёмка | QA | BE-001—BE-006 | 3 дня |

## Процесс

PM создаёт `draft` → Backend Developer даёт `approved` или замечания → исполнитель прикладывает тесты → QA даёт `passed` или `changes_required` → один пакет исправлений и общий повторный контроль. Новый блокер эскалируется, а не запускает третий круг.
