#!/bin/sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
runtime_dir="${TMPDIR:-/tmp}/llm-homework-local"
api_port="${API_PORT:-8080}"
frontend_port="${FRONTEND_PORT:-4173}"
frontend_origins="${FRONTEND_ORIGINS:-http://localhost:${frontend_port}}"

mkdir -p "$runtime_dir"

if [ -f "$runtime_dir/frontend.pid" ]; then
	frontend_pid=$(cat "$runtime_dir/frontend.pid")
	if kill -0 "$frontend_pid" 2>/dev/null; then
		printf '%s\n' "Проект уже запущен. Используйте 'make down' перед повторным запуском." >&2
		exit 1
	fi
	rm -f "$runtime_dir/frontend.pid"
fi

API_URL="http://localhost:${api_port}" make -C "$project_root" build
FRONTEND_ORIGINS="$frontend_origins" API_PORT="$api_port" \
	docker compose -f "$project_root/backend/docker-compose.yml" up -d --build api
docker compose -f "$project_root/backend/docker-compose.yml" exec -T api /migrate \
	-source file:///migrations \
	-database 'postgres://app:local-dev-password@postgres:5432/llm_homework?sslmode=disable' up

(
	cd "$project_root/frontend"
	npm run preview -- --host localhost --port "$frontend_port"
) >"$runtime_dir/frontend.log" 2>&1 &
printf '%s\n' "$!" >"$runtime_dir/frontend.pid"

printf '%s\n' "Backend:  http://localhost:${api_port}"
printf '%s\n' "Frontend: http://localhost:${frontend_port}"
printf '%s\n' "Логи:     make logs"
