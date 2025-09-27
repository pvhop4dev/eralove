package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MatchRequestStatus represents the status of a match request
type MatchRequestStatus string

const (
	MatchRequestStatusPending  MatchRequestStatus = "pending"
	MatchRequestStatusAccepted MatchRequestStatus = "accepted"
	MatchRequestStatusDeclined MatchRequestStatus = "declined"
	MatchRequestStatusIgnored  MatchRequestStatus = "ignored"
)

// MatchRequest represents a match request between users
type MatchRequest struct {
	ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	SenderID        primitive.ObjectID  `json:"sender_id" bson:"sender_id"`
	ReceiverID      primitive.ObjectID  `json:"receiver_id" bson:"receiver_id"`
	ReceiverEmail   string              `json:"receiver_email" bson:"receiver_email"`
	AnniversaryDate time.Time           `json:"anniversary_date" bson:"anniversary_date"`
	Message         string              `json:"message,omitempty" bson:"message,omitempty"`
	Status          MatchRequestStatus  `json:"status" bson:"status"`
	CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
	RespondedAt     *time.Time          `json:"responded_at,omitempty" bson:"responded_at,omitempty"`
}

// CreateMatchRequestRequest represents the request to create a match request
type CreateMatchRequestRequest struct {
	ReceiverEmail   string    `json:"receiver_email" validate:"required,email"`
	AnniversaryDate time.Time `json:"anniversary_date" validate:"required"`
	Message         string    `json:"message,omitempty"`
}

// RespondToMatchRequestRequest represents the request to respond to a match request
type RespondToMatchRequestRequest struct {
	Action string `json:"action" validate:"required,oneof=accept reject"`
}

// MatchRequestResponse represents the match request response
type MatchRequestResponse struct {
	ID              primitive.ObjectID  `json:"id"`
	SenderID        primitive.ObjectID  `json:"sender_id"`
	SenderName      string              `json:"sender_name,omitempty"`
	SenderEmail     string              `json:"sender_email,omitempty"`
	ReceiverID      primitive.ObjectID  `json:"receiver_id"`
	ReceiverEmail   string              `json:"receiver_email"`
	AnniversaryDate time.Time           `json:"anniversary_date"`
	Message         string              `json:"message,omitempty"`
	Status          MatchRequestStatus  `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	RespondedAt     *time.Time          `json:"responded_at,omitempty"`
}

// ToResponse converts MatchRequest to MatchRequestResponse
func (mr *MatchRequest) ToResponse() *MatchRequestResponse {
	return &MatchRequestResponse{
		ID:              mr.ID,
		SenderID:        mr.SenderID,
		ReceiverID:      mr.ReceiverID,
		ReceiverEmail:   mr.ReceiverEmail,
		AnniversaryDate: mr.AnniversaryDate,
		Message:         mr.Message,
		Status:          mr.Status,
		CreatedAt:       mr.CreatedAt,
		UpdatedAt:       mr.UpdatedAt,
		RespondedAt:     mr.RespondedAt,
	}
}

// MatchRequestRepository defines the interface for match request data access
type MatchRequestRepository interface {
	Create(matchRequest *MatchRequest) error
	GetByID(id primitive.ObjectID) (*MatchRequest, error)
	GetBySenderID(senderID primitive.ObjectID, limit, offset int) ([]*MatchRequest, error)
	GetByReceiverID(receiverID primitive.ObjectID, limit, offset int) ([]*MatchRequest, error)
	GetByReceiverEmail(email string, limit, offset int) ([]*MatchRequest, error)
	GetPendingByReceiverID(receiverID primitive.ObjectID) ([]*MatchRequest, error)
	Update(id primitive.ObjectID, matchRequest *MatchRequest) error
	Delete(id primitive.ObjectID) error
	ExistsPendingRequest(senderID, receiverID primitive.ObjectID) (bool, error)
}

// MatchRequestListResponse represents a list of match requests response
type MatchRequestListResponse struct {
	MatchRequests []*MatchRequestResponse `json:"match_requests"`
	Total         int64                   `json:"total"`
	Page          int                     `json:"page"`
	Limit         int                     `json:"limit"`
}

// MatchRequestService defines the interface for match request business logic
type MatchRequestService interface {
	SendMatchRequest(ctx context.Context, senderID primitive.ObjectID, req *CreateMatchRequestRequest) (*MatchRequestResponse, error)
	GetMatchRequest(ctx context.Context, requestID, userID primitive.ObjectID) (*MatchRequestResponse, error)
	GetSentRequests(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*MatchRequestResponse, int64, error)
	GetReceivedRequests(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*MatchRequestResponse, int64, error)
	RespondToMatchRequest(ctx context.Context, requestID, userID primitive.ObjectID, action string) (*MatchRequestResponse, error)
	CancelMatchRequest(ctx context.Context, requestID, userID primitive.ObjectID) error
}
