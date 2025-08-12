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
	cd deployment && docker compose -f docker-compose.dev.yml logs -f

dev-stop:
	cd deployment && docker compose -f docker-compose.dev.yml down

prod:
	cd deployment && docker compose -f docker-compose.prod.yml up --build

prod-stop:
	cd deployment && docker compose -f docker-compose.prod.yml down

build:
	cd services/api-gateway && go build -o api-gateway ./cmd/
	cd services/user-service && go build -o user-service ./cmd/

docker-build:
	cd deployment && docker compose -f docker-compose.prod.yml build

run-gateway:
	cd services/api-gateway && go run ./cmd/

run-user-service:
	cd services/user-service && go run ./cmd/

test:
	cd services/api-gateway && go test ./...
	cd services/user-service && go test ./...

deps:
	cd services/api-gateway && go mod tidy
	cd services/user-service && go mod tidy

fmt:
	cd services/api-gateway && go fmt ./...
	cd services/user-service && go fmt ./...

setup:
	make deps
	make fmt
	make build