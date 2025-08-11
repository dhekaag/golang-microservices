package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/config"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/handler"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/middleware/gateway"
	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/proxy"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/dhekaag/golang-microservices/shared/pkg/middleware"
	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
)

type Router struct {
	serviceProxy *proxy.ServiceProxy
	authHandler  *handler.AuthHandler
	config       *config.Config
}

func NewRouter(
	serviceProxy *proxy.ServiceProxy,
	authHandler *handler.AuthHandler,
	config *config.Config,
) *Router {
	return &Router{
		serviceProxy: serviceProxy,
		authHandler:  authHandler,
		config:       config,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check routes (no authentication required)
	mux.HandleFunc("/health", r.handleHealthCheck)
	mux.HandleFunc("/health/ready", r.handleHealthCheck)
	mux.HandleFunc("/health/live", r.handleHealthCheck)

	// Authentication routes (handled by gateway)
	mux.HandleFunc("/api/v1/auth/login", r.authHandler.Login)
	mux.HandleFunc("/api/v1/auth/logout", r.authHandler.Logout)
	mux.HandleFunc("/api/v1/auth/me", r.authHandler.GetUserInfo)
	mux.HandleFunc("/api/v1/auth/refresh", r.authHandler.RefreshSession)
	mux.HandleFunc("/api/v1/auth/logout-all", r.authHandler.LogoutAllSessions)

	// Registration route (proxy to user service)
	mux.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			req.URL.Path = "/auth/register"
			r.serviceProxy.ProxyToService("user", w, req)
		} else {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Password reset routes (proxy to user service)
	mux.HandleFunc("/api/v1/auth/forgot-password", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			req.URL.Path = "/auth/forgot-password"
			r.serviceProxy.ProxyToService("user", w, req)
		} else {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/auth/reset-password", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			req.URL.Path = "/auth/reset-password"
			r.serviceProxy.ProxyToService("user", w, req)
		} else {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// User service routes
	mux.HandleFunc("/api/v1/users", r.handleUserRoutes)
	mux.HandleFunc("/api/v1/users/", r.handleUserRoutes)

	// Product service routes
	mux.HandleFunc("/api/v1/products", r.handleProductRoutes)
	mux.HandleFunc("/api/v1/products/", r.handleProductRoutes)
	mux.HandleFunc("/api/v1/categories", r.handleProductRoutes)
	mux.HandleFunc("/api/v1/categories/", r.handleProductRoutes)

	// Order service routes
	mux.HandleFunc("/api/v1/orders", r.handleOrderRoutes)
	mux.HandleFunc("/api/v1/orders/", r.handleOrderRoutes)
	mux.HandleFunc("/api/v1/cart", r.handleOrderRoutes)
	mux.HandleFunc("/api/v1/cart/", r.handleOrderRoutes)

	// Admin routes (protected)
	mux.HandleFunc("/api/v1/admin/", r.handleAdminRoutes)

	// File upload routes
	mux.HandleFunc("/api/v1/upload", r.handleUploadRoutes)
	mux.HandleFunc("/api/v1/upload/", r.handleUploadRoutes)

	// Webhook routes
	mux.HandleFunc("/api/v1/webhooks/", r.handleWebhookRoutes)

	// API documentation
	mux.HandleFunc("/docs", r.handleDocsRoutes)
	mux.HandleFunc("/docs/", r.handleDocsRoutes)

	// Apply global middlewares
	handler := r.applyMiddlewares(mux)

	return handler
}

func (r *Router) handleUserRoutes(w http.ResponseWriter, req *http.Request) {
	// Apply authentication middleware for protected routes
	protectedRoutes := []string{
		"/api/v1/users/profile",
		"/api/v1/users/change-password",
		"/api/v1/users/upload-avatar",
	}

	if r.isProtectedRoute(req.URL.Path, protectedRoutes) ||
		(req.URL.Path == "/api/v1/users" && req.Method != "POST") {
		if !r.isAuthenticated(req) {
			utils.SendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}
	}

	// Remove /api/v1 prefix and forward to user service
	newPath := strings.TrimPrefix(req.URL.Path, "/api/v1")
	req.URL.Path = newPath
	r.serviceProxy.ProxyToService("user", w, req)
}

func (r *Router) handleProductRoutes(w http.ResponseWriter, req *http.Request) {
	// Apply authentication middleware for write operations
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		if !r.isAuthenticated(req) {
			utils.SendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Check admin role for product management
		if !r.isAdmin(req) {
			utils.SendError(w, http.StatusForbidden, "Admin access required")
			return
		}
	}

	// Remove /api/v1 prefix and forward to product service
	newPath := strings.TrimPrefix(req.URL.Path, "/api/v1")
	req.URL.Path = newPath
	r.serviceProxy.ProxyToService("product", w, req)
}

func (r *Router) handleOrderRoutes(w http.ResponseWriter, req *http.Request) {
	// All order routes require authentication
	if !r.isAuthenticated(req) {
		utils.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Admin can access all orders
	adminRoutes := []string{
		"/api/v1/orders/admin",
		"/api/v1/orders/analytics",
		"/api/v1/orders/export",
	}

	if r.isProtectedRoute(req.URL.Path, adminRoutes) {
		if !r.isAdmin(req) {
			utils.SendError(w, http.StatusForbidden, "Admin access required")
			return
		}
	}

	// Remove /api/v1 prefix and forward to order service
	newPath := strings.TrimPrefix(req.URL.Path, "/api/v1")
	req.URL.Path = newPath
	r.serviceProxy.ProxyToService("order", w, req)
}

func (r *Router) handleAdminRoutes(w http.ResponseWriter, req *http.Request) {
	// Admin routes require authentication and admin role
	if !r.isAuthenticated(req) {
		utils.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if !r.isAdmin(req) {
		utils.SendError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Route to appropriate service based on path
	path := req.URL.Path
	switch {
	case strings.HasPrefix(path, "/api/v1/admin/users"):
		req.URL.Path = strings.TrimPrefix(path, "/api/v1/admin")
		r.serviceProxy.ProxyToService("user", w, req)
	case strings.HasPrefix(path, "/api/v1/admin/products"):
		req.URL.Path = strings.TrimPrefix(path, "/api/v1/admin")
		r.serviceProxy.ProxyToService("product", w, req)
	case strings.HasPrefix(path, "/api/v1/admin/orders"):
		req.URL.Path = strings.TrimPrefix(path, "/api/v1/admin")
		r.serviceProxy.ProxyToService("order", w, req)
	default:
		utils.SendError(w, http.StatusNotFound, "Admin endpoint not found")
	}
}

func (r *Router) handleUploadRoutes(w http.ResponseWriter, req *http.Request) {
	// File upload requires authentication
	if !r.isAuthenticated(req) {
		utils.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Route based on upload type
	uploadType := req.URL.Query().Get("type")
	switch uploadType {
	case "avatar", "profile":
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		r.serviceProxy.ProxyToService("user", w, req)
	case "product", "category":
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		r.serviceProxy.ProxyToService("product", w, req)
	default:
		utils.SendError(w, http.StatusBadRequest, "Invalid upload type")
	}
}

func (r *Router) handleWebhookRoutes(w http.ResponseWriter, req *http.Request) {
	// Webhook routes don't require authentication but should validate webhook signature
	path := req.URL.Path

	switch {
	case strings.HasPrefix(path, "/api/v1/webhooks/payment"):
		req.URL.Path = strings.TrimPrefix(path, "/api/v1")
		r.serviceProxy.ProxyToService("order", w, req)
	case strings.HasPrefix(path, "/api/v1/webhooks/notification"):
		req.URL.Path = strings.TrimPrefix(path, "/api/v1")
		r.serviceProxy.ProxyToService("user", w, req)
	default:
		utils.SendError(w, http.StatusNotFound, "Webhook endpoint not found")
	}
}

func (r *Router) handleDocsRoutes(w http.ResponseWriter, req *http.Request) {
	// Serve API documentation
	utils.SendSuccess(w, http.StatusOK, "API Documentation", map[string]string{
		"swagger": "/docs/swagger.json",
		"postman": "/docs/postman.json",
		"version": "v1.0.0",
	})
}

func (r *Router) handleHealthCheck(w http.ResponseWriter, req *http.Request) {
	utils.SendSuccess(w, http.StatusOK, "API Gateway is healthy", map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"services": map[string]bool{
			"user":    r.serviceProxy.IsServiceHealthy("user"),
			"product": r.serviceProxy.IsServiceHealthy("product"),
			"order":   r.serviceProxy.IsServiceHealthy("order"),
		},
	})
}

func (r *Router) applyMiddlewares(handler http.Handler) http.Handler {
	handler = middleware.Timeout(r.config.Server.RequestTimeout)(handler)

	// Security headers middleware
	handler = middleware.SecurityHeaders()(handler)

	// Request ID middleware
	handler = middleware.Chain(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := req.Context()

				// Get or create request ID
				ctx, requestID := logger.GetOrCreateRequestID(ctx)

				// Get or create correlation ID
				ctx, correlationID := logger.GetOrCreateCorrelationID(ctx)

				// Set headers for downstream services
				req.Header.Set("X-Request-ID", requestID)
				req.Header.Set("X-Correlation-ID", correlationID)

				// Set response headers
				w.Header().Set("X-Request-ID", requestID)
				w.Header().Set("X-Correlation-ID", correlationID)

				// Update request context
				req = req.WithContext(ctx)

				next.ServeHTTP(w, req)
			})
		},
	)(handler)

	// Rate limiting middleware
	// rateLimitConfig := gateway.RateLimitConfig{
	// 	RequestsPerMinute: r.config.RateLimit.RequestsPerMinute,
	// 	WindowSize:        r.config.RateLimit.WindowSize,
	// }
	// handler = func(next http.Handler) http.Handler {
	// 	return gateway.RateLimit(next, rateLimitConfig)
	// }(handler)

	// Session authentication middleware
	handler = func(next http.Handler) http.Handler {
		return gateway.SessionAuthMiddleware(next, r.authHandler)
	}(handler)

	// CORS middleware
	handler = middleware.CORS()(handler)

	// Logging middleware
	handler = middleware.Logging()(handler)

	// Recovery middleware (outermost - applied first)
	handler = middleware.Recovery()(handler)

	return handler
}

func (r *Router) isProtectedRoute(path string, protectedRoutes []string) bool {
	for _, route := range protectedRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func (r *Router) isAuthenticated(req *http.Request) bool {
	sessionID := r.extractSessionID(req)
	if sessionID == "" {
		return false
	}

	// Validate session
	_, err := r.authHandler.ValidateSession(req.Context(), sessionID)
	return err == nil
}

func (r *Router) isAdmin(req *http.Request) bool {
	sessionID := r.extractSessionID(req)
	if sessionID == "" {
		return false
	}

	// Check admin role
	return r.authHandler.IsAdmin(req.Context(), sessionID)
}

func (r *Router) extractSessionID(req *http.Request) string {
	// Try cookie first
	cookie, err := req.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try X-Session-ID header
	return req.Header.Get("X-Session-ID")
}

func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
