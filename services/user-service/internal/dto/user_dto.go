package dto

import (
	"time"

	"github.com/dhekaag/golang-microservices/services/user-service/internal/domain"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=USER ADMIN"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	ID    uint            `json:"id"`
	Name  string          `json:"name"`
	Email string          `json:"email"`
	Role  domain.EnumRole `json:"role"`
}

type UpdateProfileRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Image *string `json:"image,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UserResponse struct {
	ID            uint            `json:"id"`
	PublicID      string          `json:"public_id"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	EmailVerified bool            `json:"email_verified"`
	Image         *string         `json:"image"`
	Role          domain.EnumRole `json:"role"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
}
