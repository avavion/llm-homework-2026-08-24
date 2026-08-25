#!/usr/bin/env python3
"""Create and finalize portable Markdown session reports without blocking hooks."""

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path


REPORT_DIRECTORIES = (
    ("root", Path("sessions")),
    ("backend", Path("backend/sessions")),
    ("frontend", Path("frontend/sessions")),
    ("shared", Path("shared/sessions")),
)


class NonBlockingArgumentParser(argparse.ArgumentParser):
    """Turn CLI usage errors into normal hook errors instead of exit code 2."""

    def error(self, message):
        raise ValueError(message)


def parse_arguments():
    parser = NonBlockingArgumentParser()
    parser.add_argument("event", choices=("start", "end"))
    parser.add_argument("--project-root")
    parser.add_argument("--cwd")
    parser.add_argument("--session-id")
    parser.add_argument("--timestamp")
    return parser.parse_args()


def read_payload():
    if sys.stdin.isatty():
        return {}
    try:
        value = sys.stdin.read().strip()
        return json.loads(value) if value else {}
    except (json.JSONDecodeError, OSError):
        return {}


def report_directories(project_root):
    return [project_root / relative_path for _, relative_path in REPORT_DIRECTORIES]


def route_report_directory(project_root, working_directory):
    if working_directory == project_root:
        return "root", project_root / "sessions"
    for department in ("backend", "frontend", "shared"):
        directory = project_root / department
        if working_directory == directory or directory in working_directory.parents:
            return department, directory / "sessions"
    raise ValueError("working directory is outside the project departments")


def find_report(project_root, session_id):
    expected_line = "session_id: " + session_id
    for directory in report_directories(project_root):
        if not directory.is_dir():
            continue
        for report in sorted(directory.glob("session-*.md")):
            try:
                if expected_line in report.read_text(encoding="utf-8").splitlines():
                    return report
            except OSError:
                continue
    return None


def report_text(session_id, department, timestamp):
    return """---
hook: session.started
session_id: {session_id}
department: {department}
started_at: {timestamp}
completed_at:
---

## Запрос

## Действия

## Источники

## Созданные артефакты

## Выводы

## Риски

## Самокритика

## Следующий шаг
""".format(session_id=session_id, department=department, timestamp=timestamp)


def finalize_report(report, timestamp):
    lines = report.read_text(encoding="utf-8").splitlines(keepends=True)
    if not lines or lines[0].strip() != "---":
        raise ValueError("report has no YAML frontmatter: " + str(report))
    for index in range(1, len(lines)):
        if lines[index].strip() == "---":
            break
        if lines[index].startswith("hook:"):
            lines[index] = "hook: session.completed\n"
        elif lines[index].startswith("completed_at:"):
            lines[index] = "completed_at: " + timestamp + "\n"
    else:
        raise ValueError("report has unterminated YAML frontmatter: " + str(report))
    report.write_text("".join(lines), encoding="utf-8")


def run():
    arguments = parse_arguments()
    payload = read_payload()
    project_root = Path(arguments.project_root or payload.get("project_root") or Path.cwd()).resolve()
    working_directory = Path(arguments.cwd or payload.get("cwd") or Path.cwd()).resolve()
    session_id = arguments.session_id or payload.get("session_id")
    if not session_id:
        raise ValueError("session ID is required")

    existing_report = find_report(project_root, session_id)
    if arguments.event == "end":
        if existing_report:
            finalize_report(existing_report, datetime.now().astimezone().isoformat(timespec="seconds"))
        return

    if existing_report:
        return

    department, directory = route_report_directory(project_root, working_directory)
    timestamp = arguments.timestamp or datetime.now().astimezone().isoformat(timespec="seconds")
    filename_stem = "session-" + datetime.fromisoformat(timestamp).strftime("%Y-%m-%d-%H%M%S")
    directory.mkdir(parents=True, exist_ok=True)
    report = directory / (filename_stem + ".md")
    suffix = 2
    while report.exists():
        report = directory / (filename_stem + "-" + str(suffix) + ".md")
        suffix += 1
    report.write_text(report_text(session_id, department, timestamp), encoding="utf-8")


def main():
    try:
        run()
    except Exception as error:  # Hooks must never block the caller.
        print("session hook: " + str(error), file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
