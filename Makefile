ROOT_DIR=${PWD}

.DEFAULT_GOAL := help

help:
	@echo "Type: make [rule]. Available options are:"
	@echo "  dev          - Start development environment"
	@echo "  dev-logs     - Show development logs"
	@echo "  dev-stop     - Stop development environment"
	@echo "  prod         - Start production environment"
	@echo "  prod-stop    - Stop production environment"
	@echo "  build        - Build all services"
	@echo "  docker-build - Build Docker images"
	@echo "  run-gateway  - Run API Gateway locally"
	@echo "  run-user-service - Run User Service locally"
	@echo "  test         - Run tests"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  setup        - Setup environment"

dev:
	cd deployment && docker compose -f docker-compose.dev.yml up --build

dev-logs:
	@echo "Showing development logs..."
	cd deployment && docker compose -f docker-compose.dev.yml logs -f

dev-stop:
	@echo "Stopping development environment..."
	cd deployment && docker compose -f docker-compose.dev.yml down

prod:
	@echo "Running production environment..."
	cd deployment && docker compose -f docker-compose.prod.yml up --build

prod-stop:
	@echo "Stopping production environment..."
	cd deployment && docker compose -f docker-compose.prod.yml down

build:
	cd services/api-gateway && go build -o api-gateway ./cmd/
	cd services/user-service && go build -o user-service ./cmd/

docker-build:
	@echo "Building Docker images..."
	cd deployment && docker compose -f docker-compose.prod.yml build

run-gateway:
	@echo "Running API Gateway..."
	cd services/api-gateway && go run ./cmd/

run-user-service:
	@echo "Running User Service..."
	cd services/user-service && go run ./cmd/

test:
	@echo "Running tests..."
	cd services/api-gateway && go test ./...
	cd services/user-service && go test ./...

deps:
	@echo "Installing dependencies..."
	cd services/api-gateway && go mod tidy
	cd services/user-service && go mod tidy

fmt:
	@echo "Formatting code..."
	cd services/api-gateway && go fmt ./...
	cd services/user-service && go fmt ./...

setup:
	@echo "Setting up the environment..."
	make deps
	make fmt
	make build