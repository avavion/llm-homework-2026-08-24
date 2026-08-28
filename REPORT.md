# Отчёт к домашнему заданию №2: агенты, workflow и изоляция

## 1. Что сделано

**Food Expiry** — веб-приложение для домашнего учёта продуктов с ограниченным
сроком годности. Пользователь может зарегистрироваться, вести список запасов,
видеть статусы продуктов, отмечать их использованными или выброшенными,
работать с рецептами и настройками уведомлений; интерфейс поддерживает русский
и английский языки. Инструкция по локальному запуску и требованиям находится в
[README.md](README.md).

## 2. Работа агентов

В артефактах репозитория представлены роли продуктового/PM-агента
([`agents/base-product-agent.md`](agents/base-product-agent.md)), backend и
frontend разработки (реестры
[backend-задач](backend/docs/task-register.md) и
[frontend-задач](frontend/docs/task-register.md)), а также QA/white-hat
приёмки ([BE-007](backend/docs/tasks/BE-007-backend-acceptance-whitehat.md) и
[FE-007](frontend/docs/tasks/FE-007-qa-whitehat-and-visual-acceptance.md)).

Историческая сессия [BE-007](sessions/session-2026-08-26-161236.md) прямо
фиксирует dispatch исполнителей и ревьюеров как subagents, а также работу в
изолированном worktree `.worktrees/be007-qa-whitehat` на ветке
`feat/be007-qa-whitehat`. В текущем списке `git worktree` этот worktree уже не
присутствует; это свидетельство исторического процесса, а не утверждение о
текущем состоянии. В текущем репозитории есть другой связанный worktree:
`.worktrees/editorial-pantry-redesign` на ветке
`codex/editorial-pantry-redesign`.

## 3. Независимая проверка / агент-ломатель

Подробный результат — в
[docs/independent-negative-test-report-2026-08-28.md](docs/independent-negative-test-report-2026-08-28.md).
Он включает пустые и повреждённые payload’ы, неожиданные поля и длинные
строки, числовые границы, чужие ID, повтор lifecycle-вызовов, неверный порядок
операций и пустую загрузку файла.

В среде этого прогона API не стартовал: `make up` остановился на отсутствии Go,
а Docker также был недоступен. Поэтому все 12 HTTP-попыток получили connection
refused до обработчиков; прикладные endpoint’ы, изоляция аккаунтов и гонки не
подтверждены runtime-проверкой.

## 4. Workflow

[WORKFLOW.md](WORKFLOW.md) задаёт повторяемую независимую проверку следующей
версии: prerequisites, `make up`, health-check, backend/frontend-проверки и
минимальный набор негативных сценариев. Его запускают на чистой ветке,
последовательно выполняя команды из раздела «Повторный запуск»; результатом
становится новый фактический отчёт `docs/release-readiness-YYYY-MM-DD.md` со
статусом `passed`, `changes_required` или `blocked`.

## 5. Что не сработало / ограничения

- В части QA-сред недоступны Go 1.25 и Docker, поэтому запуск API, миграции и
  интеграционные runtime-проверки блокируются.
- Реальный OCR/LLM-провайдер по умолчанию не настроен; распознавание фото не
  следует считать работающим production-сценарием.
- `python3 scripts/validate_session_reports.py --project-root .` действительно
  находит неполные исторические отчёты, в том числе
  [`sessions/session-2026-08-26-151506.md`](sessions/session-2026-08-26-151506.md),
  [`sessions/session-2026-08-26-222708.md`](sessions/session-2026-08-26-222708.md)
  и [`frontend/sessions/session-2026-08-26-231136.md`](frontend/sessions/session-2026-08-26-231136.md).
- Перед сдачей нужно разобрать незакоммиченные файлы, замеченные через
  `git status --short`: `README.md`, `REPORT.md`, `WORKFLOW.md`,
  `docs/independent-negative-test-report-2026-08-28.md`,
  `sessions/session-2026-08-28-225625.md`,
  `sessions/session-2026-08-28-230019.md`,
  `sessions/session-2026-08-28-230316.md`,
  `sessions/session-2026-08-28-232908.md` и
  `sessions/session-2026-08-28-233253.md`.

## 6. Проверки и следующий шаг

Команды проверки из README (это перечень для запуска, а не заявление о
выполненном успешном прогоне):

```sh
make build
cd backend && make test
cd backend && make test-integration
cd frontend && npm run lint
cd frontend && npm test
cd frontend && npm run test:e2e
```

Перед отправкой:

- [ ] Проверить, что в diff и отчётах нет секретов, cookie и персональных данных.
- [ ] Завершить проверку изменений, добавить нужные файлы и создать осмысленный commit.
- [ ] Убедиться, что GitHub-репозиторий доступен преподавателю.

| Артефакт | Путь | Зачем нужен для сдачи |
| --- | --- | --- |
| Описание и запуск | [README.md](README.md) | Объясняет продукт и локальный старт. |
| Повторяемая приёмка | [WORKFLOW.md](WORKFLOW.md) | Фиксирует release-процедуру и статусы. |
| Независимый QA-отчёт | [independent-negative-test-report](docs/independent-negative-test-report-2026-08-28.md) | Показывает реальные негативные попытки и блокер среды. |
| Роли и задачи | [agents/](agents/) и [task registers](backend/docs/task-register.md) | Подтверждают организацию агентной работы. |
| Изоляция subagents/worktree | [BE-007 session](sessions/session-2026-08-26-161236.md) | Документирует исторический изолированный процесс. |
