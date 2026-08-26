# Frontend: обязательный журнал сессии

Это правило обязательно для всех агентов, работающих в `frontend/`.

1. До первой содержательной работы запустите:

   ```sh
   project_root="$(git rev-parse --show-toplevel)"
   working_directory="$PWD"
   python3 "$project_root/scripts/session_hook.py" start --project-root "$project_root" --cwd "$working_directory" --session-id "<runtime-session-id>" --agent "<agent and model>"
   ```

2. В созданном `frontend/sessions/session-YYYY-MM-DD-HHMMSS.md` запишите все
   содержательные реплики в `## Промпты` блоками `<USER PROMPT>`, `<AGENT ANSWER>`,
   `<AGENT QUESTION>` и `<USER ANSWER>`. Не придумывайте отсутствующие сообщения.
3. Перед передачей результата заполните размышления, инструменты, изменения и
   все поля финального вердикта; затем завершите сессию и проверьте весь журнал:

   ```sh
   python3 "$project_root/scripts/session_hook.py" end --project-root "$project_root" --cwd "$working_directory" --session-id "<runtime-session-id>"
   python3 "$project_root/scripts/validate_session_reports.py" --project-root "$project_root"
   ```

Невалидный отчёт означает, что frontend-сессия не завершена. Архивные отчёты
старого формата не переписываются. Общие правила: `../AGENTS.md` и
`../shared/docs/WORKSPACE_GUIDE.md`.
