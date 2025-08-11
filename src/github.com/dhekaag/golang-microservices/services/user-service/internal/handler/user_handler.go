package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dhekaag/golang-microservices/services/user-service/internal/dto"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/service"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
	"github.com/go-playground/validator/v10"
)

// user_handler.go
type UserHandler struct {
	userService service.UserService
	validator   *validator.Validate
	logger      *logger.Logger
}

func NewUserHandler(userService service.UserService, validator *validator.Validate, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
		logger:      logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(r.Context(), "Invalid request body for registration", "error", err)
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn(r.Context(), "Validation failed for registration", "error", err)
		utils.SendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		h.logger.Error(r.Context(), "Registration failed", "error", err, "email", req.Email)
		if strings.Contains(err.Error(), "already exists") {
			utils.SendError(w, http.StatusConflict, err.Error())
		} else {
			utils.SendError(w, http.StatusInternalServerError, "Registration failed")
		}
		return
	}

	h.logger.Info(r.Context(), "User registered successfully", "user_id", user.ID)
	utils.SendSuccess(w, http.StatusCreated, "User registered successfully", user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	ctx := r.Context()

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn(ctx, "Invalid request body for login", "error", err)
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Warn(ctx, "Validation failed for login", "error", err)
		utils.SendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	loginResponse, err := h.userService.Login(ctx, &req)
	if err != nil {
		h.logger.Warn(ctx, "Login failed", "error", err, "email", req.Email)
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.logger.Info(ctx, "✅ User logged in successfully", "user_id", loginResponse.ID)

	response := map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"data": map[string]interface{}{
			"id":    loginResponse.ID,
			"name":  loginResponse.Name,
			"email": loginResponse.Email,
			"role":  string(loginResponse.Role),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	publicID := r.URL.Query().Get("public_id")

	var user *dto.UserResponse
	var err error

	if userID != "" {
		id, parseErr := strconv.ParseUint(userID, 10, 32)
		if parseErr != nil {
			utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		user, err = h.userService.GetUserByID(r.Context(), uint(id))
	} else if publicID != "" {
		user, err = h.userService.GetUserByPublicID(r.Context(), publicID)
	} else {
		utils.SendError(w, http.StatusBadRequest, "User ID or Public ID required")
		return
	}

	if err != nil {
		utils.SendError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User retrieved successfully", user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.SendError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), uint(userID), &req)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to update user", "error", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.SendError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), uint(userID)); err != nil {
		h.logger.Error(r.Context(), "Failed to delete user", "error", err)
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "User deleted successfully", nil)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	users, total, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error(r.Context(), "Failed to list users", "error", err)
		utils.SendError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	response := map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	}

	utils.SendSuccess(w, http.StatusOK, "Users retrieved successfully", response)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.SendError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	if err := h.userService.ChangePassword(r.Context(), uint(userID), &req); err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Password changed successfully", nil)
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		utils.SendError(w, http.StatusBadRequest, "User ID required")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userService.VerifyEmail(r.Context(), uint(userID)); err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusOK, "Email verified successfully", nil)
}
