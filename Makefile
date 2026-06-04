.DEFAULT_GOAL := help

BIN_DIR := bin
PROTO_FILES := proto/ft12/v1/common.proto proto/ft12/v1/emulator.proto proto/ft12/v1/gateway.proto

.PHONY: help fmt test test-race test-fuzz build build-client build-emulator build-gateway build-cli build-api build-healthcheck run-emulator run-client run-gateway run-api proto check-architecture check-doc-links release-check web-install web-dev web-build web-typecheck web-lint verify-web compose-config docker-build docker-up docker-down docker-logs docker-ps docker-smoke metrics-smoke ci-local verify clean

help:
	@echo "Targets:"
	@echo "  fmt             go fmt ./..."
	@echo "  test            go test ./..."
	@echo "  test-race       go test -race ./..."
	@echo "  test-fuzz       list documented fuzz entrypoint"
	@echo "  build           go build ./..."
	@echo "  build-client    build bin/ft12-client"
	@echo "  build-emulator  build bin/ft12-emulator"
	@echo "  build-gateway   build bin/ft12-gateway"
	@echo "  build-cli       build bin/ft12-cli"
	@echo "  build-api       build bin/ft12-api"
	@echo "  build-healthcheck build bin/ft12-healthcheck"
	@echo "  run-emulator    run emulator on default address"
	@echo "  run-client      run client against default emulator"
	@echo "  run-gateway     run gateway against default emulator"
	@echo "  run-api         run HTTP API on default address"
	@echo "  proto           generate protobuf/gRPC Go code"
	@echo "  check-architecture run dependency boundary checks"
	@echo "  check-doc-links validate local Markdown links"
	@echo "  release-check   run release-style local checks"
	@echo "  web-install     npm install in web/"
	@echo "  web-dev         run Vite dev server"
	@echo "  web-build       build web app"
	@echo "  web-typecheck   typecheck web app"
	@echo "  web-lint        lint web app"
	@echo "  compose-config  validate docker compose configuration"
	@echo "  docker-build    build compose images"
	@echo "  docker-up       run full compose stack"
	@echo "  docker-down     stop compose stack and remove volumes"
	@echo "  docker-smoke    run compose stack and smoke HTTP endpoints"
	@echo "  ci-local        run local equivalent quality gates"
	@echo "  verify          fmt, architecture checks, tests, build"
	@echo "  clean           remove build output"

fmt:
	go fmt ./...

test:
	go test ./...

test-race:
	go test -race ./...

test-fuzz:
	@echo "No mandatory fuzz corpus is configured yet. Use targeted go test -fuzz commands for protocol packages when adding fuzz tests."

build:
	go build ./...

build-client:
	go build -o $(BIN_DIR)/ft12-client ./cmd/ft12-client

build-emulator:
	go build -o $(BIN_DIR)/ft12-emulator ./cmd/ft12-emulator

build-gateway:
	go build -o $(BIN_DIR)/ft12-gateway ./cmd/ft12-gateway

build-cli:
	go build -o $(BIN_DIR)/ft12-cli ./cmd/ft12-cli

build-api:
	go build -o $(BIN_DIR)/ft12-api ./cmd/ft12-api

build-healthcheck:
	go build -o $(BIN_DIR)/ft12-healthcheck ./cmd/ft12-healthcheck

run-emulator:
	go run ./cmd/ft12-emulator

run-client:
	go run ./cmd/ft12-client

run-gateway:
	go run ./cmd/ft12-gateway

run-api:
	go run ./cmd/ft12-api

proto:
	protoc --go_out=. --go_opt=module=github.com/dimbo1324/ttron-ttr20-time-r --go-grpc_out=. --go-grpc_opt=module=github.com/dimbo1324/ttron-ttr20-time-r $(PROTO_FILES)

check-architecture:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-architecture.ps1
else
	sh scripts/check-architecture.sh
endif

check-doc-links:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-doc-links.ps1
else
	sh scripts/check-doc-links.sh
endif

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-typecheck:
	cd web && npm run typecheck

web-lint:
	cd web && npm run lint

compose-config:
	docker compose config

docker-build:
	docker compose build

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f

docker-ps:
	docker compose ps

docker-smoke:
	docker compose up -d --build
	docker compose ps
	docker compose exec -T ft12-api /app/ft12-healthcheck -url http://127.0.0.1:8080/health
	docker compose exec -T ft12-api /app/ft12-healthcheck -url http://127.0.0.1:8080/api/v1/ready
	docker compose exec -T ft12-api /app/ft12-healthcheck -url http://127.0.0.1:8080/api/v1/overview
	docker compose down -v

metrics-smoke:
	docker compose up -d --build ft12-api ft12-web
	docker compose exec -T ft12-api /app/ft12-healthcheck -url http://127.0.0.1:8080/metrics
	docker compose down -v

verify-web:
ifeq ($(OS),Windows_NT)
	@if where npm >NUL 2>NUL; then cd web && npm run typecheck && npm run build; else echo "npm not found; skipping web verification"; fi
else
	@if command -v npm >/dev/null 2>&1; then cd web && npm run typecheck && npm run build; else echo "npm not found; skipping web verification"; fi
endif

ci-local: fmt check-architecture test build web-typecheck web-lint web-build compose-config check-doc-links

release-check:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/release-check.ps1
else
	sh scripts/release-check.sh
endif

verify: fmt check-architecture test build verify-web

clean:
	rm -rf $(BIN_DIR)
