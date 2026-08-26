COMPOSE ?= docker compose
GO ?= go
BIN_DIR ?= bin
API_PACKAGE ?= ./cmd/api

.PHONY: up down logs run migrate-up test build-linux-amd64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

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
