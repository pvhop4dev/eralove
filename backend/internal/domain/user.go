package domain

import (
	"context"
	"time"
)

// User represents a user in the system
type User struct {
	ID                    string     `json:"id" db:"id"`
	Name                  string     `json:"name" db:"name" validate:"required,min=2,max=100"`
	Email                 string     `json:"email" db:"email" validate:"required,email"`
	PasswordHash          string     `json:"-" db:"password_hash"`
	DateOfBirth           *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	Gender                string     `json:"gender,omitempty" db:"gender" validate:"omitempty,oneof=male female other"`
	Avatar                string     `json:"avatar,omitempty" db:"avatar_url"`
	AvatarURL             string     `json:"avatar_url,omitempty" db:"avatar_url"`
	PartnerID             *string    `json:"partner_id,omitempty" db:"partner_id"`
	PartnerName           string     `json:"partner_name,omitempty" db:"-"`
	AnniversaryDate       *time.Time `json:"anniversary_date,omitempty" db:"-"`
	IsActive              bool       `json:"is_active" db:"-"`
	IsEmailVerified       bool       `json:"is_email_verified" db:"is_email_verified"`
	EmailVerificationToken string    `json:"-" db:"email_verification_token"`
	EmailVerificationExpiry *time.Time `json:"-" db:"email_verification_expires"`
	PasswordResetToken    string     `json:"-" db:"password_reset_token"`
	PasswordResetExpiry   *time.Time `json:"-" db:"password_reset_expires"`
	Bio                   string     `json:"bio,omitempty" db:"bio"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt             *time.Time `json:"-" db:"-"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=6"`
	DateOfBirth *Date   `json:"date_of_birth,omitempty"`
	Gender      string  `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Avatar      string  `json:"avatar,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	Name        string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	DateOfBirth *Date   `json:"date_of_birth,omitempty"`
	Gender      string  `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Avatar      string  `json:"avatar,omitempty"`
	PartnerName string  `json:"partner_name,omitempty"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	DateOfBirth     *time.Time `json:"date_of_birth,omitempty"`
	Gender          string     `json:"gender,omitempty"`
	Avatar          string     `json:"avatar,omitempty"`
	PartnerID       *string    `json:"partner_id,omitempty"`
	PartnerName     string     `json:"partner_name,omitempty"`
	AnniversaryDate *time.Time `json:"anniversary_date,omitempty"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		DateOfBirth:     u.DateOfBirth,
		Gender:          u.Gender,
		Avatar:          u.Avatar,
		PartnerID:       u.PartnerID,
		PartnerName:     u.PartnerName,
		AnniversaryDate: u.AnniversaryDate,
		IsActive:        u.IsActive,
		IsEmailVerified: u.IsEmailVerified,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// RefreshTokenRequest represents the request to refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents the request to logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// EmailVerificationRequest represents the request to verify email
type EmailVerificationRequest struct {
	Token string `json:"token" validate:"required"`
}

// ForgotPasswordRequest represents the request to reset password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the request to reset password with token
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ResendVerificationRequest represents the request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// UserService defines the interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	Register(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	AuthenticateUser(ctx context.Context, req *LoginRequest) (*UserResponse, string, error)
	Login(ctx context.Context, req *LoginRequest) (*UserResponse, *TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, *UserResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetProfile(ctx context.Context, userID string) (*UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateUserRequest) (*UserResponse, error)
	DeleteAccount(ctx context.Context, userID string) error
	
	// Email verification
	VerifyEmail(ctx context.Context, req *EmailVerificationRequest) error
	ResendVerificationEmail(ctx context.Context, req *ResendVerificationRequest) error
	
	// Password reset
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
