package handler

import (
	"strings"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService domain.UserService
	validator   *validator.Validate
	i18n        *i18n.I18n
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userService domain.UserService,
	validator *validator.Validate,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
		i18n:        i18n,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Success 201 {object} domain.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Registration")

	var req domain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Registration")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Registration", 
		zap.String("email", req.Email),
		zap.String("name", req.Name))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Registration", 
			zap.String("email", req.Email),
			zap.Any("validation_errors", getValidationErrors(err)))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	LogServiceCall(h.logger, "Registration", zap.String("email", req.Email))
	
	user, err := h.userService.Register(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Registration", zap.String("email", req.Email))

		// Check for specific error types
		if strings.Contains(err.Error(), "already exists") {
			h.logger.Warn("Registration failed: email already exists",
				zap.String("email", req.Email))
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:   "Email already exists",
				Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "user_already_exists", nil),
			})
		}

		if strings.Contains(err.Error(), "invalid password") {
			h.logger.Warn("Registration failed: invalid password",
				zap.String("email", req.Email))
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "Invalid password",
				Message: err.Error(),
			})
		}

		// Generic server error
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Registration failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Registration", 
		zap.String("email", req.Email),
		zap.String("user_id", user.ID.Hex()))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Data:    user,
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "registration_success", nil),
	})
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "User login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Login")

	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Login")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Login", zap.String("email", req.Email))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Login", 
			zap.String("email", req.Email),
			zap.Any("validation_errors", getValidationErrors(err)))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	LogServiceCall(h.logger, "Login", zap.String("email", req.Email))

	user, token, err := h.userService.Login(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Login", zap.String("email", req.Email))
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "Invalid credentials",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_credentials", nil),
		})
	}

	LogServiceSuccess(h.logger, "Login", 
		zap.String("email", req.Email),
		zap.String("user_id", user.ID.Hex()))

	return c.JSON(LoginResponse{
		User:         user,
		AccessToken:  token,
		RefreshToken: "",
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
		Message:      h.i18n.Translate(c.Get("Accept-Language", "en"), "login_successful", nil),
	})
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Get profile")
	
	userID := getUserIDFromContext(c)
	LogServiceCall(h.logger, "Get profile", zap.String("user_id", userID.Hex()))

	user, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		LogServiceError(h.logger, err, "Get profile", zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "User not found",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "not_found", nil),
		})
	}

	LogServiceSuccess(h.logger, "Get profile", zap.String("user_id", userID.Hex()))

	return c.JSON(SuccessResponse{
		Data:    user,
		Message: "Profile retrieved successfully",
	})
}

// UpdateProfile handles updating user profile
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateUserRequest true "Profile update data"
// @Success 200 {object} domain.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Update profile")
	
	userID := getUserIDFromContext(c)

	var req domain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Update profile")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Update profile", 
		zap.String("user_id", userID.Hex()),
		zap.String("name", req.Name))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Update profile", 
			zap.String("user_id", userID.Hex()),
			zap.Any("validation_errors", getValidationErrors(err)))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	LogServiceCall(h.logger, "Update profile", zap.String("user_id", userID.Hex()))

	user, err := h.userService.UpdateProfile(c.Context(), userID, &req)
	if err != nil {
		LogServiceError(h.logger, err, "Update profile", zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update profile",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Update profile", zap.String("user_id", userID.Hex()))

	return c.JSON(SuccessResponse{
		Data:    user,
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "profile_updated", nil),
	})
}

// DeleteAccount handles account deletion
// @Summary Delete user account
// @Description Soft delete current user's account
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/account [delete]
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Delete account")
	
	userID := getUserIDFromContext(c)
	LogServiceCall(h.logger, "Delete account", zap.String("user_id", userID.Hex()))

	if err := h.userService.DeleteAccount(c.Context(), userID); err != nil {
		LogServiceError(h.logger, err, "Delete account", zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete account",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Delete account", zap.String("user_id", userID.Hex()))

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "account_deleted", nil),
	})
}

// getValidationErrors converts validator errors to readable format
func getValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, e.Field()+" is "+e.Tag())
		}
	}
	return errors
}

// LoginResponse represents the login response
// @Description Login response with user data and authentication tokens
type LoginResponse struct {
	User         *domain.UserResponse `json:"user"`                                                            // User information
	AccessToken  string               `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`  // JWT access token
	RefreshToken string               `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT refresh token
	TokenType    string               `json:"token_type" example:"Bearer"`                                     // Token type
	ExpiresIn    int64                `json:"expires_in" example:"3600"`                                       // Token expiration time in seconds
	Message      string               `json:"message" example:"Login successful"`                              // Success message
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Refresh token")

	var req domain.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Refresh token")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Refresh token", 
		zap.String("token_prefix", SafeTokenLog(req.RefreshToken)))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Refresh token")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Refresh token")

	tokenPair, user, err := h.userService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		LogServiceError(h.logger, err, "Refresh token")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "Invalid refresh token",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_token", nil),
		})
	}

	LogServiceSuccess(h.logger, "Refresh token", zap.String("user_id", user.ID.Hex()))

	return c.JSON(LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		Message:      h.i18n.Translate(c.Get("Accept-Language", "en"), "login_successful", nil),
	})
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout user and revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LogoutRequest true "Logout data"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Logout")

	var req domain.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Logout")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Logout", 
		zap.String("token_prefix", SafeTokenLog(req.RefreshToken)))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Logout")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Logout")

	err := h.userService.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		LogServiceError(h.logger, err, "Logout")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to logout",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Logout")

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "logout_successful", nil),
	})
}

// VerifyEmail handles email verification
// @Summary Verify email address
// @Description Verify user's email address using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.EmailVerificationRequest true "Email verification data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /auth/verify-email [post]
func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Email verification")

	var req domain.EmailVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Email verification")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Email verification", 
		zap.String("token_prefix", SafeTokenLog(req.Token)))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Email verification")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Email verification")

	err := h.userService.VerifyEmail(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Email verification")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Email verification failed",
			Message: err.Error(),
		})
	}

	LogServiceSuccess(h.logger, "Email verification")

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "email_verified", nil),
	})
}

// ResendVerificationEmail handles resending verification email
// @Summary Resend verification email
// @Description Resend email verification link to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.ResendVerificationRequest true "Resend verification data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/resend-verification [post]
func (h *UserHandler) ResendVerificationEmail(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Resend verification email")

	var req domain.ResendVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Resend verification email")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Resend verification email", 
		zap.String("email", req.Email))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Resend verification email", 
			zap.String("email", req.Email))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Resend verification email", zap.String("email", req.Email))

	err := h.userService.ResendVerificationEmail(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Resend verification email", zap.String("email", req.Email))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Failed to resend verification email",
			Message: err.Error(),
		})
	}

	LogServiceSuccess(h.logger, "Resend verification email", zap.String("email", req.Email))

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "verification_email_sent", nil),
	})
}

// ForgotPassword handles password reset request
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.ForgotPasswordRequest true "Forgot password data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Forgot password")

	var req domain.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Forgot password")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Forgot password", 
		zap.String("email", req.Email))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Forgot password", 
			zap.String("email", req.Email))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Forgot password", zap.String("email", req.Email))

	err := h.userService.ForgotPassword(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Forgot password", zap.String("email", req.Email))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to process request",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Forgot password", zap.String("email", req.Email))

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "password_reset_email_sent", nil),
	})
}

// ResetPassword handles password reset
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Reset password data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Reset password")

	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Reset password")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, "Reset password", 
		zap.String("token_prefix", SafeTokenLog(req.Token)))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Reset password")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Reset password")

	err := h.userService.ResetPassword(c.Context(), &req)
	if err != nil {
		LogServiceError(h.logger, err, "Reset password")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Password reset failed",
			Message: err.Error(),
		})
	}

	LogServiceSuccess(h.logger, "Reset password")

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "password_reset_successful", nil),
	})
}
