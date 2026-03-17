.PHONY: build build-linux build-windows clean test bench cover run

BINARY_NAME=watcher
BUILD_DIR=build

build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags="-s -w" ./cmd/watcher

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux -ldflags="-s -w" ./cmd/watcher

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe -ldflags="-s -w" ./cmd/watcher

clean:
	@echo "Cleaning..."
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	go clean

test:
	@echo "Running tests..."
	go test ./...

bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

cover:
	@echo "Running tests with coverage..."
	go test -cover ./...

cover-html:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

run:
	go run ./cmd/watcher -config config.toml
