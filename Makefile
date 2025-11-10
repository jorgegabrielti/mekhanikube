.PHONY: help build run test clean docker-build docker-up docker-down analyze health logs install dev

# Variables
BINARY_NAME=mekhanikube
GO_FILES=$(shell find . -name '*.go' -type f)
DOCKER_COMPOSE=docker-compose

# Colors for output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

## help: Display this help message
help:
	@echo "$(CYAN)Mekhanikube v2.0 - Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /' | column -t -s ':'
	@echo ""

## build: Build the Go binary locally
build:
	@echo "$(CYAN)Building Mekhanikube binary...$(NC)"
	@go build -o $(BINARY_NAME) ./cmd/mekhanikube
	@echo "$(GREEN)✓ Build complete: ./$(BINARY_NAME)$(NC)"

## run: Run Mekhanikube locally (requires cluster access)
run: build
	@echo "$(CYAN)Running Mekhanikube...$(NC)"
	@./$(BINARY_NAME) analyze --explain --language Portuguese

## test: Run Go tests
test:
	@echo "$(CYAN)Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "$(CYAN)Generating coverage report...$(NC)"
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

## lint: Run linters (requires golangci-lint)
lint:
	@echo "$(CYAN)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not found. Install: https://golangci-lint.run/$(NC)"; \
	fi

## fmt: Format Go code
fmt:
	@echo "$(CYAN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(CYAN)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## clean: Remove build artifacts
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f coverage.txt coverage.html
	@rm -rf vendor/
	@echo "$(GREEN)✓ Clean complete$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(CYAN)Building Docker image...$(NC)"
	@docker build -f configs/Dockerfile.mekhanikube -t mekhanikube:latest .
	@echo "$(GREEN)✓ Docker image built: mekhanikube:latest$(NC)"

## docker-up: Start all services (Mekhanikube v2 + Ollama)
docker-up:
	@echo "$(CYAN)Starting Mekhanikube v2 services...$(NC)"
	@$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "$(YELLOW)Run 'make analyze' to analyze your cluster$(NC)"

## docker-down: Stop all services
docker-down:
	@echo "$(CYAN)Stopping services...$(NC)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✓ Services stopped$(NC)"

## docker-restart: Restart all services
docker-restart: docker-down docker-up

## analyze: Run Mekhanikube analysis (Portuguese)
analyze:
	@echo "$(CYAN)Running Mekhanikube analysis...$(NC)"
	@docker exec mekhanikube $(BINARY_NAME) analyze --explain --language Portuguese

## analyze-en: Run Mekhanikube analysis (English)
analyze-en:
	@echo "$(CYAN)Running Mekhanikube analysis...$(NC)"
	@docker exec mekhanikube $(BINARY_NAME) analyze --explain --language English

## analyze-quick: Run quick analysis (no AI explanations)
analyze-quick:
	@echo "$(CYAN)Running quick analysis...$(NC)"
	@docker exec mekhanikube $(BINARY_NAME) analyze

## health: Check health of all services
health:
	@echo "$(CYAN)Checking services health...$(NC)"
	@echo ""
	@echo "$(GREEN)Ollama:$(NC)"
	@docker exec mekhanikube-ollama ollama list || echo "$(RED)✗ Ollama not running$(NC)"
	@echo ""
	@echo "$(GREEN)Mekhanikube:$(NC)"
	@docker exec mekhanikube $(BINARY_NAME) version || echo "$(RED)✗ Mekhanikube not running$(NC)"
	@echo ""
	@echo "$(GREEN)Kubernetes:$(NC)"
	@docker exec mekhanikube kubectl get nodes || echo "$(RED)✗ Cannot connect to cluster$(NC)"

## logs: Show logs from all services
logs:
	@echo "$(CYAN)Showing logs (Ctrl+C to exit)...$(NC)"
	@$(DOCKER_COMPOSE) logs -f

## logs-mekhanikube: Show Mekhanikube logs only
logs-mekhanikube:
	@docker logs -f mekhanikube

## logs-ollama: Show Ollama logs only
logs-ollama:
	@docker logs -f mekhanikube-ollama

## pull-model: Pull a specific Ollama model (usage: make pull-model MODEL=llama3.1:8b)
pull-model:
	@if [ -z "$(MODEL)" ]; then \
		echo "$(RED)Error: MODEL not specified$(NC)"; \
		echo "Usage: make pull-model MODEL=llama3.1:8b"; \
		exit 1; \
	fi
	@echo "$(CYAN)Pulling model $(MODEL)...$(NC)"
	@docker exec mekhanikube-ollama ollama pull $(MODEL)
	@echo "$(GREEN)✓ Model $(MODEL) pulled$(NC)"

## install: Install dependencies and setup
install:
	@echo "$(CYAN)Installing dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

## dev: Setup development environment
dev: install
	@echo "$(CYAN)Setting up development environment...$(NC)"
	@cp -n .env.example .env || true
	@echo "$(GREEN)✓ Development environment ready$(NC)"
	@echo "$(YELLOW)Edit .env file with your configuration$(NC)"

## version: Show Mekhanikube version
version:
	@docker exec mekhanikube $(BINARY_NAME) version 2>/dev/null || ./$(BINARY_NAME) version 2>/dev/null || echo "$(YELLOW)Build first: make build$(NC)"

## ps: Show running containers
ps:
	@$(DOCKER_COMPOSE) ps

## shell-mekhanikube: Open shell in Mekhanikube container
shell-mekhanikube:
	@docker exec -it mekhanikube /bin/sh

## shell-ollama: Open shell in Ollama container
shell-ollama:
	@docker exec -it mekhanikube-ollama /bin/sh

## prune: Clean up Docker resources (volumes, images)
prune:
	@echo "$(RED)⚠ This will remove all Mekhanikube Docker resources$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) down -v; \
		docker rmi mekhanikube:latest 2>/dev/null || true; \
		echo "$(GREEN)✓ Cleanup complete$(NC)"; \
	fi

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "$(GREEN)✓ All checks passed$(NC)"
