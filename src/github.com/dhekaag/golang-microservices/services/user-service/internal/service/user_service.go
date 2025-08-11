package service

import (
	"context"
	"errors"

	"github.com/dhekaag/golang-microservices/services/user-service/internal/domain"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/dto"
	"github.com/dhekaag/golang-microservices/services/user-service/internal/repository"
	"github.com/dhekaag/golang-microservices/shared/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	CreateUser(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
	GetUserByPublicID(ctx context.Context, publicID string) (*dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error)
	ChangePassword(ctx context.Context, userID uint, req *dto.ChangePasswordRequest) error
	VerifyEmail(ctx context.Context, userID uint) error
}

type userService struct {
	repo   repository.UserRepository
	logger *logger.Logger
}

func NewUserService(repo repository.UserRepository, logger *logger.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	s.logger.Info(ctx, "Registering new user", "email", req.Email)

	// Check if user already exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Error(ctx, "Failed to check user existence", "error", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "Failed to hash password", "error", err)
		return nil, err
	}

	// Set default role if not provided
	role := domain.USER
	if req.Role != "" {
		role = domain.EnumRole(req.Role)
	}

	// Create user
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user", "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "User registered successfully", "user_id", user.ID, "email", user.Email)

	// Convert to DTO response
	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	s.logger.Info(ctx, "User login attempt", "email", req.Email)

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn(ctx, "Login failed - user not found", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn(ctx, "Login failed - invalid password", "email", req.Email)
		return nil, errors.New("invalid credentials")
	}

	s.logger.Info(ctx, "User logged in successfully", "user_id", user.ID, "email", user.Email)

	return &dto.LoginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	return s.Register(ctx, req)
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user by ID", "user_id", id, "error", err)
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) GetUserByPublicID(ctx context.Context, publicID string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByPublicID(ctx, publicID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user by public ID", "public_id", publicID, "error", err)
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "Failed to get user by email", "email", email, "error", err)
		return nil, err
	}

	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	s.logger.Info(ctx, "Updating user", "user_id", id)

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		// Check if email is already taken by another user
		existingUser, _ := s.repo.GetByEmail(ctx, *req.Email)
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, errors.New("email already taken")
		}
		user.Email = *req.Email
		user.EmailVerified = false // Reset verification if email changed
	}
	if req.Image != nil {
		user.Image = req.Image
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to update user", "user_id", id, "error", err)
		return nil, err
	}

	s.logger.Info(ctx, "User updated successfully", "user_id", user.ID)
	response := s.toUserResponse(user)
	return &response, nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	s.logger.Info(ctx, "Deleting user", "user_id", id)

	// Check if user exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete user", "user_id", id, "error", err)
		return err
	}

	s.logger.Info(ctx, "User deleted successfully", "user_id", id)
	return nil
}

func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*dto.UserResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, 0, err
	}

	var responses []*dto.UserResponse
	for _, user := range users {
		response := s.toUserResponse(user)
		responses = append(responses, &response)
	}

	return responses, total, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uint, req *dto.ChangePasswordRequest) error {
	s.logger.Info(ctx, "Changing password", "user_id", userID)

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "Failed to hash new password", "error", err)
		return err
	}

	user.Password = string(hashedPassword)
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to update password", "user_id", userID, "error", err)
		return err
	}

	s.logger.Info(ctx, "Password changed successfully", "user_id", userID)
	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, userID uint) error {
	s.logger.Info(ctx, "Verifying email", "user_id", userID)

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.EmailVerified = true
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to verify email", "user_id", userID, "error", err)
		return err
	}

	s.logger.Info(ctx, "Email verified successfully", "user_id", userID)
	return nil
}

// Helper method to convert domain.User to dto.UserResponse
func (s *userService) toUserResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            user.ID,
		PublicID:      user.PublicID,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Image:         user.Image,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
