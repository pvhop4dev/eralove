package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/email"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// UserService implements domain.UserService
type UserService struct {
	userRepo        domain.UserRepository
	passwordManager *auth.PasswordManager
	jwtManager      *auth.JWTManager
	emailService    *email.EmailService
	logger          *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	emailService *email.EmailService,
	logger *zap.Logger,
) domain.UserService {
	return &UserService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
		jwtManager:      jwtManager,
		emailService:    emailService,
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

	// Generate email verification token
	verificationToken, err := s.generateSecureToken()
	if err != nil {
		s.logger.Error("Failed to generate verification token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate verification token")
	}

	// Set verification token expiry (24 hours)
	verificationExpiry := time.Now().Add(24 * time.Hour)

	// Convert Date to time.Time
	var dateOfBirth *time.Time
	if req.DateOfBirth != nil {
		dateOfBirth = req.DateOfBirth.ToTimePtr()
	}

	// Create user
	user := &domain.User{
		Name:                    req.Name,
		Email:                   req.Email,
		PasswordHash:            hashedPassword,
		DateOfBirth:             dateOfBirth,
		Gender:                  req.Gender,
		Avatar:                  req.Avatar,
		IsEmailVerified:         false,
		EmailVerificationToken:  verificationToken,
		EmailVerificationExpiry: &verificationExpiry,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user")
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Name, user.Email, verificationToken); err != nil {
		s.logger.Error("Failed to send verification email",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", user.Email))
		// Don't fail registration if email fails, just log it
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return user.ToResponse(), nil
}

// Login authenticates a user and returns user data and token pair
func (s *UserService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenPair, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login attempt with non-existent email", zap.String("email", req.Email))
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Warn("Login attempt with invalid password",
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", req.Email))
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Generate token pair (access + refresh tokens)
	authTokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Name)
	if err != nil {
		s.logger.Error("Failed to generate token pair", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to generate tokens")
	}

	// Convert auth.TokenPair to domain.TokenPair
	tokenPair := &domain.TokenPair{
		AccessToken:  authTokenPair.AccessToken,
		RefreshToken: authTokenPair.RefreshToken,
		TokenType:    authTokenPair.TokenType,
		ExpiresIn:    authTokenPair.ExpiresIn,
	}

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return user.ToResponse(), tokenPair, nil
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
		user.DateOfBirth = req.DateOfBirth.ToTimePtr()
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

// AuthenticateUser authenticates a user and returns a JWT token (backward compatibility - returns only access token)
func (s *UserService) AuthenticateUser(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, string, error) {
	user, tokenPair, err := s.Login(ctx, req)
	if err != nil {
		return nil, "", err
	}
	return user, tokenPair.AccessToken, nil
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

// generateSecureToken generates a cryptographically secure random token
func (s *UserService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyEmail verifies a user's email address
func (s *UserService) VerifyEmail(ctx context.Context, req *domain.EmailVerificationRequest) error {
	// Get user by verification token
	user, err := s.userRepo.GetByEmailVerificationToken(ctx, req.Token)
	if err != nil {
		s.logger.Warn("Email verification attempt with invalid token", zap.String("token", req.Token))
		return fmt.Errorf("invalid or expired verification token")
	}

	// Check if token is expired
	if user.EmailVerificationExpiry != nil && time.Now().After(*user.EmailVerificationExpiry) {
		s.logger.Warn("Email verification attempt with expired token", 
			zap.String("user_id", user.ID.Hex()),
			zap.Time("expiry", *user.EmailVerificationExpiry))
		return fmt.Errorf("verification token has expired")
	}

	// Update user to mark email as verified and clear verification token
	user.IsEmailVerified = true
	user.EmailVerificationToken = ""
	user.EmailVerificationExpiry = nil

	if err := s.userRepo.Update(ctx, user.ID, user); err != nil {
		s.logger.Error("Failed to update user after email verification",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()))
		return fmt.Errorf("failed to verify email")
	}

	s.logger.Info("Email verified successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return nil
}

// ResendVerificationEmail resends verification email to user
func (s *UserService) ResendVerificationEmail(ctx context.Context, req *domain.ResendVerificationRequest) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Resend verification attempt for non-existent email", zap.String("email", req.Email))
		// Don't reveal if email exists or not for security
		return nil
	}

	// Check if email is already verified
	if user.IsEmailVerified {
		s.logger.Info("Resend verification attempt for already verified email", 
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", req.Email))
		return fmt.Errorf("email is already verified")
	}

	// Generate new verification token
	token, err := s.generateSecureToken()
	if err != nil {
		s.logger.Error("Failed to generate verification token", zap.Error(err))
		return fmt.Errorf("failed to generate verification token")
	}

	// Set token expiry (24 hours)
	expiry := time.Now().Add(24 * time.Hour)
	user.EmailVerificationToken = token
	user.EmailVerificationExpiry = &expiry

	// Update user with new token
	if err := s.userRepo.Update(ctx, user.ID, user); err != nil {
		s.logger.Error("Failed to update user with new verification token",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()))
		return fmt.Errorf("failed to generate new verification token")
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(user.Name, user.Email, token); err != nil {
		s.logger.Error("Failed to send verification email",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", user.Email))
		return fmt.Errorf("failed to send verification email")
	}

	s.logger.Info("Verification email resent successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return nil
}

// ForgotPassword initiates password reset process
func (s *UserService) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Password reset attempt for non-existent email", zap.String("email", req.Email))
		// Don't reveal if email exists or not for security
		return nil
	}

	// Generate password reset token
	token, err := s.generateSecureToken()
	if err != nil {
		s.logger.Error("Failed to generate password reset token", zap.Error(err))
		return fmt.Errorf("failed to generate reset token")
	}

	// Set token expiry (1 hour)
	expiry := time.Now().Add(1 * time.Hour)
	user.PasswordResetToken = token
	user.PasswordResetExpiry = &expiry

	// Update user with reset token
	if err := s.userRepo.Update(ctx, user.ID, user); err != nil {
		s.logger.Error("Failed to update user with password reset token",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()))
		return fmt.Errorf("failed to generate reset token")
	}

	// Send password reset email
	if err := s.emailService.SendPasswordResetEmail(user.Name, user.Email, token); err != nil {
		s.logger.Error("Failed to send password reset email",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()),
			zap.String("email", user.Email))
		return fmt.Errorf("failed to send reset email")
	}

	s.logger.Info("Password reset email sent successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return nil
}

// ResetPassword resets user password using reset token
func (s *UserService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	// Validate new password
	if err := s.passwordManager.IsValidPassword(req.NewPassword); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Get user by reset token
	user, err := s.userRepo.GetByPasswordResetToken(ctx, req.Token)
	if err != nil {
		s.logger.Warn("Password reset attempt with invalid token", zap.String("token", req.Token))
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	if user.PasswordResetExpiry != nil && time.Now().After(*user.PasswordResetExpiry) {
		s.logger.Warn("Password reset attempt with expired token",
			zap.String("user_id", user.ID.Hex()),
			zap.Time("expiry", *user.PasswordResetExpiry))
		return fmt.Errorf("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := s.passwordManager.HashPassword(req.NewPassword)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return fmt.Errorf("failed to process new password")
	}

	// Update user with new password and clear reset token
	user.PasswordHash = hashedPassword
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = nil

	if err := s.userRepo.Update(ctx, user.ID, user); err != nil {
		s.logger.Error("Failed to update user password",
			zap.Error(err),
			zap.String("user_id", user.ID.Hex()))
		return fmt.Errorf("failed to reset password")
	}

	s.logger.Info("Password reset successfully",
		zap.String("user_id", user.ID.Hex()),
		zap.String("email", user.Email))

	return nil
}
