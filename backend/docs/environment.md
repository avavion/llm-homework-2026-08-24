# Локальная среда backend

Среда предназначена только для локальной разработки. Она запускает Go API и
PostgreSQL во внутренней Docker-сети; PostgreSQL не публикует порт на хост.

## Версии

- Go: `1.25.0` (образ `golang:1.25.0-alpine3.22`).
- PostgreSQL: `17.6` (образ `postgres:17.6-alpine3.22`).
- Инструмент миграций: `golang-migrate` `v4.18.3` внутри образа API.

## Настройка и запуск

1. Скопируйте пример окружения: `cp .env.example .env`.
2. При необходимости измените безопасные локальные значения в `.env`.
   Не добавляйте этот файл в Git.
3. Запустите среду: `make up`.
4. Проверьте API: `curl --fail http://localhost:${API_PORT:-8080}/healthz`.
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
| `make migrate-up` | Выполнить SQL-файлы из `/migrations` в API-контейнере. |
| `make test` | Запустить `go test ./...`. |
| `make build-linux-amd64` | Создать `bin/api-linux-amd64`. |
| `make build-darwin-amd64` | Создать `bin/api-darwin-amd64`. |
| `make build-darwin-arm64` | Создать `bin/api-darwin-arm64`. |
| `make build-windows-amd64` | Создать `bin/api-windows-amd64.exe`. |

## Точка интеграции BE-001

Dockerfile и Makefile ожидают, что следующий пакет добавит `go.mod`, `go.sum`,
`cmd/api` и каталог `migrations/`. После этого API обязан слушать
`API_PORT` и предоставлять `GET /healthz`; `make migrate-up` использует его
контейнерную переменную `DATABASE_URL` и SQL-файлы из `migrations/`.

До выполнения BE-001 `make up`, `make test`, cross-build цели и `make
migrate-up` намеренно неработоспособны: в рабочем дереве отсутствует Go API и
миграции. Конфигурацию Compose уже можно проверять командой `docker compose
config`.

## Ограничения

Это не production-конфигурация: здесь нет CI/CD, облачных ресурсов, внешних
интеграций, резервного копирования и управления настоящими секретами. Команда
`make down` удаляет локальный named volume `postgres_data`; не используйте её
для данных, которые необходимо сохранить.
