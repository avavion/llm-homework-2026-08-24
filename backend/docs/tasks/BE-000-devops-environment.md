# BE-000 — DevOps-среда Go + PostgreSQL

**Исполнитель:** System Administrator / DevOps  
**Статус:** draft  
**Зависимости:** нет  
**Timebox:** до 2 рабочих дней после технического одобрения
**План:** [2026-08-26-be-000-devops-environment.md](../superpowers/plans/2026-08-26-be-000-devops-environment.md)

## SMART-цель

Подготовить воспроизводимую локальную среду, в которой Go API и PostgreSQL запускаются Docker Compose-командой, проходят health check и могут быть собраны под Linux, macOS и Windows без хранения секретов в Git.

## Границы

- `backend/docker-compose.yml`, `.env.example`, Makefile, Dockerfile, документация среды и кроссплатформенные build-команды.
- PostgreSQL, API, миграции, health check, internal network и именованный volume.
- Не включает production deployment, CI/CD, облачную инфраструктуру, реальные e-mail/OCR/LLM ключи и прикладную бизнес-логику.

## Критерии приёмки

- `docker compose up --build` поднимает PostgreSQL и API; API проходит health check.
- `.env.example` содержит только имена и безопасные примеры переменных, а локальный `.env` игнорируется Git.
- PostgreSQL не публикует порт наружу без документированной причины; данные хранятся в именованном volume.
- Makefile содержит `up`, `down`, `logs`, `migrate-up`, `test` и `build-linux-amd64`, `build-darwin-amd64`, `build-darwin-arm64`, `build-windows-amd64`.
- `backend/docs/environment.md` описывает версии Go/PostgreSQL, запуск, остановку, миграции и известные ограничения.

## Обязательное техническое ревью

Senior Backend Developer проверяет совместимость Docker Compose с будущим Go API, `DATABASE_URL`, миграциями и локальным workflow. Статус задачи не становится `approved` без этого ревью.

## QA-проверка

Senior QA запускает среду в чистой локальной директории, проверяет отсутствие секретов в compose/документации/логах, health endpoint, остановку с удалением volume и наличие всех четырёх артефактов cross-build.
