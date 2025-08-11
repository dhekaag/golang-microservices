package router

import (
	"net/http"

	"github.com/dhekaag/golang-microservices/services/user-service/internal/handler"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/dhekaag/golang-microservices/shared/pkg/middleware"
)

type Router struct {
	userHandler *handler.UserHandler
}

func NewRouter(userHandler *handler.UserHandler) *Router {
	return &Router{
		userHandler: userHandler,
	}
}

func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"user-service"}`))
	})

	// Auth routes (no authentication required)
	mux.HandleFunc("/auth/register", r.userHandler.Register)
	mux.HandleFunc("/auth/login", r.userHandler.Login)

	// User management routes (authentication required)
	mux.HandleFunc("/users", r.handleUserRoutes)
	mux.HandleFunc("/users/", r.handleUserRoutes)

	// Apply middlewares
	handler := middleware.Chain(
		middleware.Recovery(),
		r.contextMiddleware,
		middleware.Logging(),
		middleware.CORS(),
	)(mux)

	return handler
}

func (r *Router) contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// Extract request ID from header
		if requestID := req.Header.Get("X-Request-ID"); requestID != "" {
			ctx = logger.WithRequestID(ctx, requestID)
		} else {
			// Generate new request ID if not provided
			ctx, _ = logger.GetOrCreateRequestID(ctx)
		}

		// Extract correlation ID from header
		if correlationID := req.Header.Get("X-Correlation-ID"); correlationID != "" {
			ctx = logger.WithCorrelationID(ctx, correlationID)
		} else {
			// Generate new correlation ID if not provided
			ctx, _ = logger.GetOrCreateCorrelationID(ctx)
		}

		// Extract user ID if provided (for authenticated requests)
		if userID := req.Header.Get("X-User-ID"); userID != "" {
			ctx = logger.WithUserID(ctx, userID)
		}

		// Update request with enhanced context
		req = req.WithContext(ctx)

		// Set response headers
		w.Header().Set("X-Request-ID", logger.GetRequestID(ctx))
		w.Header().Set("X-Correlation-ID", logger.GetCorrelationID(ctx))

		next.ServeHTTP(w, req)
	})
}

func (r *Router) handleUserRoutes(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		if req.URL.Query().Get("id") != "" || req.URL.Query().Get("public_id") != "" {
			r.userHandler.GetUser(w, req)
		} else {
			r.userHandler.ListUsers(w, req)
		}
	case http.MethodPut:
		r.userHandler.UpdateUser(w, req)
	case http.MethodDelete:
		r.userHandler.DeleteUser(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
