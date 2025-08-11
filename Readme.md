# Golang Microservices

A simple microservices architecture built with Go, featuring API Gateway and User Service.

## Architecture

```
Client → API Gateway (8080) → User Service (8081)
```

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose -f deployments/docker-compose.dev.yml up -d

# Check health
curl http://localhost:8080/health
```

### Manual Setup

```bash
# 1. Start MySQL and Redis
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0
docker run -d -p 6379:6379 redis:alpine

# 2. Start User Service
cd services/user-service
go run cmd/main.go

# 3. Start API Gateway
cd services/api-gateway
go run cmd/main.go
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get user info

### Health Check

- `GET /health` - API Gateway health
- `GET /api/v1/users/health` - User Service health

## Environment Variables

Copy `.env.example` to `.env` in each service directory and configure:

### API Gateway

```env
PORT=8080
USER_SERVICE_URL=http://localhost:8081
REDIS_ADDR=localhost:6379
```

### User Service

```env
PORT=8081
DB_HOST=localhost
DB_PASSWORD=password
DB_NAME=user_service
```

## Services

- **API Gateway** (Port 8080): Entry point, authentication, routing
- **User Service** (Port 8081): User management and authentication
- **Shared Package**: Common utilities (logger, middleware, database)

## Development

```bash
# Install dependencies
go work sync

# Run tests
make test

# Build all services
make build
```

## API Examples

### Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -c cookies.txt
```

### Get User Info

```bash
curl http://localhost:8080/api/v1/auth/me -b cookies.txt
```

## 🐳 Docker

```bash
# Build images
docker build -f services/api-gateway/Dockerfile -t api-gateway .
docker build -f services/user-service/Dockerfile -t user-service .

# Run with docker-compose
docker-compose -f deployments/docker-compose.dev.yml up
```
