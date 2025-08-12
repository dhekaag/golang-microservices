# User Service

Microservice for user management and authentication.

## Features

- User registration and authentication
- Password hashing with bcrypt
- Profile management
- MySQL database with GORM
- Health monitoring

## Endpoints

### Public

- `POST /auth/register` - Register new user
- `POST /auth/login` - User login

### Authenticated

- `GET /users/{id}` - Get user by ID
- `PUT /users/{id}` - Update user profile
- `PUT /users/{id}/change-password` - Change password

### Health

- `GET /health` - Service health check

## Configuration

```env
PORT=8081
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=user_service
```

## Development

```bash
# Start MySQL
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0

# Run service
go run cmd/main.go

# Test
curl http://localhost:8081/health
```

## 🐳 Docker

```bash
# Build
docker build -f Dockerfile -t user-service .

# Run with MySQL
docker-compose up user-service mysql
```

## 🏗️ Architecture

```
Handler → Service → Repository → Database
```

- **Handler**: HTTP request handling
- **Service**: Business logic
- **Repository**: Data access layer
- **Domain**: Entity models
