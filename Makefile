.PHONY: help dev prod build test clean run stop

PROJECT_NAME := golang-microservices
DOCKER_COMPOSE_DEV := docker-compose.dev.yml
DOCKER_COMPOSE_PROD := docker-compose.prod.yml

help: ## Show help
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

dev: ## Start development environment
    docker-compose -f $(DOCKER_COMPOSE_DEV) up -d

dev-logs: ## Show development logs
    docker-compose -f $(DOCKER_COMPOSE_DEV) logs -f

dev-stop: ## Stop development environment
    docker-compose -f $(DOCKER_COMPOSE_DEV) down

prod: ## Start production environment
    docker-compose -f $(DOCKER_COMPOSE_PROD) up -d

prod-stop: ## Stop production environment
    docker-compose -f $(DOCKER_COMPOSE_PROD) down

build: ## Build all services
    @echo "Building services..."
    cd services/api-gateway && go build -o api-gateway ./cmd/
    cd services/user-service && go build -o user-service ./cmd/

docker-build: ## Build Docker images
    docker-compose -f $(DOCKER_COMPOSE_DEV) build

run-gateway: ## Run API Gateway locally
    cd services/api-gateway && go run ./cmd/

run-user: ## Run User Service locally
    cd services/user-service && go run ./cmd/

test: ## Run tests
    go test ./... -v

test-cover: ## Run tests with coverage
    go test ./... -v -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html

deps: ## Download dependencies
    go mod download
    go mod tidy

fmt: ## Format code
    go fmt ./...

clean: ## Clean build artifacts
    rm -f services/*/api-gateway
    rm -f services/*/user-service
    rm -f coverage.out coverage.html

setup: ## Setup development environment
    go mod download
    @echo "Development environment ready!"