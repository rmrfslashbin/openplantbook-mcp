BINARY := openplantbook-mcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

.PHONY: build test lint clean build-all install help

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

.DEFAULT_GOAL := build
