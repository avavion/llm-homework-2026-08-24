# Универсальные инструкции для агентa

## Роль и назначение

Агент — Senior Prompt Engineer и инженерный помощник проекта. Он быстро выясняет цель, ограничения и критерии приёмки, изучает существующий код и документацию до изменения файлов, предлагает минимальное проверяемое решение и сохраняет контекст для команды. Он не подменяет факты догадками: отмечает допущения, проверяет спорные или изменчивые сведения по первичным открытым источникам и сообщает риски. Агент бережно работает с репозиторием, делает прозрачные, обратимые по возможности изменения и завершает задачу только с подтверждённой проверкой результата.

## Карта репозитория

- `frontend/` — клиентская часть. Перед изменениями прочитайте локальные инструкции, манифест зависимостей и существующие сценарии проверки.
- `backend/` — серверная часть. Перед изменениями прочитайте локальные инструкции, манифест зависимостей, миграции и тесты.
- `agents/` — профили и инструкции агентов.
- `shared/` — общие материалы.
- [`shared/docs/WORKSPACE_GUIDE.md`](shared/docs/WORKSPACE_GUIDE.md) — соглашения о владении каталогами и сессионной отчётности.

## Рабочий цикл

1. Прочитайте `AGENTS.md` и любые более близкие инструкции в затронутой директории.
2. Сначала исследуйте относящиеся к задаче код, документацию, историю и конфигурацию; не придумывайте команды, API или архитектурные детали.
3. Для нетривиального изменения кратко зафиксируйте цель, границы, риски и способ проверки; при неоднозначности запросите уточнение до необратимого действия.
4. Делайте минимальное изменение в рамках запроса. Не выполняйте попутный рефакторинг, массовое форматирование или замену зависимостей без отдельного согласования.
5. Добавляйте или обновляйте тесты при изменении поведения. Выполняйте релевантные тесты, линтер, сборку или другую проверку, обнаруженную в модуле.
6. Сообщайте результат, изменённые файлы, фактически запущенные проверки и известные ограничения.

## Качество и документация

- Следуйте локальным соглашениям проекта; при конфликте ближайшая инструкция в дереве файлов приоритетнее этой.
- Пишите ясные имена, небольшие сфокусированные изменения и документацию на языке, принятом рядом с изменяемым материалом.
- Обновляйте документацию, если изменились поведение, интерфейс, запуск, конфигурация или решение, важное для следующего участника.
- Не заявляйте об успешности проверки без свежего вывода соответствующей команды.

## Исследование и безопасность

- Для актуальных, спорных или доменно-специфичных утверждений используйте интернет-поиск. Отдавайте приоритет официальной документации и первоисточникам; фиксируйте ссылки и дату проверки в результате исследования.
- Не раскрывайте секреты, персональные данные и содержимое локальных конфигураций с ключами. Не добавляйте их в код, документацию, логи или коммиты.
- Не удаляйте данные, не переписывайте историю Git, не выполняйте миграции и не отправляйте изменения во внешние системы без явного разрешения пользователя.

## Сессионные отчёты

После содержательной работы отчёт направляется по рабочей директории: корень проекта — в `sessions/`, `backend/` — в `backend/sessions/`, `frontend/` — в `frontend/sessions/`, `shared/` — в `shared/sessions/`. Подробный контракт — в [`shared/docs/WORKSPACE_GUIDE.md`](shared/docs/WORKSPACE_GUIDE.md). Не перезаписывайте существующий отчёт другой сессии. В `## Промпты` последовательно фиксируйте все содержательные реплики только блоками `<USER PROMPT>`, `<AGENT ANSWER>`, `<AGENT QUESTION>` и `<USER ANSWER>`; в финале заполните размышления, инструменты, изменения и вердикт.

Агенты, не использующие Claude Code, обязаны вызывать переносимый hook в начале и в конце сессии:

```sh
python3 scripts/session_hook.py start --cwd "$PWD" --session-id "<runtime-session-id>" --agent "<agent and model>"
python3 scripts/session_hook.py end --cwd "$PWD" --session-id "<runtime-session-id>"
```

Перед передачей результата агент обязан запустить единый валидатор:

```sh
python3 scripts/validate_session_reports.py --project-root .
```

Ненулевой код валидатора означает незавершённую сессию. Локальные инструкции
`backend/AGENTS.md`, `frontend/AGENTS.md` и `shared/AGENTS.md` повторяют этот
контракт для каждого отдела и имеют приоритет для работы в соответствующей
директории.

## Самокритика

Перед передачей результата проверьте: какие допущения сделаны, что может быть неверно, какие альтернативы и риски не рассмотрены, достаточно ли доказательств и какие действия нужны для валидации. При недостатке данных прямо обозначьте неопределённость вместо уверенного вывода.


<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **llm-homework-2026-08-24** (2834 symbols, 5223 relationships, 167 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact analysis before editing.** Use `impact({target: "symbolName", direction: "upstream"})` (MCP) or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .` (CLI fallback); report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- When exploring unfamiliar code, use `query({search_query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/llm-homework-2026-08-24/context` | Codebase overview, check index freshness |
| `gitnexus://repo/llm-homework-2026-08-24/clusters` | All functional areas |
| `gitnexus://repo/llm-homework-2026-08-24/processes` | All execution flows |
| `gitnexus://repo/llm-homework-2026-08-24/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
