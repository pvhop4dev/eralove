package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Photo represents a photo in the system
type Photo struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	PartnerID   *primitive.ObjectID `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required,min=1,max=200"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	ImageURL    string             `json:"image_url" bson:"image_url" validate:"required"`
	Date        time.Time          `json:"date" bson:"date"`
	Location    string             `json:"location,omitempty" bson:"location,omitempty"`
	Tags        []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	IsPrivate   bool               `json:"is_private" bson:"is_private"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time         `json:"-" bson:"deleted_at,omitempty"`
}

// CreatePhotoRequest represents the request to create a new photo
type CreatePhotoRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=200"`
	Description string    `json:"description,omitempty"`
	FilePath    string    `json:"file_path" validate:"required"` // Path from upload endpoint
	ImageURL    string    `json:"image_url,omitempty"`           // Will be generated from FilePath
	Date        time.Time `json:"date"`
	Location    string    `json:"location,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	IsPrivate   bool      `json:"is_private"`
}

// UpdatePhotoRequest represents the request to update a photo
type UpdatePhotoRequest struct {
	Title       string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description string    `json:"description,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	Date        time.Time `json:"date,omitempty"`
	Location    string    `json:"location,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	IsPrivate   *bool     `json:"is_private,omitempty"`
}

// PhotoResponse represents the photo response
type PhotoResponse struct {
	ID          primitive.ObjectID `json:"id"`
	UserID      primitive.ObjectID `json:"user_id"`
	PartnerID   *primitive.ObjectID `json:"partner_id,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	ImageURL    string             `json:"image_url"`
	Date        time.Time          `json:"date"`
	Location    string             `json:"location,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	IsPrivate   bool               `json:"is_private"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ToResponse converts Photo to PhotoResponse
func (p *Photo) ToResponse() *PhotoResponse {
	return &PhotoResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		PartnerID:   p.PartnerID,
		Title:       p.Title,
		Description: p.Description,
		ImageURL:    p.ImageURL,
		Date:        p.Date,
		Location:    p.Location,
		Tags:        p.Tags,
		IsPrivate:   p.IsPrivate,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// PhotoRepository defines the interface for photo data access
type PhotoRepository interface {
	Create(ctx context.Context, photo *Photo) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Photo, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*Photo, error)
	GetByCoupleID(ctx context.Context, userID, partnerID primitive.ObjectID, limit, offset int) ([]*Photo, error)
	GetByDate(ctx context.Context, userID primitive.ObjectID, date time.Time) ([]*Photo, error)
	Update(ctx context.Context, id primitive.ObjectID, photo *Photo) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Search(ctx context.Context, userID primitive.ObjectID, query string, limit, offset int) ([]*Photo, error)
	
	// Soft delete management
	Restore(ctx context.Context, id primitive.ObjectID) error
	HardDelete(ctx context.Context, id primitive.ObjectID) error
	ListDeleted(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*Photo, error)
}

// PhotoService defines the interface for photo business logic
type PhotoService interface {
	CreatePhoto(ctx context.Context, userID primitive.ObjectID, req *CreatePhotoRequest, file interface{}) (*PhotoResponse, error)
	CreatePhotoWithPath(ctx context.Context, userID primitive.ObjectID, req *CreatePhotoRequest) (*PhotoResponse, error)
	GetPhoto(ctx context.Context, photoID, userID primitive.ObjectID) (*PhotoResponse, error)
	GetUserPhotos(ctx context.Context, userID primitive.ObjectID, partnerID *primitive.ObjectID, page, limit int) ([]*PhotoResponse, int64, error)
	UpdatePhoto(ctx context.Context, photoID, userID primitive.ObjectID, req *UpdatePhotoRequest) (*PhotoResponse, error)
	DeletePhoto(ctx context.Context, photoID, userID primitive.ObjectID) error
}

// PhotoListResponse represents a list of photos response
type PhotoListResponse struct {
	Photos []*PhotoResponse `json:"photos"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}
