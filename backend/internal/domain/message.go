package domain

import (
	"context"
	"time"
)

// Message represents a message between users
type Message struct {
	ID          string     `bson:"_id,omitempty" json:"id"`
	SenderID    string     `bson:"sender_id" json:"sender_id"`
	ReceiverID  string     `bson:"receiver_id" json:"receiver_id"`
	Content     string     `bson:"content" json:"content"`
	MessageType string     `bson:"message_type" json:"message_type"` // text, image, etc.
	IsRead      bool       `bson:"is_read" json:"is_read"`
	IsDeleted   bool       `bson:"is_deleted" json:"is_deleted"`
	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
	ReadAt      *time.Time `bson:"read_at,omitempty" json:"read_at,omitempty"`
	DeletedAt   *time.Time `json:"-" bson:"deleted_at,omitempty"`
}

// Conversation represents a conversation summary
type Conversation struct {
	PartnerID     string    `json:"partner_id"`
	PartnerName   string    `json:"partner_name"`
	PartnerAvatar string    `json:"partner_avatar,omitempty"`
	LastMessage   *Message  `json:"last_message"`
	UnreadCount   int64     `json:"unread_count"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateMessageRequest represents the request to create a message
type CreateMessageRequest struct {
	ReceiverID  string `json:"receiver_id" validate:"required"`
	Content     string `json:"content" validate:"required,min=1,max=1000"`
	MessageType string `json:"message_type" validate:"omitempty,oneof=text image"`
}

// MarkAsReadRequest represents the request to mark messages as read
type MarkAsReadRequest struct {
	PartnerID string `json:"partner_id" validate:"required"`
}

// MessageResponse represents a message response
type MessageResponse struct {
	ID          string     `json:"id"`
	SenderID    string     `json:"sender_id"`
	ReceiverID  string     `json:"receiver_id"`
	Content     string     `json:"content"`
	MessageType string     `json:"message_type"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
}

// MessageListResponse represents a list of messages response
type MessageListResponse struct {
	Messages []*MessageResponse `json:"messages"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
}

// ConversationListResponse represents a list of conversations response
type ConversationListResponse struct {
	Conversations []*Conversation `json:"conversations"`
	Total         int64           `json:"total"`
	Page          int             `json:"page"`
	Limit         int             `json:"limit"`
}

// MessageService defines the interface for message operations
type MessageService interface {
	SendMessage(ctx context.Context, senderID string, req *CreateMessageRequest) (*MessageResponse, error)
	GetConversation(ctx context.Context, userID, partnerID string, page, limit int) ([]*MessageResponse, int64, error)
	GetUserConversations(ctx context.Context, userID string, page, limit int) ([]*Conversation, int64, error)
	MarkAsRead(ctx context.Context, userID, partnerID string) error
	DeleteMessage(ctx context.Context, messageID, userID string) error
}

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	Create(ctx context.Context, message *Message) error
	FindByID(ctx context.Context, id string) (*Message, error)
	FindConversation(ctx context.Context, userID, partnerID string, page, limit int) ([]*Message, int64, error)
	FindUserConversations(ctx context.Context, userID string, page, limit int) ([]*Conversation, int64, error)
	MarkAsRead(ctx context.Context, userID, partnerID string) error
	SoftDelete(ctx context.Context, messageID, userID string) error
	Update(ctx context.Context, message *Message) error
}
