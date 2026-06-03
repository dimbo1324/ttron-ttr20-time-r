.DEFAULT_GOAL := help

BIN_DIR := bin
PROTO_FILES := proto/ft12/v1/common.proto proto/ft12/v1/emulator.proto proto/ft12/v1/gateway.proto

.PHONY: help fmt test build build-client build-emulator build-gateway build-cli run-emulator run-client run-gateway proto clean

help:
	@echo "Targets:"
	@echo "  fmt             go fmt ./..."
	@echo "  test            go test ./..."
	@echo "  build           go build ./..."
	@echo "  build-client    build bin/ft12-client"
	@echo "  build-emulator  build bin/ft12-emulator"
	@echo "  build-gateway   build bin/ft12-gateway"
	@echo "  build-cli       build bin/ft12-cli"
	@echo "  run-emulator    run emulator on default address"
	@echo "  run-client      run client against default emulator"
	@echo "  run-gateway     run gateway against default emulator"
	@echo "  proto           generate protobuf/gRPC Go code"
	@echo "  clean           remove build output"

fmt:
	go fmt ./...

test:
	go test ./...

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

run-emulator:
	go run ./cmd/ft12-emulator

run-client:
	go run ./cmd/ft12-client

run-gateway:
	go run ./cmd/ft12-gateway

proto:
	protoc --go_out=. --go_opt=module=github.com/dimbo1324/ttron-ttr20-time-r --go-grpc_out=. --go-grpc_opt=module=github.com/dimbo1324/ttron-ttr20-time-r $(PROTO_FILES)

clean:
	rm -rf $(BIN_DIR)
