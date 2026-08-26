#!/usr/bin/env python3
"""Validate new-format session reports across every project department."""

import argparse
import re
import sys
from pathlib import Path


REPORT_DIRECTORIES = (
    ("root", Path("sessions")),
    ("backend", Path("backend/sessions")),
    ("frontend", Path("frontend/sessions")),
    ("shared", Path("shared/sessions")),
)
REQUIRED_SECTIONS = (
    "Промпты",
    "Размышления",
    "Использованные инструменты",
    "Изменения в проекте",
    "Финальный вердикт",
)
REQUIRED_VERDICTS = ("Закрыто", "Не закрыто", "Что сломано", "С чего продолжать")


def parse_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument("--project-root", default=Path.cwd())
    return parser.parse_args()


def remove_comments(content):
    return re.sub(r"<!--.*?-->", "", content, flags=re.DOTALL)


def section_content(content, title):
    match = re.search(
        r"^## " + re.escape(title) + r"\s*$\n?(.*?)(?=^## |\Z)",
        content,
        flags=re.MULTILINE | re.DOTALL,
    )
    return None if match is None else match.group(1).strip()


def validate_report(report, department):
    raw_content = report.read_text(encoding="utf-8")
    if "<!-- hook:" not in raw_content:
        return []

    content = remove_comments(raw_content)
    errors = []
    if not re.search(r"^# Сессия: \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\s*$", content, re.MULTILINE):
        errors.append("нет заголовка сессии с датой и временем")
    if not re.search(r"^- \*\*Агент:\*\* (?!Не указан\s*$).+", content, re.MULTILINE):
        errors.append("не указаны агент и модель")
    if "- **Отдел:** " + department not in content:
        errors.append("отдел не соответствует каталогу " + department)
    if "<!-- hook: session.completed -->" in raw_content and "- **Статус:** завершена" not in content:
        errors.append("закрытая сессия не имеет статуса «завершена»")
    for title in REQUIRED_SECTIONS:
        if section_content(content, title) is None:
            errors.append("отсутствует раздел «" + title + "»")
    prompts = section_content(content, "Промпты") or ""
    if "### <USER PROMPT>" not in prompts:
        errors.append("нет фактического <USER PROMPT>")
    if "### <AGENT ANSWER>" not in prompts:
        errors.append("нет фактического <AGENT ANSWER>")
    tools = section_content(content, "Использованные инструменты") or ""
    if "| Инструмент | Действие | Зачем |" not in tools:
        errors.append("нет таблицы использованных инструментов")
    changes = section_content(content, "Изменения в проекте") or ""
    if not changes:
        errors.append("не заполнен раздел изменений проекта")
    verdict = section_content(content, "Финальный вердикт") or ""
    for title in REQUIRED_VERDICTS:
        if not re.search(r"^\*\*" + re.escape(title) + r":\*\*\s*\S+", verdict, re.MULTILINE):
            errors.append("не заполнен вердикт «" + title + "»")
    return errors


def validate_reports(project_root):
    failures = []
    for department, relative_directory in REPORT_DIRECTORIES:
        directory = project_root / relative_directory
        if not directory.is_dir():
            continue
        for report in sorted(directory.glob("session-*.md")):
            for error in validate_report(report, department):
                failures.append(str(report) + ": " + error)
    return failures


def main():
    arguments = parse_arguments()
    failures = validate_reports(Path(arguments.project_root).resolve())
    if failures:
        print("Invalid session reports:", file=sys.stderr)
        print("\n".join(failures), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
