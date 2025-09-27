package service

import (
	"context"
	"fmt"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// UserService implements domain.UserService
type UserService struct {
	userRepo        domain.UserRepository
	passwordManager *auth.PasswordManager
	jwtManager      *auth.JWTManager
	logger          *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	logger *zap.Logger,
) domain.UserService {
	return &UserService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
		jwtManager:      jwtManager,
		logger:          logger,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Validate password
	if err := s.passwordManager.IsValidPassword(req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to process password")
	}

	// Create user
	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		Avatar:       req.Avatar,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user")
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return user.ToResponse(), nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login attempt with non-existent email", zap.String("email", req.Email))
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Warn("Login attempt with invalid password",
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", req.Email))
		return nil, "", fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, "", fmt.Errorf("failed to generate token")
	}

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return user.ToResponse(), token, nil
}

// GetProfile retrieves user profile
func (s *UserService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user profile",
			zap.Error(err),
			zap.String("user_id", userID.Hex()))
		return nil, fmt.Errorf("failed to get user profile")
	}

	return user.ToResponse(), nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, req *domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.PartnerName != "" {
		user.PartnerName = req.PartnerName
	}

	// Update user
	if err := s.userRepo.Update(ctx, userID, user); err != nil {
		s.logger.Error("Failed to update user profile",
			zap.Error(err),
			zap.String("user_id", userID.Hex()))
		return nil, fmt.Errorf("failed to update profile")
	}

	s.logger.Info("User profile updated successfully",
		zap.String("user_id", userID.Hex()))

	return user.ToResponse(), nil
}

// CreateUser creates a new user account (alias for Register)
func (s *UserService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	return s.Register(ctx, req)
}

// AuthenticateUser authenticates a user and returns a JWT token (alias for Login)
func (s *UserService) AuthenticateUser(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, string, error) {
	return s.Login(ctx, req)
}

// RefreshToken refreshes an access token using a refresh token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, *domain.UserResponse, error) {
	// Validate refresh token and extract user info
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, nil, fmt.Errorf("invalid refresh token")
	}

	userID := claims.UserID
	email := claims.Email
	name := claims.Name

	// Get user to ensure they still exist and are active
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user for token refresh",
			zap.Error(err),
			zap.String("user_id", userID.Hex()))
		return nil, nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("user account is inactive")
	}

	// Generate new token pair
	accessToken, err := s.jwtManager.GenerateToken(userID, email, name)
	if err != nil {
		s.logger.Error("Failed to generate new access token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate token")
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(userID, email, name)
	if err != nil {
		s.logger.Error("Failed to generate new refresh token", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate refresh token")
	}

	// Return new tokens and user info
	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, user.ToResponse(), nil
}

// Logout revokes a refresh token
func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	// Validate refresh token and extract user info
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token during logout", zap.Error(err))
		return fmt.Errorf("invalid refresh token")
	}

	// In a production system, you would typically:
	// 1. Add the refresh token to a blacklist/revocation list
	// 2. Store revoked tokens in Redis with expiration
	// For now, we'll just log the logout event

	s.logger.Info("User logged out successfully",
		zap.String("user_id", claims.UserID.Hex()))

	return nil
}

// DeleteAccount soft deletes a user account
func (s *UserService) DeleteAccount(ctx context.Context, userID primitive.ObjectID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user account",
			zap.Error(err),
			zap.String("user_id", userID.Hex()))
		return fmt.Errorf("failed to delete account")
	}

	s.logger.Info("User account deleted successfully",
		zap.String("user_id", userID.Hex()))

	return nil
}
