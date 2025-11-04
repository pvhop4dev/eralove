package domain

import (
	"context"
	"time"
)

// Photo represents a photo in the system
type Photo struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	PartnerID   *string    `json:"partner_id,omitempty" db:"partner_id"`
	FilePath    string     `json:"file_path" db:"file_path" validate:"required"`
	FileSize    int64      `json:"file_size,omitempty" db:"file_size"`
	MimeType    string     `json:"mime_type,omitempty" db:"mime_type"`
	Description string     `json:"description,omitempty" db:"description"`
	Location    string     `json:"location,omitempty" db:"location"`
	TakenAt     *time.Time `json:"taken_at,omitempty" db:"taken_at"`
	UploadedBy  string     `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CreatePhotoRequest represents the request to create a new photo
type CreatePhotoRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=200"`
	Description string  `json:"description,omitempty"`
	FilePath    string  `json:"file_path" validate:"required"` // Path from upload endpoint
	ImageURL    string  `json:"image_url,omitempty"`           // Will be generated from FilePath
	Date        *Date   `json:"date"`
	Location    string  `json:"location,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsPrivate   bool    `json:"is_private"`
}

// UpdatePhotoRequest represents the request to update a photo
type UpdatePhotoRequest struct {
	Title       string   `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description string   `json:"description,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	Date        *Date    `json:"date,omitempty"`
	Location    string   `json:"location,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsPrivate   *bool     `json:"is_private,omitempty"`
}

// PhotoResponse represents the photo response
type PhotoResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	PartnerID   *string    `json:"partner_id,omitempty"`
	FilePath    string     `json:"file_path"`
	FileSize    int64      `json:"file_size,omitempty"`
	MimeType    string     `json:"mime_type,omitempty"`
	Description string     `json:"description,omitempty"`
	Location    string     `json:"location,omitempty"`
	TakenAt     *time.Time `json:"taken_at,omitempty"`
	UploadedBy  string     `json:"uploaded_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts Photo to PhotoResponse
func (p *Photo) ToResponse() *PhotoResponse {
	return &PhotoResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		PartnerID:   p.PartnerID,
		FilePath:    p.FilePath,
		FileSize:    p.FileSize,
		MimeType:    p.MimeType,
		Description: p.Description,
		Location:    p.Location,
		TakenAt:     p.TakenAt,
		UploadedBy:  p.UploadedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// PhotoRepository defines the interface for photo data access
type PhotoRepository interface {
	Create(ctx context.Context, photo *Photo) error
	FindByID(ctx context.Context, id string) (*Photo, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*Photo, error)
	Update(ctx context.Context, photo *Photo) error
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

// PhotoService defines the interface for photo business logic
type PhotoService interface {
	CreatePhoto(ctx context.Context, userID string, req *CreatePhotoRequest, file interface{}) (*PhotoResponse, error)
	CreatePhotoWithPath(ctx context.Context, userID string, req *CreatePhotoRequest) (*PhotoResponse, error)
	GetPhoto(ctx context.Context, photoID, userID string) (*PhotoResponse, error)
	GetUserPhotos(ctx context.Context, userID string, partnerID *string, page, limit int) ([]*PhotoResponse, int64, error)
	UpdatePhoto(ctx context.Context, photoID, userID string, req *UpdatePhotoRequest) (*PhotoResponse, error)
	DeletePhoto(ctx context.Context, photoID, userID string) error
}

// PhotoListResponse represents a list of photos response
type PhotoListResponse struct {
	Photos []*PhotoResponse `json:"photos"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}
