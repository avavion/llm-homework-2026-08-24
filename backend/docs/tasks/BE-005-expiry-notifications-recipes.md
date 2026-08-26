# BE-005 — Статусы сроков, e-mail и рецепты

**Исполнитель:** Senior Backend Developer · **Статус:** draft · **Зависимости:** BE-002, BE-004 · **Timebox:** 4 дня

**План:** [2026-08-26-be-005-expiry-notifications-recipes.md](../superpowers/plans/2026-08-26-be-005-expiry-notifications-recipes.md)

## SMART-цель

Рассчитывать статусы по подтверждённой конфигурации, отправлять идемпотентные e-mail от 60 минут и исключать просроченный `use_by` из собственных рецептов.

## Объём и приёмка

- Rule evaluator, scheduler, delivery log, development mail sender и детерминированные рецепты по группам.
- `use_by` после правила — `expired` и без рецептов; `best_before` — `attention`, предупреждение и рецепты доступны.
- Порог <60 минут отклонён; повторный запуск не создаёт второе письмо.
- `research_required` не получает автоматический статус/расписание.

## Контроль

**Ревью Developer:** конфигурация, idempotency, часовые пояса и границы времени.  
**QA:** повторный scheduler, даты, неверный порог и недоступный mail adapter.
