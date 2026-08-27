#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
runtime_dir="${TMPDIR:-/tmp}/llm-homework-local"

for service in frontend; do
	pid_file="$runtime_dir/${service}.pid"
	if [ -f "$pid_file" ]; then
		pid=$(cat "$pid_file")
		if kill -0 "$pid" 2>/dev/null; then
			kill "$pid"
		fi
		rm -f "$pid_file"
	fi
done

docker compose -f "$project_root/backend/docker-compose.yml" stop api postgres
rm -f "$runtime_dir/api.log" "$runtime_dir/frontend.log"
rmdir "$runtime_dir" 2>/dev/null || true
