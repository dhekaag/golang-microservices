# Build stage
FROM golang:1.24.6-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy all
COPY . .

# Initialize workspace
RUN go work sync

# Change to user service directory
WORKDIR /app/services/user-service

# Download dependencies
RUN go mod download

# Verify dependencies
RUN go mod verify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/

# Final stage
FROM alpine:3.20

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/services/user-service/main .

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]