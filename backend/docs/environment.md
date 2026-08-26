# Локальная среда backend

Среда предназначена только для локальной разработки. API доступен с хоста по
опубликованному порту и также подключён к внутренней сети PostgreSQL;
PostgreSQL не публикует порт на хост.

## Версии

- Go: `1.25.0` (образ `golang:1.25.0-alpine3.22`).
- PostgreSQL: `17.6` (образ `postgres:17.6-alpine3.22`).
- Инструмент миграций: `golang-migrate` `v4.18.3` внутри образа API.

## Настройка и запуск

1. Скопируйте пример окружения: `cp .env.example .env`.
2. При необходимости измените безопасные локальные значения в `.env`.
   Не добавляйте этот файл в Git.
3. Запустите среду: `make up`.
4. Проверьте API, загрузив порт из `.env`:
   `set -a && . ./.env && set +a && curl --fail "http://localhost:${API_PORT}/healthz"`.
5. Остановите среду и удалите локальные данные PostgreSQL: `make down`.

`DATABASE_URL` внутри Compose обращается к хосту `postgres`. Значения
`POSTGRES_*`, `DATABASE_URL` и `API_PORT` задаются в `.env`; в репозитории
остаётся только безопасный пример.

## Операторские команды

| Команда | Назначение |
| --- | --- |
| `make up` | Собрать и запустить API и PostgreSQL в фоне. |
| `make down` | Остановить Compose и удалить `postgres_data` (`down -v`). |
| `make logs` | Следить за логами сервисов. |
| `make run` | Собрать и запустить Compose API из `cmd/api` вместе с PostgreSQL. |
| `make migrate-up` | Выполнить SQL-файлы из `/migrations` в API-контейнере. |
| `make test` | Запустить `go test ./...`. |
| `make test-integration` | В изолированном Compose-проекте применить миграции и запустить PostgreSQL auth integration test, затем удалить тестовые контейнеры и volume. |
| `make build-linux-amd64` | Создать `bin/api-linux-amd64`. |
| `make build-darwin-amd64` | Создать `bin/api-darwin-amd64`. |
| `make build-darwin-arm64` | Создать `bin/api-darwin-arm64`. |
| `make build-windows-amd64` | Создать `bin/api-windows-amd64.exe`. |

## Точка интеграции BE-001

BE-001 поставляет `go.mod`, API из `cmd/api`, health endpoint `GET /healthz` и
миграции `000001_init.up.sql` / `000001_init.down.sql`. API слушает `API_PORT`;
`make run` запускает его через Compose, а `make migrate-up` использует
контейнерную переменную `DATABASE_URL`, бинарник `/migrate` и SQL-файлы из
`migrations/`.

Начальная миграция оставляет bookkeeping `golang-migrate` самому инструменту
и не создаёт прикладных таблиц. Схемы accounts, products и другие feature
schemas относятся к следующим backend-задачам.

## PostgreSQL integration tests

Воспроизводимая проверка BE-003 запускается из `backend/` одной командой:

```sh
make test-integration
```

Она использует отдельный Compose project
`llm-homework-backend-integration`, поднимает только PostgreSQL и tagged Go
test-контейнер, применяет миграции через тот же `golang-migrate`, затем запускает
`go test -tags=integration -count=1 ./internal/auth`. PostgreSQL остаётся только
во внутренней сети Compose и не публикует порт на хост. После любого исхода
команды trap удаляет тестовые контейнеры, сети и volume, поэтому проверка не
зависит от ранее сохранённого состояния. Make target также принудительно задаёт
безопасные локальные test credentials и внутренний host `postgres`, поэтому
унаследованный из shell `DATABASE_URL` не может перенаправить тест на внешнюю БД.

Эквивалентный Compose-сценарий с тем же обязательным cleanup:

```sh
integration_project=llm-homework-backend-integration
cleanup() {
  docker compose -p "$integration_project" --profile test down -v --remove-orphans
}
trap cleanup EXIT HUP INT TERM

POSTGRES_DB=llm_homework \
POSTGRES_USER=app \
POSTGRES_PASSWORD=local-dev-password \
DATABASE_URL='postgres://app:local-dev-password@postgres:5432/llm_homework?sslmode=disable' \
docker compose -p "$integration_project" --profile test up \
  --build --abort-on-container-exit \
  --exit-code-from auth-integration-test auth-integration-test
```

## Ограничения

Это не production-конфигурация: здесь нет CI/CD, облачных ресурсов, внешних
интеграций, резервного копирования и управления настоящими секретами. Команда
`make down` удаляет локальный named volume `postgres_data`; не используйте её
для данных, которые необходимо сохранить.
