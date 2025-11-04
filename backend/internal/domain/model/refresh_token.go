package model

import (
	"time"
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Token     string    `bson:"token" json:"token"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	IsRevoked bool      `bson:"is_revoked" json:"is_revoked"`
	DeviceID  string    `bson:"device_id,omitempty" json:"device_id,omitempty"`
	UserAgent string    `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}
