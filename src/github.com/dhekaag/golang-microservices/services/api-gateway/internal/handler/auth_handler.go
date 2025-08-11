package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dhekaag/golang-microservices/services/api-gateway/internal/config"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/dhekaag/golang-microservices/shared/pkg/session"
	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
)

type AuthHandler struct {
	userServiceURL string
	httpClient     *http.Client
	sessionManager *session.SessionManager
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Data      UserLoginData `json:"data"`
	SessionID string        `json:"session_id,omitempty"`
}

type UserLoginData struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Name  string `json:"name"`
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

func NewAuthHandler(config *config.ServicesConfig, sessionManager *session.SessionManager) *AuthHandler {
	// Configure HTTP client with optimized settings
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	}

	return &AuthHandler{
		userServiceURL: config.UserService,
		httpClient: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},
		sessionManager: sessionManager,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logger.GetOrCreateRequestID(r.Context())
	ctx, _ = logger.GetOrCreateCorrelationID(ctx)
	r = r.WithContext(ctx)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn(ctx, "Invalid request body", "error", err)
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		logger.Warn(ctx, "Missing email or password")
		utils.SendError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	userData, err := h.validateCredentials(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn(ctx, "Login validation failed", "error", err, "email", req.Email)
		utils.SendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		logger.Error(ctx, "Failed to generate session ID", "error", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	userSession := &session.UserSession{
		UserID:    userData.ID,
		Email:     userData.Email,
		Role:      userData.Role,
		Name:      userData.Name,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	}

	if err := h.sessionManager.CreateSession(ctx, sessionID, userSession); err != nil {
		logger.Error(ctx, "Failed to create session", "error", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(24 * time.Hour.Seconds()),
	})

	response := LoginResponse{
		Success:   true,
		Message:   "Login successful",
		Data:      *userData,
		SessionID: sessionID,
	}

	utils.SendSuccess(w, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) validateCredentials(ctx context.Context, email, password string) (*UserLoginData, error) {
	start := time.Now()

	// Get request context information
	requestID := logger.GetRequestID(ctx)
	correlationID := logger.GetCorrelationID(ctx)

	// Create the request URL
	url := fmt.Sprintf("%s/auth/login", h.userServiceURL)

	// Create request payload
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers including context information
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "API-Gateway/1.0")
	req.Header.Set("Connection", "keep-alive")

	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	if correlationID != "" {
		req.Header.Set("X-Correlation-ID", correlationID)
	}

	// Make the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		duration := time.Since(start)
		logger.Error(ctx, "❌ User service call failed",
			"error", err,
			"duration", duration,
			"service_url", url,
		)
		return nil, fmt.Errorf("failed to make request to user service: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "Failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Warn(ctx, "User service returned error",
			"status_code", resp.StatusCode,
			"response_body", string(body),
			"duration", duration,
		)

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	// Parse response
	var userResponse struct {
		Success bool          `json:"success"`
		Message string        `json:"message"`
		Data    UserLoginData `json:"data"`
	}

	if err := json.Unmarshal(body, &userResponse); err != nil {
		logger.Error(ctx, "Failed to parse user service response", "error", err, "body", string(body))
		return nil, fmt.Errorf("failed to parse user service response: %w", err)
	}

	// Check if login was successful
	if !userResponse.Success {
		logger.Warn(ctx, "User service login failed", "message", userResponse.Message)
		return nil, fmt.Errorf("login failed: %s", userResponse.Message)
	}

	return &userResponse.Data, nil
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		utils.SendError(w, http.StatusBadRequest, "No active session")
		return
	}

	// Delete session from Redis
	if err := h.sessionManager.DeleteSession(r.Context(), sessionID); err != nil {
		// Log error but don't fail the logout
		fmt.Printf("Failed to delete session: %v\n", err)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete cookie
	})

	utils.SendSuccess(w, http.StatusOK, "Logout successful", nil)
}

func (h *AuthHandler) ValidateSession(ctx context.Context, sessionID string) (*session.UserSession, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	userSession, err := h.sessionManager.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	return userSession, nil
}

func (h *AuthHandler) IsAdmin(ctx context.Context, sessionID string) bool {
	userSession, err := h.ValidateSession(ctx, sessionID)
	if err != nil {
		return false
	}

	return userSession.Role == "admin"
}

func (h *AuthHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		utils.SendError(w, http.StatusUnauthorized, "No active session")
		return
	}

	userSession, err := h.ValidateSession(r.Context(), sessionID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User info retrieved", userSession)
}

func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		utils.SendError(w, http.StatusUnauthorized, "No active session")
		return
	}

	if err := h.sessionManager.ExtendSession(r.Context(), sessionID); err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Failed to refresh session")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Session refreshed", nil)
}

func (h *AuthHandler) LogoutAllSessions(w http.ResponseWriter, r *http.Request) {
	sessionID := h.extractSessionID(r)
	if sessionID == "" {
		utils.SendError(w, http.StatusUnauthorized, "No active session")
		return
	}

	userSession, err := h.ValidateSession(r.Context(), sessionID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	if err := h.sessionManager.DeleteSessions(r.Context(), userSession.UserID); err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to logout all sessions")
		return
	}

	// Clear current session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	utils.SendSuccess(w, http.StatusOK, "All sessions logged out", nil)
}

func (h *AuthHandler) extractSessionID(r *http.Request) string {
	// Try cookie first
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

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use remote address
	return r.RemoteAddr
}
