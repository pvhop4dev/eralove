package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Email        string             `json:"email" bson:"email" validate:"required,email"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	DateOfBirth  *time.Time         `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	Gender       string             `json:"gender,omitempty" bson:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Avatar       string             `json:"avatar,omitempty" bson:"avatar,omitempty"`
	PartnerID    *primitive.ObjectID `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	PartnerName  string             `json:"partner_name,omitempty" bson:"partner_name,omitempty"`
	AnniversaryDate *time.Time      `json:"anniversary_date,omitempty" bson:"anniversary_date,omitempty"`
	IsActive     bool               `json:"is_active" bson:"is_active"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Email       string     `json:"email" validate:"required,email"`
	Password    string     `json:"password" validate:"required,min=6"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      string     `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Avatar      string     `json:"avatar,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	Name        string     `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      string     `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Avatar      string     `json:"avatar,omitempty"`
	PartnerName string     `json:"partner_name,omitempty"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID              primitive.ObjectID `json:"id"`
	Name            string             `json:"name"`
	Email           string             `json:"email"`
	DateOfBirth     *time.Time         `json:"date_of_birth,omitempty"`
	Gender          string             `json:"gender,omitempty"`
	Avatar          string             `json:"avatar,omitempty"`
	PartnerID       *primitive.ObjectID `json:"partner_id,omitempty"`
	PartnerName     string             `json:"partner_name,omitempty"`
	AnniversaryDate *time.Time         `json:"anniversary_date,omitempty"`
	IsActive        bool               `json:"is_active"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
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
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id primitive.ObjectID, user *User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// RefreshTokenRequest represents the request to refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents the request to logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserService defines the interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	Register(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	AuthenticateUser(ctx context.Context, req *LoginRequest) (*UserResponse, string, error)
	Login(ctx context.Context, req *LoginRequest) (*UserResponse, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, *UserResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetProfile(ctx context.Context, userID primitive.ObjectID) (*UserResponse, error)
	UpdateProfile(ctx context.Context, userID primitive.ObjectID, req *UpdateUserRequest) (*UserResponse, error)
	DeleteAccount(ctx context.Context, userID primitive.ObjectID) error
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
