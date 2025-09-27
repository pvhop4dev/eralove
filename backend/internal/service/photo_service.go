package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// PhotoService implements domain.PhotoService
type PhotoService struct {
	photoRepo domain.PhotoRepository
	userRepo  domain.UserRepository
	logger    *zap.Logger
}

// NewPhotoService creates a new photo service
func NewPhotoService(
	photoRepo domain.PhotoRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.PhotoService {
	return &PhotoService{
		photoRepo: photoRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// CreatePhoto creates a new photo
func (s *PhotoService) CreatePhoto(ctx context.Context, userID primitive.ObjectID, req *domain.CreatePhotoRequest, file interface{}) (*domain.PhotoResponse, error) {
	// Get user to check if they have a partner
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Create photo
	photo := &domain.Photo{
		UserID:      userID,
		PartnerID:   user.PartnerID,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Date:        req.Date,
		Location:    req.Location,
		Tags:        req.Tags,
		IsPrivate:   req.IsPrivate,
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		s.logger.Error("Failed to create photo", zap.Error(err))
		return nil, fmt.Errorf("failed to create photo")
	}

	s.logger.Info("Photo created successfully",
		zap.String("photo_id", photo.ID.Hex()),
		zap.String("user_id", userID.Hex()))

	return photo.ToResponse(), nil
}

// GetPhoto retrieves a photo by ID
func (s *PhotoService) GetPhoto(ctx context.Context, photoID, userID primitive.ObjectID) (*domain.PhotoResponse, error) {
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return nil, fmt.Errorf("photo not found")
	}

	// Check if user has access to this photo
	if photo.UserID != userID && (photo.PartnerID == nil || *photo.PartnerID != userID) {
		return nil, fmt.Errorf("access denied")
	}

	return photo.ToResponse(), nil
}

// GetUserPhotos retrieves photos by user ID with pagination
func (s *PhotoService) GetUserPhotos(ctx context.Context, userID primitive.ObjectID, partnerID *primitive.ObjectID, page, limit int) ([]*domain.PhotoResponse, int64, error) {
	// Calculate offset from page
	offset := (page - 1) * limit
	
	photos, err := s.photoRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get user photos", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get photos")
	}

	// Get total count for pagination
	// Note: You may need to add a Count method to your repository
	total := int64(len(photos)) // This is a simplified approach

	responses := make([]*domain.PhotoResponse, len(photos))
	for i, photo := range photos {
		responses[i] = photo.ToResponse()
	}

	return responses, total, nil
}

// GetCouplePhotos retrieves photos for a couple with pagination
func (s *PhotoService) GetCouplePhotos(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*domain.PhotoResponse, error) {
	// Get user to find partner
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.PartnerID == nil {
		return nil, fmt.Errorf("user has no partner")
	}

	photos, err := s.photoRepo.GetByCoupleID(ctx, userID, *user.PartnerID, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get couple photos", zap.Error(err))
		return nil, fmt.Errorf("failed to get photos")
	}

	responses := make([]*domain.PhotoResponse, len(photos))
	for i, photo := range photos {
		responses[i] = photo.ToResponse()
	}

	return responses, nil
}

// GetPhotosByDate retrieves photos by date
func (s *PhotoService) GetPhotosByDate(ctx context.Context, userID primitive.ObjectID, date time.Time) ([]*domain.PhotoResponse, error) {
	photos, err := s.photoRepo.GetByDate(ctx, userID, date)
	if err != nil {
		s.logger.Error("Failed to get photos by date", zap.Error(err))
		return nil, fmt.Errorf("failed to get photos")
	}

	responses := make([]*domain.PhotoResponse, len(photos))
	for i, photo := range photos {
		responses[i] = photo.ToResponse()
	}

	return responses, nil
}

// UpdatePhoto updates a photo
func (s *PhotoService) UpdatePhoto(ctx context.Context, photoID, userID primitive.ObjectID, req *domain.UpdatePhotoRequest) (*domain.PhotoResponse, error) {
	// Get existing photo
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return nil, fmt.Errorf("photo not found")
	}

	// Check ownership
	if photo.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	// Update fields if provided
	if req.Title != "" {
		photo.Title = req.Title
	}
	if req.Description != "" {
		photo.Description = req.Description
	}
	if req.ImageURL != "" {
		photo.ImageURL = req.ImageURL
	}
	if !req.Date.IsZero() {
		photo.Date = req.Date
	}
	if req.Location != "" {
		photo.Location = req.Location
	}
	if req.Tags != nil {
		photo.Tags = req.Tags
	}
	if req.IsPrivate != nil {
		photo.IsPrivate = *req.IsPrivate
	}

	if err := s.photoRepo.Update(ctx, photoID, photo); err != nil {
		s.logger.Error("Failed to update photo", zap.Error(err))
		return nil, fmt.Errorf("failed to update photo")
	}

	s.logger.Info("Photo updated successfully",
		zap.String("photo_id", photoID.Hex()),
		zap.String("user_id", userID.Hex()))

	return photo.ToResponse(), nil
}

// DeletePhoto deletes a photo
func (s *PhotoService) DeletePhoto(ctx context.Context, photoID, userID primitive.ObjectID) error {
	// Get existing photo
	photo, err := s.photoRepo.GetByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("photo not found")
	}

	// Check ownership
	if photo.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if err := s.photoRepo.Delete(ctx, photoID); err != nil {
		s.logger.Error("Failed to delete photo", zap.Error(err))
		return fmt.Errorf("failed to delete photo")
	}

	s.logger.Info("Photo deleted successfully",
		zap.String("photo_id", photoID.Hex()),
		zap.String("user_id", userID.Hex()))

	return nil
}

// SearchPhotos searches photos by query
func (s *PhotoService) SearchPhotos(ctx context.Context, userID primitive.ObjectID, query string, limit, offset int) ([]*domain.PhotoResponse, error) {
	photos, err := s.photoRepo.Search(ctx, userID, query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search photos", zap.Error(err))
		return nil, fmt.Errorf("failed to search photos")
	}

	responses := make([]*domain.PhotoResponse, len(photos))
	for i, photo := range photos {
		responses[i] = photo.ToResponse()
	}

	return responses, nil
}
