BINARY := openplantbook-mcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# MCPB configuration
MCPB_DIR := mcpb
MCPB_SERVER_DIR := $(MCPB_DIR)/server

.PHONY: build test lint clean build-all install help mcpb-validate mcpb-clean mcpb-pack-linux-amd64 mcpb-pack-linux-arm64 mcpb-pack-darwin-amd64 mcpb-pack-darwin-arm64 mcpb-pack-windows-amd64 mcpb-pack-all mcpb-info

build: ## Build binary for current platform
	@echo "Building $(BINARY) $(VERSION)..."
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

build-all: ## Build for multiple platforms
	@echo "Building $(BINARY) for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/$(BINARY)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	go vet ./...
	go fmt ./...
	@command -v staticcheck >/dev/null 2>&1 || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
	staticcheck ./...

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

install: build ## Install binary to $GOPATH/bin
	cp bin/$(BINARY) $(GOPATH)/bin/

run: build ## Build and run with example config
	@echo "Set OPENPLANTBOOK_API_KEY environment variable first"
	./bin/$(BINARY)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

mcpb-validate: ## Validate MCPB manifest
	@echo "Validating MCPB manifest..."
	@which mcpb > /dev/null || (echo "Error: mcpb not found. Install with: npm install -g @anthropic-ai/mcpb" && exit 1)
	mcpb validate $(MCPB_DIR)/manifest.json
	@echo "Manifest is valid!"

mcpb-clean: ## Clean MCPB artifacts
	@echo "Cleaning MCPB artifacts..."
	rm -rf $(MCPB_SERVER_DIR)/*
	rm -f bin/*.mcpb
	@echo "MCPB artifacts cleaned!"

mcpb-pack-linux-amd64: mcpb-validate ## Build MCPB package for Linux amd64
	@echo "Building MCPB package for Linux amd64..."
	@rm -rf $(MCPB_SERVER_DIR)/* && mkdir -p $(MCPB_SERVER_DIR) bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(MCPB_SERVER_DIR)/$(BINARY) ./cmd/$(BINARY)
	mcpb pack $(MCPB_DIR) bin/$(BINARY)-linux-amd64.mcpb
	@echo "Package created: bin/$(BINARY)-linux-amd64.mcpb"

mcpb-pack-linux-arm64: mcpb-validate ## Build MCPB package for Linux arm64
	@echo "Building MCPB package for Linux arm64..."
	@rm -rf $(MCPB_SERVER_DIR)/* && mkdir -p $(MCPB_SERVER_DIR) bin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(MCPB_SERVER_DIR)/$(BINARY) ./cmd/$(BINARY)
	mcpb pack $(MCPB_DIR) bin/$(BINARY)-linux-arm64.mcpb
	@echo "Package created: bin/$(BINARY)-linux-arm64.mcpb"

mcpb-pack-darwin-amd64: mcpb-validate ## Build MCPB package for macOS Intel
	@echo "Building MCPB package for macOS Intel..."
	@rm -rf $(MCPB_SERVER_DIR)/* && mkdir -p $(MCPB_SERVER_DIR) bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(MCPB_SERVER_DIR)/$(BINARY) ./cmd/$(BINARY)
	mcpb pack $(MCPB_DIR) bin/$(BINARY)-darwin-amd64.mcpb
	@echo "Package created: bin/$(BINARY)-darwin-amd64.mcpb"

mcpb-pack-darwin-arm64: mcpb-validate ## Build MCPB package for macOS Apple Silicon
	@echo "Building MCPB package for macOS Apple Silicon..."
	@rm -rf $(MCPB_SERVER_DIR)/* && mkdir -p $(MCPB_SERVER_DIR) bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(MCPB_SERVER_DIR)/$(BINARY) ./cmd/$(BINARY)
	mcpb pack $(MCPB_DIR) bin/$(BINARY)-darwin-arm64.mcpb
	@echo "Package created: bin/$(BINARY)-darwin-arm64.mcpb"

mcpb-pack-windows-amd64: mcpb-validate ## Build MCPB package for Windows
	@echo "Building MCPB package for Windows..."
	@rm -rf $(MCPB_SERVER_DIR)/* && mkdir -p $(MCPB_SERVER_DIR) bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(MCPB_SERVER_DIR)/$(BINARY).exe ./cmd/$(BINARY)
	mcpb pack $(MCPB_DIR) bin/$(BINARY)-windows-amd64.mcpb
	@echo "Package created: bin/$(BINARY)-windows-amd64.mcpb"

mcpb-pack-all: mcpb-clean mcpb-pack-linux-amd64 mcpb-pack-linux-arm64 mcpb-pack-darwin-amd64 mcpb-pack-darwin-arm64 mcpb-pack-windows-amd64 ## Build MCPB packages for all platforms
	@echo "All platform-specific MCPB packages created!"
	@ls -lh bin/*.mcpb

mcpb-info: ## Display MCPB package info (requires package to exist)
	@if [ -f bin/$(BINARY)-darwin-arm64.mcpb ]; then \
		echo "Package info for macOS Apple Silicon:"; \
		mcpb info bin/$(BINARY)-darwin-arm64.mcpb; \
	else \
		echo "No MCPB packages found. Run 'make mcpb-pack-all' first"; \
	fi

.DEFAULT_GOAL := build
