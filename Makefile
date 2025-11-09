.PHONY: help build up down restart logs status analyze install-model change-model clean test lint version

# Default model
MODEL ?= gemma:7b
NAMESPACE ?= default

help: ## Show this help message
	@echo "Mekhanikube - Kubernetes AI Mechanic"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build Docker images
	docker-compose build

up: ## Start all services
	docker-compose up -d
	@echo "Waiting for services to start..."
	@timeout /t 5 /nobreak > nul
	@echo "Services started! Use 'make install-model' to download AI model."

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs from all services
	docker-compose logs -f

logs-k8sgpt: ## Show k8sgpt logs only
	docker logs -f mekhanikube-k8sgpt

logs-ollama: ## Show ollama logs only
	docker logs -f mekhanikube-ollama

status: ## Check service status
	@docker-compose ps
	@echo ""
	@echo "K8sGPT configuration:"
	@docker exec mekhanikube-k8sgpt k8sgpt auth list

analyze: ## Run cluster analysis with AI explanations
	docker exec mekhanikube-k8sgpt k8sgpt analyze --explain

analyze-ns: ## Run analysis on specific namespace (use NAMESPACE=name)
	docker exec mekhanikube-k8sgpt k8sgpt analyze -n $(NAMESPACE) --explain

analyze-pods: ## Analyze only Pods
	docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Pod --explain

analyze-services: ## Analyze only Services
	docker exec mekhanikube-k8sgpt k8sgpt analyze --filter=Service --explain

filters: ## List available resource filters
	docker exec mekhanikube-k8sgpt k8sgpt filters list

install-model: ## Download AI model (default: gemma:7b, use MODEL=name to change)
	@echo "Downloading model: $(MODEL) (this may take several minutes)..."
	docker exec mekhanikube-ollama ollama pull $(MODEL)
	@echo "Model $(MODEL) installed successfully!"

list-models: ## List installed models
	docker exec mekhanikube-ollama ollama list

change-model: ## Change active AI model (use MODEL=name)
	@echo "Changing model to: $(MODEL)"
	docker exec mekhanikube-k8sgpt k8sgpt auth remove --backend ollama
	docker exec mekhanikube-k8sgpt k8sgpt auth add --backend ollama --model $(MODEL) --baseurl http://localhost:11434
	docker exec mekhanikube-k8sgpt k8sgpt auth default -p ollama
	@echo "Model changed to $(MODEL)"

shell-k8sgpt: ## Open shell in k8sgpt container
	docker exec -it mekhanikube-k8sgpt sh

shell-ollama: ## Open shell in ollama container
	docker exec -it mekhanikube-ollama sh

clean: ## Remove all containers, volumes, and images
	docker-compose down -v
	docker rmi mekhanikube-k8sgpt 2>/dev/null || true

clean-models: ## Remove all downloaded AI models
	docker volume rm mekhanikube-ollama-data

test: ## Run integration tests
	@echo "Running tests..."
	@./scripts/test.sh

lint: ## Lint configuration files
	@echo "Linting docker-compose.yml..."
	@docker-compose config -q
	@echo "Linting shell scripts..."
	@shellcheck configs/entrypoint.sh scripts/*.sh 2>/dev/null || echo "shellcheck not installed, skipping"

version: ## Show versions of components
	@echo "Mekhanikube version: $$(cat VERSION)"
	@echo "K8sGPT version: $$(docker exec mekhanikube-k8sgpt k8sgpt version 2>/dev/null || echo 'not running')"
	@echo "Ollama version: $$(docker exec mekhanikube-ollama ollama --version 2>/dev/null || echo 'not running')"

prune: ## Remove unused Docker resources
	docker system prune -f

health: ## Check health of all components
	@./scripts/healthcheck.sh

setup: build up install-model ## Complete setup (build, start, install model)
	@echo ""
	@echo "✓ Setup complete! Run 'make analyze' to start."
