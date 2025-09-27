package handler

import (
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

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
// @Param request body domain.CreateUserRequest true "User registration data"
// @Success 201 {object} domain.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req domain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	user, err := h.userService.Register(c.Context(), &req)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))

		if err.Error() == "user with email "+req.Email+" already exists" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error:   "Email already exists",
				Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "email_exists", nil),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Registration failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "registration_error", nil),
		})
	}

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
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	user, token, err := h.userService.Login(c.Context(), &req)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))

		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "Invalid credentials",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_credentials", nil),
		})
	}

	return c.JSON(LoginResponse{
		User:         user,
		AccessToken:  token,
		RefreshToken: "",
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
		Message:      h.i18n.Translate(c.Get("Accept-Language", "en"), "login_success", nil),
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
	userID := getUserIDFromContext(c)

	user, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.String("user_id", userID.Hex()), zap.Error(err))

		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "User not found",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "user_not_found", nil),
		})
	}

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
	userID := getUserIDFromContext(c)

	var req domain.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	user, err := h.userService.UpdateProfile(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to update profile", zap.String("user_id", userID.Hex()), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update profile",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "failed_update_personal_info", nil),
		})
	}

	return c.JSON(SuccessResponse{
		Data:    user,
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "personal_info_updated", nil),
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
	userID := getUserIDFromContext(c)

	if err := h.userService.DeleteAccount(c.Context(), userID); err != nil {
		h.logger.Error("Failed to delete account", zap.String("user_id", userID.Hex()), zap.Error(err))

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete account",
			Message: "Failed to delete account",
		})
	}

	return c.JSON(SuccessResponse{
		Message: "Account deleted successfully",
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
	User         *domain.UserResponse `json:"user"`                                    // User information
	AccessToken  string               `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT access token
	RefreshToken string               `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT refresh token
	TokenType    string               `json:"token_type" example:"Bearer"`            // Token type
	ExpiresIn    int64                `json:"expires_in" example:"3600"`              // Token expiration time in seconds
	Message      string               `json:"message" example:"Login successful"`     // Success message
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
	var req domain.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	tokenPair, user, err := h.userService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh token", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error:   "Invalid refresh token",
			Message: err.Error(),
		})
	}

	return c.JSON(LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		Message:      h.i18n.Translate(c.Get("Accept-Language", "en"), "token_refreshed", nil),
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
	var req domain.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	err := h.userService.Logout(c.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Failed to logout", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to logout",
			Message: err.Error(),
		})
	}

	return c.JSON(SuccessResponse{
		Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "logout_successful", nil),
	})
}
