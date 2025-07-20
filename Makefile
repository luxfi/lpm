# Copyright (C) 2019-2021, Lux Partners Limited. All rights reserved.
# See the file LICENSE for licensing terms.

.PHONY: all build test lint clean install release tag help

# Variables
BINARY_NAME := lpm
BUILD_DIR := ./build
MAIN_PATH := ./main
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# Default target
all: build

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^##/ { printf("  %-15s %s\n", substr($$0, 4, index($$0, ":")-4), substr($$0, index($$0, ":")+2)) }' $(MAKEFILE_LIST)

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-short: Run short tests
test-short:
	@echo "Running short tests..."
	@go test -v -short ./...

## coverage: Generate test coverage report
coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint:
	@echo "Running linters..."
	@./scripts/lint.sh

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## install: Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

## tidy: Tidy go modules
tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "Go modules tidied"

## verify: Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "Dependencies verified"

## release: Create a new release
release: clean build test
	@echo "Creating release $(VERSION)..."
	@mkdir -p dist
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Creating checksums..."
	@cd dist && sha256sum * > checksums.txt
	@echo "Release artifacts created in dist/"

## tag: Create and push a git tag
tag:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "Error: VERSION must be set (e.g., make tag VERSION=v1.0.0)"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag $(VERSION) created. Use 'git push origin $(VERSION)' to push the tag"

## docker: Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t luxfi/$(BINARY_NAME):$(VERSION) .
	@docker tag luxfi/$(BINARY_NAME):$(VERSION) luxfi/$(BINARY_NAME):latest
	@echo "Docker image built: luxfi/$(BINARY_NAME):$(VERSION)"

## run: Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## check: Run all checks (fmt, lint, test)
check: fmt lint test
	@echo "All checks passed"

## ci: Run continuous integration checks
ci: deps verify check
	@echo "CI checks complete"

# Print variables for debugging
print-%:
	@echo $* = $($*)