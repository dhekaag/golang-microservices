package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dhekaag/golang-microservices/shared/pkg/errors"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
)

// Response writer wrapper
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: 200}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// Logging middleware
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create request context with IDs
			ctx, requestID := logger.GetOrCreateRequestID(r.Context())
			ctx, correlationID := logger.GetOrCreateCorrelationID(ctx)
			r = r.WithContext(ctx)

			// Wrap response writer
			wrapped := newResponseWriter(w)

			// Set headers
			wrapped.Header().Set("X-Request-ID", requestID)
			wrapped.Header().Set("X-Correlation-ID", correlationID)

			// Log request start
			logger.Info(ctx, "Request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", getClientIP(r),
			)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			logger.HTTPRequest(ctx, r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// Recovery middleware
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)

					// Log panic
					logger.Error(r.Context(), "Panic recovered",
						"error", fmt.Sprintf("%v", err),
						"stack", string(stack[:length]),
						"path", r.URL.Path,
					)

					// Return error response
					appErr := errors.NewInternalServerError("Internal server error", fmt.Errorf("%v", err))
					errors.WriteErrorResponse(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan bool, 1)
			go func() {
				next.ServeHTTP(w, r)
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				logger.Warn(r.Context(), "Request timeout", "timeout", timeout.String())
				appErr := errors.NewRequestTimeoutError("Request timeout", ctx.Err())
				errors.WriteErrorResponse(w, appErr)
			}
		})
	}
}

// Rate limiting middleware (simplified)
type RateLimiter struct {
	requests map[string][]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(clientIP string, maxRequests int, window time.Duration) bool {
	now := time.Now()

	// Clean old requests
	if requests, exists := rl.requests[clientIP]; exists {
		var validRequests []time.Time
		for _, req := range requests {
			if now.Sub(req) < window {
				validRequests = append(validRequests, req)
			}
		}
		rl.requests[clientIP] = validRequests
	}

	// Check if under limit
	if len(rl.requests[clientIP]) >= maxRequests {
		return false
	}

	// Add current request
	rl.requests[clientIP] = append(rl.requests[clientIP], now)
	return true
}

func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.Allow(clientIP, maxRequests, window) {
				logger.Warn(r.Context(), "Rate limit exceeded", "client_ip", clientIP)
				appErr := errors.NewTooManyRequestsError("Rate limit exceeded", nil)
				errors.WriteErrorResponse(w, appErr)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Simple health check middleware
func HealthCheck(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == path {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Utility function
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// Middleware chain helper
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
