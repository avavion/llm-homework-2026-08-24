# Как frontend поднимает backend локально

Эта инструкция — для frontend-разработчика, которому нужно поднять backend
рядом со своим dev-сервером и реально дёргать API (не мокать). Справочник по
самим эндпоинтам — [`backend-api-reference.md`](backend-api-reference.md).
Полная (более техническая) версия этой инструкции для backend-разработчиков —
[`backend/docs/environment.md`](../../backend/docs/environment.md); здесь —
сокращённый путь именно под нужды frontend.

⚠️ Backend активно разрабатывается, возможны баги и незадокументированные
изменения. Эта конфигурация — только для локальной разработки, не для
продакшена.

## Что понадобится

- Docker и Docker Compose (Docker Desktop либо аналог).
- Ничего больше — Go и PostgreSQL ставить на хост не нужно, всё поднимается
  в контейнерах.

## Шаги

1. Перейдите в каталог backend:

   ```sh
   cd backend
   ```

2. Скопируйте пример окружения и при необходимости отредактируйте:

   ```sh
   cp .env.example .env
   ```

   По умолчанию API поднимется на `http://localhost:8080` и уже разрешит
   CORS-запросы с `http://localhost:5173` (Vite) и `http://localhost:3000`
   (CRA/Next.js). **Если ваш dev-сервер слушает другой порт** — допишите его
   в `.env` в переменную `FRONTEND_ORIGINS` (через запятую, без пробелов
   внутри значения обязательно быть не должно, но пробелы после запятой
   допустимы):

   ```
   FRONTEND_ORIGINS=http://localhost:5173,http://localhost:4173
   ```

   Не добавляйте `.env` в Git — это локальный файл.

3. Поднимите API и PostgreSQL:

   ```sh
   make up
   ```

   Это соберёт образ backend и запустит два контейнера: `api` и `postgres`.
   PostgreSQL не публикует порт наружу — обращаться к нему напрямую с хоста
   не нужно и не получится, только через API.

4. Примените миграции базы данных (обязательно после первого поднятия и
   после каждого `make down`, который стирает volume):

   ```sh
   make migrate-up
   ```

5. Проверьте, что API отвечает:

   ```sh
   set -a && . ./.env && set +a
   curl --fail "http://localhost:${API_PORT}/healthz"
   ```

   Ожидаемый ответ: `{"status":"ok"}`.

6. Настройте свой frontend-клиент на этот базовый URL, например через
   `.env`/`.env.local` вашего фронтенд-проекта:

   ```
   VITE_API_BASE_URL=http://localhost:8080
   ```

   и обязательно отправляйте запросы с credentials, иначе сессионная cookie
   не долетит до backend и обратно:

   ```js
   fetch(`${API_BASE_URL}/v1/products`, { credentials: "include" })
   ```

   ```js
   axios.create({ baseURL: API_BASE_URL, withCredentials: true })
   ```

## Быстрая проверка руками (без вашего frontend)

Полный сценарий curl — регистрация → логин → создание продукта → список:

```sh
BASE=http://localhost:8080
COOKIES=/tmp/backend-cookies.txt

curl -s -c "$COOKIES" -X POST $BASE/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"a-long-enough-passphrase"}'

curl -s -c "$COOKIES" -b "$COOKIES" -X POST $BASE/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"dev@example.com","password":"a-long-enough-passphrase"}'

curl -s -b "$COOKIES" -X POST $BASE/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Milk","date_type":"use_by","expiry_date":"2026-09-01T00:00:00Z"}'

curl -s -b "$COOKIES" $BASE/v1/products
```

Если последняя команда возвращает JSON-массив с созданным продуктом — backend
поднят правильно и сессия работает.

## Повседневные команды

| Команда (из `backend/`) | Что делает |
| --- | --- |
| `make up` | Собрать и запустить API + PostgreSQL в фоне. |
| `make migrate-up` | Применить SQL-миграции к базе. Нужно после каждого `make up` с нуля. |
| `make logs` | Смотреть логи API и PostgreSQL (полезно при `401`/`500`, чтобы увидеть stacktrace). |
| `make down` | Остановить всё и **удалить данные PostgreSQL** (volume). Используйте, когда нужно начать с чистой базы. |

## Если что-то не работает

- **Запрос падает с CORS-ошибкой в консоли браузера, а не с `401`.**
  Origin вашего dev-сервера не в `FRONTEND_ORIGINS` — допишите его в `.env`
  backend и перезапустите `make up` (или `docker compose restart api`,
  если база уже поднята и трогать её не нужно).
- **Всегда `401` на защищённых эндпоинтах, хотя `login` вернул `200`.**
  Проверьте, что запросы уходят с `credentials: 'include'`
  (`withCredentials: true` в axios), и что и backend, и frontend открыты
  именно через `http://localhost:...`, а не через `127.0.0.1` или IP —
  cookie с флагом `Secure` по `http://` браузеры считают безопасной только
  для `localhost`.
- **`make migrate-up` падает или API не отвечает.** Проверьте
  `make logs` — чаще всего это гонка при первом старте (PostgreSQL ещё не
  готов) или забытый `.env`. Полностью пересоздать окружение: `make down`
  (сотрёт данные) → `make up` → `make migrate-up`.
- **`POST /v1/product-drafts/recognize` всегда возвращает `400`.** Это
  ожидаемо в текущем состоянии backend — реальный OCR/LLM ещё не подключён,
  см. [`backend-api-reference.md`](backend-api-reference.md#известные-ограничения-и-не-закрытые-баги).
  Используйте `/v1/products` для ручного ввода при разработке UI.
- Что-то ещё не сходится с документом — считайте документ потенциально
  устаревшим (backend меняется) и уточняйте у backend-команды.
