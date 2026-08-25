# Руководство по рабочему пространству

## Назначение каталогов

- `backend/` — серверная часть проекта. Его документация принадлежит `backend/docs/`, а отчёты сессий — `backend/sessions/`.
- `frontend/` — клиентская часть проекта. Его документация принадлежит `frontend/docs/`, а отчёты сессий — `frontend/sessions/`.
- `shared/` — материалы, общие для всех частей проекта. Общая документация принадлежит `shared/docs/`, а общие отчёты сессий — `shared/sessions/`.
- `docs/` в корне — документация уровня всего проекта.
- `sessions/` в корне — аудит сессий, работающих из корня проекта; такие отчёты имеют `department: root`.

Рабочая директория определяет место отчёта: путь внутри `backend/`, `frontend/` или `shared/` направляется в соответствующий каталог `sessions/`; корень проекта направляется в `sessions/`. Отчёты никогда не размещаются в других каталогах и не перезаписываются отчётом другой сессии.

## Контракт сессионного отчёта

Каждый отчёт — Markdown-файл `session-YYYY-MM-DD-HHMMSS.md` с YAML frontmatter. Frontmatter обязан содержать следующие ключи:

```yaml
---
hook: session.started
session_id: <runtime-session-id>
department: root
started_at: 2026-08-25T23:45:16+03:00
completed_at:
---
```

Значение `hook` при завершении изменяется на `session.completed`, а `completed_at` заполняется временем завершения. `department` принимает одно из значений `root`, `backend`, `frontend` или `shared`.

После frontmatter обязательны разделы:

1. Запрос
2. Действия
3. Источники
4. Созданные артефакты
5. Выводы
6. Риски
7. Самокритика
8. Следующий шаг

## Переносимый запуск

Claude Code вызывает hook автоматически. Все остальные runtime должны запускать его из рабочей директории в начале и в конце сессии, передав свой неизменный идентификатор сессии:

```sh
python3 scripts/session_hook.py start --cwd "$PWD" --session-id "<runtime-session-id>"
python3 scripts/session_hook.py end --cwd "$PWD" --session-id "<runtime-session-id>"
```
