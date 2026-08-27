BACKEND_DIR := backend
FRONTEND_DIR := frontend
BIN_DIR := bin
API_BINARY := $(BIN_DIR)/api
API_URL ?= http://localhost:8080

.PHONY: build frontend-build backend-build up down logs

build: frontend-build backend-build

frontend-build:
	npm --prefix $(FRONTEND_DIR) ci
	VITE_API_URL=$(API_URL) npm --prefix $(FRONTEND_DIR) run build

backend-build:
	mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && go build -trimpath -o ../$(API_BINARY) ./cmd/api

up:
	./scripts/local-up.sh

down:
	./scripts/local-down.sh

logs:
	./scripts/local-logs.sh
