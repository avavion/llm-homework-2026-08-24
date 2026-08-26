COMPOSE ?= docker compose
GO ?= go
BIN_DIR ?= bin
API_PACKAGE ?= ./cmd/api
INTEGRATION_COMPOSE_PROJECT ?= llm-homework-backend-integration

.PHONY: up down logs run migrate-up test test-integration build-linux-amd64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

run: up

migrate-up:
	$(COMPOSE) exec -T api sh -ec '/migrate -source file:///migrations -database "$$DATABASE_URL" up'

test:
	$(GO) test ./...

test-integration:
	@set -eu; \
	cleanup() { $(COMPOSE) -p $(INTEGRATION_COMPOSE_PROJECT) --profile test down -v --remove-orphans; }; \
	trap cleanup EXIT HUP INT TERM; \
	POSTGRES_DB=llm_homework \
	POSTGRES_USER=app \
	POSTGRES_PASSWORD=local-dev-password \
	DATABASE_URL='postgres://app:local-dev-password@postgres:5432/llm_homework?sslmode=disable' \
	$(COMPOSE) -p $(INTEGRATION_COMPOSE_PROJECT) --profile test up --build --abort-on-container-exit --exit-code-from auth-integration-test auth-integration-test

build-linux-amd64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o $(BIN_DIR)/api-linux-amd64 $(API_PACKAGE)

build-darwin-amd64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -o $(BIN_DIR)/api-darwin-amd64 $(API_PACKAGE)

build-darwin-arm64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -o $(BIN_DIR)/api-darwin-arm64 $(API_PACKAGE)

build-windows-amd64:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o $(BIN_DIR)/api-windows-amd64.exe $(API_PACKAGE)
