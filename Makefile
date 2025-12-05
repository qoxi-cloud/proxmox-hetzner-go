# Makefile for proxmox-hetzner-go

# Build variables
BINARY_NAME := pve-install
BUILD_DIR := build
CMD_DIR := ./cmd/pve-install
VERSION_PKG := github.com/qoxi-cloud/proxmox-hetzner-go/pkg/version

# Version information (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -ldflags "-s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).Date=$(BUILD_DATE)"

# Go commands
GO := go
GOFMT := gofmt
GOLINT := golangci-lint

.PHONY: all build build-linux test test-coverage clean lint fmt help

# Default target
all: build

## build: Build for current platform
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

## build-linux: Cross-compile for Linux AMD64
build-linux:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)

## test: Run tests
test:
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## lint: Run golangci-lint
lint:
	$(GOLINT) run ./...

## fmt: Format code
fmt:
	$(GOFMT) -s -w .

## help: Show available targets
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -E 's/## /  /'
