# Food Expiry — локальный production-запуск

Из корня проекта запустите:

```sh
make up
```

Команда устанавливает зависимости frontend строго по `package-lock.json`, создаёт production-сборку в `frontend/dist` с `VITE_API_URL`, собирает Go API в `bin/api`, поднимает Docker-сервисы API и PostgreSQL, применяет миграции и запускает production preview frontend. Адреса выводятся в терминал: frontend по умолчанию доступен на `http://localhost:4173`, API — на `http://localhost:8080`.

`bin/api` — собранный Go backend для текущей платформы. В `make up` API запускается как production-бинарник `/api` внутри Docker-сети вместе с PostgreSQL; frontend не обслуживается этим бинарником. Vite preview раздаёт готовую production-сборку frontend.

Дополнительные команды:

```sh
make build # только собрать frontend и Go binary
make logs  # вывести последние логи процессов
make down  # остановить процессы и PostgreSQL, сохранив локальные данные
```

Для изменения портов задайте `API_PORT` или `FRONTEND_PORT` перед командой `make up`; скрипт автоматически соберёт frontend с соответствующим адресом API.
