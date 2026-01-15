# Sancho CLI
# Makefile for building, testing, and linting

# Build variables
BINARY_NAME := sancho
BUILD_DIR := bin
CMD_PATH := ./cmd/sancho
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go variables
GOOS := darwin
GOARCH := arm64
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w \
	-X 'github.com/javiermolinar/sancho/internal/ui.Version=$(VERSION)' \
	-X 'github.com/javiermolinar/sancho/internal/ui.Commit=$(COMMIT)'

# Tools
GOLANGCI_LINT := golangci-lint
GOLANGCI_LINT_VERSION := v1.62.2

.PHONY: all build clean test lint lint-install fmt vet run help seed-integration test-integration

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

## all: Build the binary
all: build

## build: Build the binary for darwin/arm64
build: $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## build-dev: Build without optimizations for faster compilation
build-dev: $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME) (dev)"

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	$(GO) clean -cache -testcache

## test: Run all tests
test:
	$(GO) test -v -race -cover ./...

## test-short: Run tests without race detector
test-short:
	$(GO) test -v -cover ./...

## test-coverage: Run tests with coverage report
test-coverage: $(BUILD_DIR)
	$(GO) test -v -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

## lint: Run golangci-lint
lint:
	$(GOLANGCI_LINT) run ./...

## lint-fix: Run golangci-lint with auto-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

## lint-install: Install golangci-lint
lint-install:
	@which $(GOLANGCI_LINT) > /dev/null || \
		(echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))

## fmt: Format code
fmt:
	$(GO) fmt ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	$(GO) vet ./...

## mod: Tidy go modules
mod:
	$(GO) mod tidy
	$(GO) mod verify

## run: Build and run the application
run: build-dev
	./$(BUILD_DIR)/$(BINARY_NAME)

## check: Run fmt, vet, lint, and test
check: fmt vet lint test
	@echo "All checks passed"

## install: Install the binary to GOPATH/bin
install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(CMD_PATH)
	@echo "Installed $(BINARY_NAME) to $$(go env GOPATH)/bin"

## seed-integration: Build and create integration test database
seed-integration: build
	./integration/seed.sh

## test-integration: Run integration tests
test-integration: seed-integration
	$(GO) test -v ./integration/...
