#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
runtime_dir="${TMPDIR:-/tmp}/llm-homework-local"

printf '%s\n' "--- api ---"
docker compose -f "$project_root/backend/docker-compose.yml" logs --tail 100 api
printf '%s\n' "--- frontend ---"
if [ -f "$runtime_dir/frontend.log" ]; then
	tail -n 100 "$runtime_dir/frontend.log"
else
	printf '%s\n' "Лог пока отсутствует."
fi
