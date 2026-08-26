# BE-001 — Каркас Go API и PostgreSQL

**Исполнитель:** Senior Backend Developer · **Статус:** draft · **Зависимости:** нет · **Timebox:** 2 дня

**План:** [2026-08-26-be-001-api-postgres-foundation.md](../superpowers/plans/2026-08-26-be-001-api-postgres-foundation.md)

## SMART-цель

Создать локально воспроизводимый Go/PostgreSQL-сервис, в котором `GET /healthz` возвращает HTTP 200 и `{"status":"ok"}`.

## Объём и приёмка

- Создать `go.mod`, `cmd/api`, конфигурацию `DATABASE_URL`, HTTP-сервер, первую миграцию, Docker Compose и Makefile.
- `make run`, `make migrate-up`, `go test ./...` и `go vet ./...` проходят.
- Не включать аккаунты, продукты, e-mail или OCR.

## Контроль

**Ревью Developer:** пакеты, миграции, конфигурация и тестируемость без внешних сервисов.  
**QA:** отсутствующий `DATABASE_URL`, отсутствие секретов в репозитории и доступность health endpoint.
