.PHONY: help check fmt fix vet lint vuln test integration coverage build clean

# Project
BINARY_NAME := nautikube
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -ldflags="-s -w -X github.com/jorgegabrielti/nautikube/internal/cli.version=$(VERSION)"

## help: Show this help message
help:
	@echo "NautiKube — Kubernetes Cluster Diagnostic Tool"
	@echo ""
	@echo "Usage:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  make /' | column -t -s ':'
	@echo ""

## check: Run all quality gates (fmt, vet, lint, vuln, test, build)
check: fmt vet lint vuln test build
	@echo "✅ All checks passed"

## fmt: Check code formatting
fmt:
	@echo "→ Checking formatting..."
	@test -z "$$(gofmt -l .)" || (echo "❌ Files need formatting:"; gofmt -l .; exit 1)

## fix: Auto-fix formatting and lint issues
fix:
	@echo "→ Fixing formatting..."
	@gofmt -w .
	@echo "→ Fixing lint issues..."
	@golangci-lint run --fix 2>/dev/null || true
	@echo "✅ Fixes applied"

## vet: Run go vet
vet:
	@echo "→ Running go vet..."
	@go vet ./...

## lint: Run golangci-lint
lint:
	@echo "→ Running linters..."
	@golangci-lint run || (echo "⚠ Install golangci-lint: https://golangci-lint.run/"; exit 1)

## vuln: Check for known vulnerabilities
vuln:
	@echo "→ Checking vulnerabilities..."
	@govulncheck ./... 2>/dev/null || (echo "⚠ Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1)

## test: Run unit tests with race detector
test:
	@echo "→ Running tests..."
	@go test -race -cover ./...

## integration: Run integration tests (requires cluster)
integration:
	@echo "→ Running integration tests..."
	@go test -race -tags=integration ./...

## coverage: Generate HTML coverage report
coverage:
	@echo "→ Generating coverage report..."
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## build: Build the binary
build:
	@echo "→ Building $(BINARY_NAME) $(VERSION)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/nautikube
	@echo "✅ Built: ./$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@rm -f $(BINARY_NAME) coverage.txt coverage.html
	@echo "✅ Clean"
