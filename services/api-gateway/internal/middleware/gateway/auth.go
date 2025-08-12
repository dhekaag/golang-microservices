package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/handler"
	"github.com/dhekaag/golang-microservices/shared/pkg/session"
	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
)

type contextKey string

const (
	userSessionKey contextKey = "user_session"
	userIDKey      contextKey = "user_id"
	userRoleKey    contextKey = "user_role"
	sessionIDKey   contextKey = "session_id"
)

func SessionAuthMiddleware(next http.Handler, authHandler *handler.AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain paths
		skipPaths := []string{
			"/health",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/users",
			"/docs",
			"/api/v1/webhooks",
		}

		// Check if path should skip authentication
		for _, path := range skipPaths {
			if strings.HasPrefix(r.URL.Path, path) &&
				(r.Method == "POST" || strings.Contains(path, "health") || strings.Contains(path, "docs") || strings.Contains(path, "webhooks")) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Extract session ID
		sessionID := extractSessionIDFromRequest(r)
		if sessionID == "" {
			utils.SendError(w, http.StatusUnauthorized, "Missing session")
			return
		}

		// Validate session
		userSession, err := authHandler.ValidateSession(r.Context(), sessionID)
		if err != nil {
			utils.SendError(w, http.StatusUnauthorized, "Invalid session")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), userSessionKey, userSession)
		ctx = context.WithValue(ctx, userIDKey, userSession.UserID)
		ctx = context.WithValue(ctx, userRoleKey, userSession.Role)
		ctx = context.WithValue(ctx, sessionIDKey, sessionID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSession, ok := r.Context().Value("user_session").(*session.UserSession)
		if !ok || userSession.Role != "admin" {
			utils.SendError(w, http.StatusForbidden, "Access denied")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractSessionIDFromRequest(r *http.Request) string {
	// Try cookie first (preferred method)
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}

	// Try X-Session-ID header
	return r.Header.Get("X-Session-ID")
}
