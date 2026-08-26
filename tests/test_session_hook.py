"""Integration tests for the portable session accounting hook."""

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]
HOOK = PROJECT_ROOT / "scripts" / "session_hook.py"
VALIDATOR = PROJECT_ROOT / "scripts" / "validate_session_reports.py"
TIMESTAMP = "2026-08-25T23:45:16+03:00"


class SessionHookTest(unittest.TestCase):
    def test_validator_accepts_completed_reports_in_all_scopes(self):
        """Every new-format report in root and departments passes one validator."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            for scope, department in (
                ("sessions", "root"),
                ("backend/sessions", "backend"),
                ("frontend/sessions", "frontend"),
                ("shared/sessions", "shared"),
            ):
                report = project_root / scope / "session-2026-08-26-100000.md"
                report.parent.mkdir(parents=True, exist_ok=True)
                report.write_text(self.valid_report(department), encoding="utf-8")

            result = subprocess.run(
                [sys.executable, str(VALIDATOR), "--project-root", str(project_root)],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertEqual(result.returncode, 0, result.stderr)

    def test_validator_rejects_new_report_without_dialogue(self):
        """A new-format report cannot pass with template-only dialogue content."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            report = project_root / "backend/sessions/session-2026-08-26-100000.md"
            report.parent.mkdir(parents=True)
            report.write_text(
                "<!-- hook: session.completed -->\n\n# Сессия: 2026-08-26 10:00:00\n"
                "\n- **Агент:** Backend Agent\n- **Отдел:** backend\n"
                "- **Статус:** завершена\n\n## Промпты\n\n## Размышления\n"
                "\n## Использованные инструменты\n\n## Изменения в проекте\n"
                "\n## Финальный вердикт\n\n**Закрыто:**\n\n**Не закрыто:**"
                "\n\n**Что сломано:**\n\n**С чего продолжать:**\n",
                encoding="utf-8",
            )

            result = subprocess.run(
                [sys.executable, str(VALIDATOR), "--project-root", str(project_root)],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0)
            self.assertIn(str(report), result.stderr)

    def test_start_creates_conversational_template_with_agent(self):
        """A new report contains the agreed dialogue journal skeleton and agent."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            result = subprocess.run(
                [
                    sys.executable,
                    str(HOOK),
                    "start",
                    "--project-root",
                    str(project_root),
                    "--cwd",
                    str(project_root),
                    "--session-id",
                    "dialogue-1",
                    "--timestamp",
                    TIMESTAMP,
                    "--agent",
                    "Codex GPT-5.6 Terra Med",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            report = project_root / "sessions" / "session-2026-08-25-234516.md"
            content = report.read_text(encoding="utf-8")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("# Сессия: 2026-08-25 23:45:16", content)
            self.assertIn("- **Агент:** Codex GPT-5.6 Terra Med", content)
            self.assertIn("- **Статус:** активна", content)
            self.assertIn("## Промпты", content)
            self.assertIn("### <USER PROMPT>", content)
            self.assertIn("### <AGENT ANSWER>", content)
            self.assertIn("### <AGENT QUESTION>", content)
            self.assertIn("### <USER ANSWER>", content)
            self.assertIn("## Размышления", content)
            self.assertIn("## Использованные инструменты", content)
            self.assertIn("## Изменения в проекте", content)
            self.assertIn("## Финальный вердикт", content)

    def test_start_routes_backend(self):
        """A backend session creates its report only in backend/sessions."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            backend = project_root / "backend"
            backend.mkdir()

            result = subprocess.run(
                [
                    sys.executable,
                    str(HOOK),
                    "start",
                    "--project-root",
                    str(project_root),
                    "--cwd",
                    str(backend),
                    "--session-id",
                    "backend-1",
                    "--timestamp",
                    TIMESTAMP,
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            report = backend / "sessions" / "session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue(report.is_file())
            self.assertEqual(list((project_root / "sessions").glob("*.md")), [])
            self.assertIn("hook: session.started", report.read_text())
            self.assertIn("session_id: backend-1", report.read_text())
            self.assertIn("department: backend", report.read_text())

    def test_start_routes_project_root(self):
        """A root session creates its report only in the root sessions folder."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            result = self.run_hook(
                "start", project_root, project_root, "root-1", TIMESTAMP
            )

            report = project_root / "sessions" / "session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue(report.is_file())
            self.assertEqual(list((project_root / "backend/sessions").glob("*.md")), [])
            self.assertIn("department: root", report.read_text())

    def test_start_routes_frontend_descendant(self):
        """A frontend descendant writes its report in frontend/sessions."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            working_directory = project_root / "frontend" / "src" / "views"
            working_directory.mkdir(parents=True)
            result = self.run_hook("start", project_root, working_directory, "frontend-1", TIMESTAMP)

            report = project_root / "frontend/sessions/session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue(report.is_file())
            self.assertIn("department: frontend", report.read_text())

    def test_start_routes_shared_descendant(self):
        """A shared descendant writes its report in shared/sessions."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            working_directory = project_root / "shared" / "docs"
            working_directory.mkdir(parents=True)
            result = self.run_hook("start", project_root, working_directory, "shared-1", TIMESTAMP)

            report = project_root / "shared/sessions/session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue(report.is_file())
            self.assertIn("department: shared", report.read_text())

    def test_start_reads_session_id_and_cwd_from_json_stdin(self):
        """JSON stdin supplies missing session ID and working directory values."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            backend = project_root / "backend" / "jobs"
            backend.mkdir(parents=True)
            result = self.run_hook(
                "start",
                project_root,
                None,
                None,
                TIMESTAMP,
                payload={"session_id": "stdin-1", "cwd": str(backend)},
            )

            report = project_root / "backend/sessions/session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue(report.is_file())
            self.assertIn("session_id: stdin-1", report.read_text())

    def test_invalid_arguments_are_nonblocking(self):
        """Missing required CLI input produces stderr but no hook failure."""
        result = subprocess.run(
            [sys.executable, str(HOOK)], capture_output=True, text=True, check=False
        )

        self.assertEqual(result.returncode, 0)
        self.assertIn("session hook:", result.stderr)

    def test_start_is_idempotent(self):
        """A repeated start preserves the first report despite a new timestamp."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            first = self.run_hook("start", project_root, project_root, "same-1", TIMESTAMP)
            second = self.run_hook(
                "start", project_root, project_root, "same-1", "2026-08-25T23:46:16+03:00"
            )

            reports = list((project_root / "sessions").glob("*.md"))
            self.assertEqual(first.returncode, 0, first.stderr)
            self.assertEqual(second.returncode, 0, second.stderr)
            self.assertEqual([report.name for report in reports], ["session-2026-08-25-234516.md"])

    def test_end_finalizes_existing_report(self):
        """An end event updates the existing report instead of creating another."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            self.run_hook("start", project_root, project_root, "finish-1", TIMESTAMP)
            result = self.run_hook("end", project_root, project_root, "finish-1")

            report = project_root / "sessions" / "session-2026-08-25-234516.md"
            self.assertEqual(result.returncode, 0, result.stderr)
            content = report.read_text()
            self.assertIn("hook: session.completed", content)
            self.assertRegex(content, r"completed_at: .+")

    def test_end_marks_new_dialogue_template_completed(self):
        """An end event updates the visible status of the new report template."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            self.run_hook("start", project_root, project_root, "new-finish-1", TIMESTAMP)
            result = self.run_hook("end", project_root, project_root, "new-finish-1")

            report = project_root / "sessions" / "session-2026-08-25-234516.md"
            content = report.read_text(encoding="utf-8")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("<!-- hook: session.completed -->", content)
            self.assertRegex(content, r"<!-- completed_at: .+ -->")
            self.assertIn("- **Статус:** завершена", content)

    def test_start_finds_existing_report_in_another_department(self):
        """An existing root report prevents a duplicate backend report for its ID."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            backend = project_root / "backend"
            backend.mkdir()
            self.run_hook("start", project_root, project_root, "cross-1", TIMESTAMP)
            result = self.run_hook("start", project_root, backend, "cross-1", TIMESTAMP)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(list((backend / "sessions").glob("*.md")), [])

    def test_distinct_session_ids_with_same_timestamp_get_separate_reports(self):
        """A timestamp collision never overwrites another session's report."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            self.run_hook("start", project_root, project_root, "collision-1", TIMESTAMP)
            result = self.run_hook("start", project_root, project_root, "collision-2", TIMESTAMP)

            reports = list((project_root / "sessions").glob("*.md"))
            contents = [report.read_text() for report in reports]
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(len(reports), 2)
            self.assertTrue(any("session_id: collision-1" in content for content in contents))
            self.assertTrue(any("session_id: collision-2" in content for content in contents))

    def test_concurrent_distinct_sessions_with_same_timestamp_keep_all_reports(self):
        """Concurrent starts atomically retain every distinct same-time session."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            session_ids = ["concurrent-" + str(number) for number in range(8)]
            processes = [
                subprocess.Popen(
                    [
                        sys.executable,
                        str(HOOK),
                        "start",
                        "--project-root",
                        str(project_root),
                        "--cwd",
                        str(project_root),
                        "--session-id",
                        session_id,
                        "--timestamp",
                        TIMESTAMP,
                    ],
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    text=True,
                )
                for session_id in session_ids
            ]

            results = [process.communicate() for process in processes]
            reports = list((project_root / "sessions").glob("*.md"))
            found_ids = {
                line.removeprefix("<!-- session_id: ").removesuffix(" -->")
                for report in reports
                for line in report.read_text().splitlines()
                if line.startswith("<!-- session_id: ")
            }
            self.assertTrue(all(process.returncode == 0 for process in processes), results)
            self.assertEqual(len(reports), len(session_ids))
            self.assertEqual(found_ids, set(session_ids))

    def test_end_updates_frontmatter_of_backend_report_only(self):
        """Finalization finds backend reports and leaves body lookalikes intact."""
        with tempfile.TemporaryDirectory() as directory:
            project_root = Path(directory)
            backend = project_root / "backend" / "work"
            backend.mkdir(parents=True)
            self.run_hook("start", project_root, backend, "body-1", TIMESTAMP)
            report = project_root / "backend/sessions/session-2026-08-25-234516.md"
            report.write_text(report.read_text() + "\nhook: body value\ncompleted_at: body value\n")
            result = self.run_hook("end", project_root, project_root, "body-1")

            content = report.read_text()
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("hook: session.completed", content)
            self.assertRegex(content, r"completed_at: .+")
            self.assertIn("hook: body value", content)
            self.assertIn("completed_at: body value", content)

    @staticmethod
    def run_hook(event, project_root, working_directory, session_id, timestamp=None, payload=None):
        command = [
            sys.executable,
            str(HOOK),
            event,
            "--project-root",
            str(project_root),
        ]
        if working_directory is not None:
            command.extend(("--cwd", str(working_directory)))
        if session_id is not None:
            command.extend(("--session-id", session_id))
        if timestamp:
            command.extend(("--timestamp", timestamp))
        return subprocess.run(
            command,
            input=None if payload is None else json.dumps(payload),
            capture_output=True,
            text=True,
            check=False,
        )

    @staticmethod
    def valid_report(department):
        return """<!-- hook: session.completed -->

# Сессия: 2026-08-26 10:00:00

- **Агент:** Test Agent
- **Отдел:** {department}
- **Статус:** завершена

## Промпты

### <USER PROMPT>
> Сделать проверку.

### <AGENT ANSWER>
> Проверка сделана.

## Размышления

Решение проверено.

## Использованные инструменты

| Инструмент | Действие | Зачем |
|---|---|---|
| unittest | Проверка | Подтверждение |

## Изменения в проекте

- Тестовый отчёт.

## Финальный вердикт

**Закрыто:** да

**Не закрыто:** нет

**Что сломано:** ничего

**С чего продолжать:** завершить
""".format(department=department)


if __name__ == "__main__":
    unittest.main()
