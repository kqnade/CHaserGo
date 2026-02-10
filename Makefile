.PHONY: all build test clean install fmt lint integration-test mapgen help

# デフォルトターゲット
all: build

# ヘルプ
help:
	@echo "CHaserGo Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make build            - Build all binaries"
	@echo "  make test             - Run unit tests"
	@echo "  make integration-test - Run integration tests"
	@echo "  make mapgen           - Generate sample maps"
	@echo "  make install          - Install all commands"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Run linter"
	@echo "  make clean            - Clean build artifacts"

# ビルド
build: build-server build-mapgen build-examples

build-server:
	@echo "Building chaser-server..."
	@go build -o bin/chaser-server ./cmd/chaser-server

build-mapgen:
	@echo "Building chaser-mapgen..."
	@go build -o bin/chaser-mapgen ./cmd/chaser-mapgen

build-examples:
	@echo "Building examples..."
	@go build -o bin/test1 ./examples/test1
	@go build -o bin/test2 ./examples/test2
	@go build -o bin/test3 ./examples/test3

# テスト
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

# 統合テスト
integration-test: build mapgen
	@echo "Running integration tests..."
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File scripts/integration-test.ps1
else
	@bash scripts/integration-test.sh
endif

# マップ生成
mapgen: build-mapgen
	@echo "Generating test maps..."
	@./bin/chaser-mapgen -b 5 -i 5 -o testdata 1

# インストール
install:
	@echo "Installing commands..."
	@go install ./cmd/chaser-server
	@go install ./cmd/chaser-mapgen

# フォーマット
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint
lint:
	@echo "Running linter..."
	@golangci-lint run

# クリーンアップ
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf generated_map/
	@rm -f coverage.txt coverage.html
	@rm -f chaser.dump

# 開発環境セットアップ
setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Done!"
