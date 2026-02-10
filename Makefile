# EdgeFlow Modular Build System
# Supports minimal, standard, and full profiles for different IoT boards

# ============================================================================
# Variables
# ============================================================================

BINARY_NAME=edgeflow
VERSION?=0.1.0
BUILD_DIR=bin
GO=go
GOPROXY?=direct
PROFILE?=standard

export GOPROXY

# Build profile configurations
ifeq ($(PROFILE),minimal)
	BUILD_TAGS=minimal
	LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.Profile=minimal"
	OPTIMIZATION=-trimpath
	TARGET_BOARDS=Pi Zero, BeagleBone (512MB RAM)
else ifeq ($(PROFILE),standard)
	BUILD_TAGS=standard,network,gpio,database
	LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION) -X main.Profile=standard"
	OPTIMIZATION=-trimpath
	TARGET_BOARDS=Pi 3/4, Orange Pi (1GB RAM)
else ifeq ($(PROFILE),full)
	BUILD_TAGS=full,network,gpio,database,messaging,ai,industrial,advanced
	LDFLAGS=-ldflags "-w -X main.Version=$(VERSION) -X main.Profile=full"
	OPTIMIZATION=
	TARGET_BOARDS=Pi 4/5, Jetson Nano (2GB+ RAM)
endif

# Colors for terminal output
GREEN=\033[0;32m
BLUE=\033[0;34m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m

# ============================================================================
# Help Target (Default)
# ============================================================================

.PHONY: help
help:
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║           EdgeFlow Modular Build System v$(VERSION)              ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)Build Profiles:$(NC)"
	@echo "  make build PROFILE=minimal   - Pi Zero, BeagleBone (10MB, 50MB RAM)"
	@echo "  make build PROFILE=standard  - Pi 3/4 (20MB, 200MB RAM) [default]"
	@echo "  make build PROFILE=full      - Pi 4/5, Jetson (35MB, 400MB RAM)"
	@echo ""
	@echo "$(GREEN)Platform Builds:$(NC)"
	@echo "  make build-pi-minimal        - Minimal for Pi Zero (ARM64)"
	@echo "  make build-pi-standard       - Standard for Pi 3/4 (ARM64)"
	@echo "  make build-pi-full           - Full for Pi 4/5 (ARM64)"
	@echo "  make build-all-profiles      - Build all profiles for all platforms"
	@echo ""
	@echo "$(GREEN)Development:$(NC)"
	@echo "  make run                     - Build and run (PROFILE=standard)"
	@echo "  make dev                     - Run with hot reload"
	@echo "  make test                    - Run all tests"
	@echo "  make lint                    - Run linter"
	@echo ""
	@echo "$(GREEN)Module Management:$(NC)"
	@echo "  make modules-list            - List available modules"
	@echo "  make modules-info            - Show module information"
	@echo ""
	@echo "$(GREEN)Frontend:$(NC)"
	@echo "  make frontend-build          - Build frontend"
	@echo "  make frontend-dev            - Run frontend dev server"
	@echo ""

.DEFAULT_GOAL := help

# ============================================================================
# Build Targets
# ============================================================================

.PHONY: build
build:
	@echo "$(GREEN)Building $(BINARY_NAME) [$(PROFILE) profile]...$(NC)"
	@echo "$(BLUE)Target boards: $(TARGET_BOARDS)$(NC)"
	@echo "$(BLUE)Build tags: $(BUILD_TAGS)$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -tags "$(BUILD_TAGS)" $(LDFLAGS) $(OPTIMIZATION) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/edgeflow
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: all
all: build

.PHONY: build-pi-minimal
build-pi-minimal:
	@echo "$(GREEN)Building minimal profile for Pi Zero (ARM64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build -tags "minimal" \
		-ldflags "-w -s -X main.Version=$(VERSION) -X main.Profile=minimal" \
		-trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-minimal-linux-arm64 ./cmd/edgeflow
	@echo "$(GREEN)✓ Minimal build complete$(NC)"

.PHONY: build-pi-standard
build-pi-standard:
	@echo "$(GREEN)Building standard profile for Pi 3/4 (ARM64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build -tags "standard,network,gpio,database" \
		-ldflags "-w -s -X main.Version=$(VERSION) -X main.Profile=standard" \
		-trimpath -o $(BUILD_DIR)/$(BINARY_NAME)-standard-linux-arm64 ./cmd/edgeflow
	@echo "$(GREEN)✓ Standard build complete$(NC)"

.PHONY: build-pi-full
build-pi-full:
	@echo "$(GREEN)Building full profile for Pi 4/5 (ARM64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build -tags "full,network,gpio,database,messaging,ai,industrial,advanced" \
		-ldflags "-w -X main.Version=$(VERSION) -X main.Profile=full" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-full-linux-arm64 ./cmd/edgeflow
	@echo "$(GREEN)✓ Full build complete$(NC)"

.PHONY: build-all-profiles
build-all-profiles: build-pi-minimal build-pi-standard build-pi-full
	@echo "$(GREEN)✓ All profiles built successfully$(NC)"

# Cross-platform builds
.PHONY: build-linux
build-linux:
	@echo "$(GREEN)Building for Linux (AMD64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -tags "$(BUILD_TAGS)" $(LDFLAGS) $(OPTIMIZATION) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/edgeflow
	@echo "$(GREEN)✓ Linux build complete$(NC)"

.PHONY: build-windows
build-windows:
	@echo "$(GREEN)Building for Windows (AMD64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -tags "$(BUILD_TAGS)" $(LDFLAGS) $(OPTIMIZATION) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/edgeflow
	@echo "$(GREEN)✓ Windows build complete$(NC)"

.PHONY: build-darwin
build-darwin:
	@echo "$(GREEN)Building for macOS (AMD64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build -tags "$(BUILD_TAGS)" $(LDFLAGS) $(OPTIMIZATION) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/edgeflow
	@echo "$(GREEN)✓ macOS build complete$(NC)"

.PHONY: build-all-platforms
build-all-platforms: build-linux build-windows build-darwin build-all-profiles
	@echo "$(GREEN)✓ All platforms built successfully$(NC)"

# Module-specific builds
.PHONY: build-module-core
build-module-core:
	@echo "$(GREEN)Building core module...$(NC)"
	$(GO) build -tags "core" -buildmode=plugin -o $(BUILD_DIR)/modules/core.so ./pkg/modules/core

.PHONY: build-module-network
build-module-network:
	@echo "$(GREEN)Building network module...$(NC)"
	$(GO) build -tags "network" -buildmode=plugin -o $(BUILD_DIR)/modules/network.so ./pkg/modules/network

.PHONY: build-module-gpio
build-module-gpio:
	@echo "$(GREEN)Building GPIO module...$(NC)"
	$(GO) build -tags "gpio" -buildmode=plugin -o $(BUILD_DIR)/modules/gpio.so ./pkg/modules/gpio

.PHONY: build-modules
build-modules: build-module-core build-module-network build-module-gpio
	@echo "$(GREEN)✓ All modules built successfully$(NC)"

# ============================================================================
# Module Management
# ============================================================================

.PHONY: modules-list
modules-list:
	@echo "$(BLUE)Available Modules:$(NC)"
	@echo ""
	@echo "$(GREEN)✓ core$(NC)       - Essential nodes (inject, debug, function) [REQUIRED]"
	@echo "$(YELLOW)○ network$(NC)    - HTTP, WebSocket, MQTT, TCP/UDP nodes"
	@echo "$(YELLOW)○ gpio$(NC)       - GPIO, I2C, SPI, PWM control"
	@echo "$(YELLOW)○ database$(NC)   - SQLite, PostgreSQL, MongoDB, InfluxDB"
	@echo "$(YELLOW)○ messaging$(NC)  - Telegram, Email, SMS, Webhook"
	@echo "$(YELLOW)○ ai$(NC)         - TensorFlow Lite, OpenCV, Image Recognition"
	@echo "$(YELLOW)○ industrial$(NC) - Modbus, OPC-UA, industrial protocols"
	@echo "$(YELLOW)○ advanced$(NC)   - Custom scripting, advanced transformations"
	@echo ""
	@echo "Current profile: $(PROFILE)"
	@echo "Enabled modules: $(BUILD_TAGS)"

.PHONY: modules-info
modules-info:
	@echo "$(BLUE)Module Resource Requirements:$(NC)"
	@echo ""
	@echo "Profile    | Binary Size | RAM Usage  | Modules"
	@echo "-----------|-------------|------------|----------------------------------"
	@echo "Minimal    | ~10MB       | ~50MB      | core"
	@echo "Standard   | ~20MB       | ~200MB     | core, network, gpio, database"
	@echo "Full       | ~35MB       | ~400MB     | all modules"

# ============================================================================
# Development
# ============================================================================

.PHONY: run
run: build
	@echo "$(GREEN)Running $(BINARY_NAME) [$(PROFILE) profile]...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: dev
dev:
	@echo "$(GREEN)Running in development mode...$(NC)"
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v -race -cover -tags "unit,$(BUILD_TAGS)" ./...

.PHONY: test-short
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GO) test -v -short -tags "unit" ./...

.PHONY: test-coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out -tags "unit,$(BUILD_TAGS)" ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

.PHONY: test-integration
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GO) test -v -tags "integration,$(BUILD_TAGS)" ./...

.PHONY: test-e2e
test-e2e:
	@echo "$(GREEN)Running end-to-end tests...$(NC)"
	$(GO) test -v -tags "e2e,$(BUILD_TAGS)" ./tests/e2e/...

.PHONY: test-security
test-security:
	@echo "$(GREEN)Running security tests...$(NC)"
	$(GO) test -v -tags "security" ./internal/security/... ./internal/api/middleware/...

.PHONY: test-nodes
test-nodes:
	@echo "$(GREEN)Running node tests...$(NC)"
	$(GO) test -v -tags "nodes,$(BUILD_TAGS)" ./pkg/nodes/...

.PHONY: test-network
test-network:
	@echo "$(GREEN)Running network node tests...$(NC)"
	$(GO) test -v -tags "network" ./pkg/nodes/network/...

.PHONY: test-gpio
test-gpio:
	@echo "$(GREEN)Running GPIO node tests...$(NC)"
	$(GO) test -v -tags "gpio" ./pkg/nodes/gpio/...

.PHONY: test-database
test-database:
	@echo "$(GREEN)Running database node tests...$(NC)"
	$(GO) test -v -tags "database" ./pkg/nodes/database/...

.PHONY: test-messaging
test-messaging:
	@echo "$(GREEN)Running messaging node tests...$(NC)"
	$(GO) test -v -tags "messaging" ./pkg/nodes/messaging/...

.PHONY: test-ai
test-ai:
	@echo "$(GREEN)Running AI node tests...$(NC)"
	$(GO) test -v -tags "ai" ./pkg/nodes/ai/...

.PHONY: test-core
test-core:
	@echo "$(GREEN)Running core node tests...$(NC)"
	$(GO) test -v -tags "core" ./pkg/nodes/core/...

.PHONY: test-engine
test-engine:
	@echo "$(GREEN)Running engine tests...$(NC)"
	$(GO) test -v ./internal/engine/...

.PHONY: test-health
test-health:
	@echo "$(GREEN)Running health check tests...$(NC)"
	$(GO) test -v ./internal/health/...

.PHONY: test-all
test-all: test test-integration test-security
	@echo "$(GREEN)All tests completed$(NC)"

.PHONY: bench
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem -tags "$(BUILD_TAGS)" ./...

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: lint
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

.PHONY: security
security:
	@echo "$(GREEN)Running security checks...$(NC)"
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec ./...

# ============================================================================
# Frontend
# ============================================================================

.PHONY: frontend-install
frontend-install:
	@echo "$(GREEN)Installing frontend dependencies...$(NC)"
	cd web && npm install

.PHONY: frontend-build
frontend-build:
	@echo "$(GREEN)Building frontend...$(NC)"
	cd web && npm run build

.PHONY: frontend-dev
frontend-dev:
	@echo "$(GREEN)Running frontend dev server...$(NC)"
	cd web && npm run dev

# ============================================================================
# Utilities
# ============================================================================

.PHONY: clean
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: deps
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

.PHONY: install
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"
