package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID    primitive.ObjectID `json:"user_id"`
	Email     string             `json:"email"`
	Name      string             `json:"name"`
	TokenType string             `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey            string
	accessExpiration     time.Duration
	refreshExpiration    time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessExpirationMinutes, refreshExpirationHours int) *JWTManager {
	return &JWTManager{
		secretKey:         secretKey,
		accessExpiration:  time.Duration(accessExpirationMinutes) * time.Minute,
		refreshExpiration: time.Duration(refreshExpirationHours) * time.Hour,
	}
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(userID primitive.ObjectID, email, name string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := j.generateToken(userID, email, name, "access", j.accessExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(userID, email, name, "refresh", j.refreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessExpiration.Seconds()),
	}, nil
}

// GenerateToken generates a new JWT access token (for backward compatibility)
func (j *JWTManager) GenerateToken(userID primitive.ObjectID, email, name string) (string, error) {
	return j.generateToken(userID, email, name, "access", j.accessExpiration)
}

// generateToken generates a JWT token with specified type and expiration
func (j *JWTManager) generateToken(userID primitive.ObjectID, email, name, tokenType string, expiration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "eralove-api",
			Subject:   userID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshTokenString generates a secure random refresh token string
func (j *JWTManager) GenerateRefreshTokenString() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// RefreshTokenPair generates new token pair from refresh token
func (j *JWTManager) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.Name)
}

// RefreshToken generates a new access token from refresh token (for backward compatibility)
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	tokenPair, err := j.RefreshTokenPair(tokenString)
	if err != nil {
		return "", err
	}
	return tokenPair.AccessToken, nil
}

// ValidateAccessToken validates specifically an access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates specifically a refresh token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	return claims, nil
}

// GetUserIDFromToken extracts user ID from token
func (j *JWTManager) GetUserIDFromToken(tokenString string) (primitive.ObjectID, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return claims.UserID, nil
}

// GenerateRefreshToken generates a new JWT refresh token
func (j *JWTManager) GenerateRefreshToken(userID primitive.ObjectID, email, name string) (string, error) {
	return j.generateToken(userID, email, name, "refresh", j.refreshExpiration)
}

// GetSecretKey returns the secret key (for middleware)
func (j *JWTManager) GetSecretKey() string {
	return j.secretKey
}
