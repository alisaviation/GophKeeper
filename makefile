.PHONY: all build build-client build-server build-all release clean test deps install help

# Настройки
APP_NAME := gophkeeper
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR := dist

# Флаги линковки
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.builtBy=make

# Платформы для сборки
CLIENT_PLATFORMS := windows-amd64 linux-amd64 linux-arm64 darwin-amd64 darwin-arm64
SERVER_PLATFORMS := linux-amd64 linux-arm64  # Сервер обычно на Linux

# Основные цели
all: clean test build

build: build-client build-server

build-client:
	@echo "Building client for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/client
	@chmod +x $(BUILD_DIR)/$(APP_NAME)
	@echo "Client built: $(BUILD_DIR)/$(APP_NAME)"

build-server:
	@echo "Building server for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-server ./cmd/server
	@chmod +x $(BUILD_DIR)/$(APP_NAME)-server
	@echo "Server built: $(BUILD_DIR)/$(APP_NAME)-server"

# Кроссплатформенная сборка клиента
build-client-all:
	@echo "Building client for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(CLIENT_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'-' -f1); \
		GOARCH=$$(echo $$platform | cut -d'-' -f2); \
		output=$(BUILD_DIR)/$(APP_NAME)-$$platform; \
		if [ "$$GOOS" = "windows" ]; then output=$$output.exe; fi; \
		echo "🔨 Building $$GOOS-$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o $$output ./cmd/client || exit 1; \
	done
	@echo "Client built for all platforms"

# Кроссплатформенная сборка сервера
build-server-all:
	@echo "Building server for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(SERVER_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'-' -f1); \
		GOARCH=$$(echo $$platform | cut -d'-' -f2); \
		output=$(BUILD_DIR)/$(APP_NAME)-server-$$platform; \
		echo "🔨 Building $$GOOS-$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -o $$output ./cmd/server || exit 1; \
	done
	@echo "Server built for all platforms"

# Сборка всех бинарников для всех платформ
build-all: build-client-all build-server-all
	@echo "All binaries built successfully!"

# Очистка
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)
	@echo "Build directory cleaned"

# Тестирование
test:
	@echo "Running tests..."
	go test ./... -v -race
	@echo "Tests completed"

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Установка зависимостей
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify
	go mod tidy
	@echo "Dependencies installed"

# Установка бинарников в систему (Linux/macOS)
install: build
	@echo "Installing binaries to /usr/local/bin/"
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	@sudo cp $(BUILD_DIR)/$(APP_NAME)-server /usr/local/bin/$(APP_NAME)-server
	@sudo chmod +x /usr/local/bin/$(APP_NAME) /usr/local/bin/$(APP_NAME)-server
	@echo "Binaries installed successfully"
	@echo "Usage:"
	@echo "  $(APP_NAME) --help"
	@echo "  $(APP_NAME)-server"

# Проверка установки
verify:
	@echo "Verifying installation..."
	@which $(APP_NAME) && echo "$(APP_NAME) found: $$(which $(APP_NAME))" || echo "$(APP_NAME) not found"
	@which $(APP_NAME)-server && echo "$(APP_NAME)-server found: $$(which $(APP_NAME)-server)" || echo "$(APP_NAME)-server not found"

# Помощь
help:
	@echo "GophKeeper Build System"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build client and server for current platform"
	@echo "  build-client   - Build only client for current platform"
	@echo "  build-server   - Build only server for current platform"
	@echo "  build-all      - Build client and server for all platforms"
	@echo "  build-client-all - Build client for all platforms"
	@echo "  build-server-all - Build server for all platforms"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean          - Clean build directory"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  deps           - Install dependencies"
	@echo "  install        - Install binaries to system"
	@echo "  verify         - Verify installation"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make install                  # Install to system"
	@echo ""
	@echo "Build info:"
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build dir: $(BUILD_DIR)"

# Показываем помощь по умолчанию
.DEFAULT_GOAL := help