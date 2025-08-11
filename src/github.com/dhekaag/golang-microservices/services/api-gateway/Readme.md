# API Gateway Service

Entry point for all client requests. Handles authentication and routes to backend services.

## Features

- Request routing to downstream services
- Session-based authentication with Redis
- Request logging and CORS handling
- Health checks and graceful shutdown

## Endpoints

### Authentication

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/refresh` - Refresh session

### Proxy Routes

- `POST /api/v1/auth/register` → User Service
- `GET /api/v1/users/*` → User Service (authenticated)

### Health

- `GET /health` - Service health check

## Configuration

```env
PORT=8080
USER_SERVICE_URL=http://localhost:8081
REDIS_ADDR=localhost:6379
SESSION_TTL=24h
```

## Development

```bash
# Run locally
go run cmd/main.go

# Test endpoints
curl http://localhost:8080/health
```

## Docker

```bash
# Build
docker build -f Dockerfile -t api-gateway .

# Run
docker run -p 8080:8080 api-gateway
```

## Middleware Stack

1. Recovery - Panic recovery
2. Logging - Request/response logging
3. CORS - Cross-origin headers
4. Session Auth - Authentication
5. Security Headers - Security headers
6. Request Timeout - Timeout handling
