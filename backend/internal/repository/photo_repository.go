package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// PhotoRepository implements domain.PhotoRepository
type PhotoRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewPhotoRepository creates a new photo repository
func NewPhotoRepository(db *mongo.Database, logger *zap.Logger) domain.PhotoRepository {
	return &PhotoRepository{
		collection: db.Collection("photos"),
		logger:     logger,
	}
}

// Create creates a new photo
func (r *PhotoRepository) Create(ctx context.Context, photo *domain.Photo) error {

	photo.CreatedAt = time.Now()
	photo.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, photo)
	if err != nil {
		r.logger.Error("Failed to create photo", zap.Error(err))
		return fmt.Errorf("failed to create photo: %w", err)
	}

	photo.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a photo by ID
func (r *PhotoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Photo, error) {
	var photo domain.Photo
	filter := SoftDelete.GetActiveFilterByID(id)
	
	err := r.collection.FindOne(ctx, filter).Decode(&photo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("photo not found")
		}
		r.logger.Error("Failed to get photo by ID", zap.Error(err), zap.String("id", id.Hex()))
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	return &photo, nil
}

// GetByUserID retrieves photos by user ID with pagination
func (r *PhotoRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*domain.Photo, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	filter := SoftDelete.GetActiveFilterByUserID(userID)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get photos by user ID", zap.Error(err), zap.String("user_id", userID.Hex()))
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode photos: %w", err)
	}

	return photos, nil
}

// GetByCoupleID retrieves photos by couple (user and partner) with pagination
func (r *PhotoRepository) GetByCoupleID(ctx context.Context, userID, partnerID primitive.ObjectID, limit, offset int) ([]*domain.Photo, error) {
	baseFilter := SoftDelete.GetActiveFilter()
	coupleFilter := bson.M{
		"$or": []bson.M{
			{"user_id": userID, "partner_id": partnerID},
			{"user_id": partnerID, "partner_id": userID},
			{"user_id": userID, "partner_id": nil},
			{"user_id": partnerID, "partner_id": nil},
		},
		"is_private": false,
	}
	
	// Combine soft delete filter with couple filter
	filter := bson.M{"$and": []bson.M{baseFilter, coupleFilter}}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get photos by couple ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode photos: %w", err)
	}

	return photos, nil
}

// GetByDate retrieves photos by user ID and date
func (r *PhotoRepository) GetByDate(ctx context.Context, userID primitive.ObjectID, date time.Time) ([]*domain.Photo, error) {
	// Create date range for the entire day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := SoftDelete.GetActiveFilterWithCondition(bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	})

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get photos by date", zap.Error(err))
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode photos: %w", err)
	}

	return photos, nil
}

// Update updates a photo
func (r *PhotoRepository) Update(ctx context.Context, id primitive.ObjectID, photo *domain.Photo) error {
	photo.UpdatedAt = time.Now()

	update := bson.M{
		"$set": photo,
	}

	filter := SoftDelete.GetActiveFilterByID(id)
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update photo", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to update photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}

// Delete soft deletes a photo
func (r *PhotoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := SoftDelete.GetActiveFilterByID(id)
	update := SoftDelete.CreateSoftDeleteUpdate()

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to soft delete photo", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	r.logger.Info("Photo soft deleted successfully", zap.String("photo_id", id.Hex()))
	return nil
}

// Search searches photos by title, description, or tags
func (r *PhotoRepository) Search(ctx context.Context, userID primitive.ObjectID, query string, limit, offset int) ([]*domain.Photo, error) {
	filter := SoftDelete.GetActiveFilterWithCondition(bson.M{
		"user_id": userID,
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{query}}},
		},
	})

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to search photos", zap.Error(err))
		return nil, fmt.Errorf("failed to search photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode photos: %w", err)
	}

	return photos, nil
}

// Restore restores a soft deleted photo
func (r *PhotoRepository) Restore(ctx context.Context, id primitive.ObjectID) error {
	filter := SoftDelete.GetDeletedFilterWithCondition(bson.M{"_id": id})
	update := SoftDelete.CreateRestoreUpdate()

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to restore photo", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to restore photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("deleted photo not found")
	}

	r.logger.Info("Photo restored successfully", zap.String("photo_id", id.Hex()))
	return nil
}

// HardDelete permanently deletes a photo from database
func (r *PhotoRepository) HardDelete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error("Failed to hard delete photo", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to hard delete photo: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	r.logger.Warn("Photo hard deleted permanently", zap.String("photo_id", id.Hex()))
	return nil
}

// ListDeleted retrieves soft deleted photos with pagination
func (r *PhotoRepository) ListDeleted(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*domain.Photo, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "deleted_at", Value: -1}})

	filter := SoftDelete.GetDeletedFilterWithCondition(bson.M{"user_id": userID})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to list deleted photos", zap.Error(err))
		return nil, fmt.Errorf("failed to list deleted photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode deleted photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode deleted photos: %w", err)
	}

	return photos, nil
}
