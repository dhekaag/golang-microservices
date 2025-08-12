# Golang Microservices

A simple microservices architecture built with Go, featuring API Gateway and User Service.

## Architecture

```
Client → API Gateway (8080) → User Service (8081)
```

## Tech Stack

### Backend

- **[Go 1.24.6](https://golang.org/)** - Primary programming language
- **[GORM](https://gorm.io/)** - ORM library for database operations
- **[MySQL 8.0](https://www.mysql.com/)** - Primary database
- **[Redis](https://redis.io/)** - Session storage and caching
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** - Password hashing

### HTTP & Networking

- **[net/http](https://pkg.go.dev/net/http)** - Built-in HTTP server
- **[httputil](https://pkg.go.dev/net/http/httputil)** - Reverse proxy for API Gateway

### Data & Validation

- **[go-playground/validator](https://github.com/go-playground/validator)** - Struct validation
- **[google/uuid](https://github.com/google/uuid)** - UUID generation
- **[joho/godotenv](https://github.com/joho/godotenv)** - Environment variable loading

### Development & Tools

- **[Docker](https://www.docker.com/)** - Containerization
- **[Docker Compose](https://docs.docker.com/compose/)** - Multi-container orchestration
- **[Make](https://www.gnu.org/software/make/)** - Build automation
- **[Go Workspace](https://go.dev/ref/mod#workspaces)** - Multi-module development

### Architecture Patterns

- **Microservices** - Service-oriented architecture
- **API Gateway** - Single entry point pattern
- **Repository Pattern** - Data access abstraction
- **Middleware Stack** - Request/response processing
- **Structured Logging** - Centralized logging with correlation IDs

### DevOps & Deployment

- **Multi-stage Docker builds** - Optimized container images
- **Health checks** - Service monitoring
- **Graceful shutdown** - Clean resource cleanup
- **Environment-based configuration** - Flexible deployment

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

## Docker

```bash
# Build images
docker build -f services/api-gateway/Dockerfile -t api-gateway .
docker build -f services/user-service/Dockerfile -t user-service .

# Run with docker-compose
docker-compose -f deployments/docker-compose.dev.yml up
```

## Project Structure

```
golang-microservices/
├── deployments/                    # Docker compose files
├── services/
│   ├── api-gateway/               # API Gateway service
│   └── user-service/              # User management service
├── shared/                        # Shared libraries
│   └── pkg/
│       ├── database/              # Database utilities
│       ├── logger/                # Structured logging
│       ├── middleware/            # HTTP middlewares
│       └── utils/                 # Common utilities
├── go.work                        # Go workspace
└── Makefile                       # Build automation
```
