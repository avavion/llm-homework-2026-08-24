# Реестр Front-End задач MVP

**Источник:** `shared/docs/product-description.md`, `shared/docs/product-context.md`, `frontend/docs/competitor-analysis.md`.

Перед исполнением каждой задачи обязателен статус `technical_review: approved` от [Senior Frontend Developer](../agents/senior-frontend-developer.md). Допускаются **не более двух** кругов взаимного ревью: первичный и, только при исправлениях, один повторный. Новый блокер эскалируется Senior Project Manager, а не создаёт третий цикл.

| ID | Задача | Исполнитель | Зависимости | Timebox | Техническое ревью |
| --- | --- | --- | --- | --- | --- |
| FE-001 | UX-система и адаптивная спецификация | Senior UI/UX Designer | — | 2 дня | approved, 26.08.2026 |
| FE-002 | Каркас клиента: маршрутизация, сессия, i18n и a11y | Senior Frontend Developer | FE-001, BE-003 + country/profile API | 3 дня | blocked: эскалация PM, 26.08.2026 |
| FE-003 | Инвентарь, статусы срока и жизненный цикл | Senior Frontend Developer | FE-001, FE-002, BE-004, BE-005 | 4 дня | implementation complete; awaiting live API smoke, 28.08.2026 |
| FE-004 | Ручное добавление и редактирование продукта | Senior Frontend Developer | FE-001, FE-002, BE-004 | 3 дня | approved, 26.08.2026 |
| FE-005 | Фото-черновик с проверкой и approve/reject | Senior Frontend Developer | FE-001, FE-002, FE-004, BE-006 | 3 дня | approved, 26.08.2026 |
| FE-006 | Рецепты и настройки e-mail-напоминаний | Senior Frontend Developer | FE-001, FE-002, FE-003, BE-005 + согласованный HTTP-контракт | 3 дня | approved, 26.08.2026 |
| FE-007 | Приёмка, white-hat и визуальное QA | Senior QA | FE-002—FE-006 | 3 дня | approved, 26.08.2026 |

## Последовательность

`FE-001 → FE-002 → (FE-003 + FE-004) → FE-005 → FE-006 → FE-007`.

`FE-003` и `FE-004` разрешено выполнять параллельно только после готовности `FE-002` и соответствующих backend-контрактов.

## Эскалация PM: country/profile API

`shared/docs/product-description.md` требует выбора страны и хранения группы регулятора. В текущем плане BE-003 регистрационные DTO и хранилище страны не опубликованы, а BE-004 не содержит profile endpoint. Нужна отдельная согласованная backend-задача или расширение BE-003 с account-scoped полями `country_code`, `regulator_group`, чтением/изменением профиля, DTO и кодами ошибок. До её одобрения FE-002 и все зависимые задачи не начинают production-интеграцию.
