# Food Expiry

Веб-приложение для домашнего учёта продуктов с ограниченным сроком годности.
Пользователь может зарегистрироваться, вести запасы, видеть статус продукта,
отмечать его использованным или выброшенным, а также работать с рецептами и
настройками уведомлений. Интерфейс доступен на русском и английском языках.

Проект состоит из React/Vite-клиента в `frontend/`, Go API в `backend/` и
PostgreSQL. API и база данных запускаются в Docker, а собранный frontend
раздаётся Vite Preview на хосте.

## Требования

- Docker с плагином Docker Compose (`docker compose`);
- Node.js и npm — для установки зависимостей и сборки frontend;
- Go **1.25** — корневой сценарий собирает API локально перед запуском;
- свободные порты `4173` (frontend) и `8080` (API).

Проверить инструменты можно так:

```sh
docker compose version
node --version
npm --version
go version
```

## Быстрый запуск

Из корня репозитория выполните:

```sh
make up
```

Команда:

1. устанавливает frontend-зависимости строго по `frontend/package-lock.json`;
2. собирает frontend с адресом API `http://localhost:8080`;
3. собирает Go API в `bin/api`;
4. запускает API и PostgreSQL в Docker;
5. применяет миграции базы данных;
6. запускает production-preview frontend.

После успешного запуска откройте:

- Frontend: <http://localhost:4173>
- API: <http://localhost:8080>
- Проверка API: <http://localhost:8080/healthz>

Остановить локальное окружение, сохранив данные PostgreSQL:

```sh
make down
```

## Полезные команды

```sh
make build                 # собрать frontend и Go API без запуска
make logs                  # показать последние логи локальных процессов

cd backend && make test    # запустить тесты Go API
cd backend && make test-integration  # запустить интеграционный auth-тест в Docker

cd frontend && npm run lint  # проверить TypeScript
cd frontend && npm test      # запустить unit-тесты frontend
cd frontend && npm run test:e2e  # запустить Playwright-тесты
```

Для frontend-команд при первом запуске сначала выполните `cd frontend && npm ci`.
Для интеграционного теста backend нужен работающий Docker.

## Настройка портов

Порты можно изменить перед запуском:

```sh
API_PORT=8090 FRONTEND_PORT=4174 make up
```

Сценарий автоматически соберёт frontend с новым адресом API. Если вы запускаете
frontend отдельно на другом origin, настройте `FRONTEND_ORIGINS` для CORS;
подробная инструкция есть в
[shared/docs/frontend-backend-local-setup.md](shared/docs/frontend-backend-local-setup.md).

## Ограничения текущего MVP

Распознавание продуктов по фотографии имеет интерфейс и тестовые реализации,
но реальный OCR/LLM-провайдер не настроен по умолчанию. Ручное добавление
продуктов остаётся рабочим базовым сценарием. Полные границы MVP и правила
обработки дат описаны в [shared/docs/product-description.md](shared/docs/product-description.md).

## Вклад в проект

Таблица составлена по авторам и содержанию коммитов в истории Git; она не
приписывает кому-либо незакоммиченные файлы в рабочей директории.

| Участник | Основной вклад |
| --- | --- |
| `avavion` | Backend и локальная инфраструктура: Go API, PostgreSQL и миграции, авторизация и сессии, продукты, профили, рецепты, уведомления и распознавание черновиков. Также подготовлены API/интеграционные тесты, сценарии запуска, документация backend, QA-отчёты и часть финальной доводки frontend. |
| `violetta_chernysheva` | Frontend и продуктовый дизайн: React/Vite-основа, интерфейсы запасов, авторизации и жизненного цикла продуктов, локализация, адаптивная вёрстка и дизайн-токены. Также подготовлены дизайн- и frontend-спецификации, accessibility-улучшения и визуальная полировка интерфейса. |

## Документация

- [Workflow: независимая проверка готовности релиза](WORKFLOW.md)
- [Отчёт к домашнему заданию №2](REPORT.md)
- [Описание продукта](shared/docs/product-description.md)
- [Справочник API](shared/docs/backend-api-reference.md)
- [Локальная интеграция frontend и backend](shared/docs/frontend-backend-local-setup.md)
- [Окружение backend](backend/docs/environment.md)
