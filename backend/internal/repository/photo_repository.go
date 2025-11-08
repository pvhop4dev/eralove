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

// PhotoRepositoryNew implements domain.PhotoRepository with MatchCode
type PhotoRepositoryNew struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewPhotoRepositoryWithMatchCode creates a new photo repository
func NewPhotoRepositoryWithMatchCode(db *mongo.Database, logger *zap.Logger) domain.PhotoRepository {
	return &PhotoRepositoryNew{
		collection: db.Collection("photos"),
		logger:     logger,
	}
}

// Create creates a new photo
func (r *PhotoRepositoryNew) Create(ctx context.Context, photo *domain.Photo) error {
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
func (r *PhotoRepositoryNew) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Photo, error) {
	var photo domain.Photo
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}
	
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

// GetByMatchCode retrieves photos by match code with pagination
func (r *PhotoRepositoryNew) GetByMatchCode(ctx context.Context, matchCode string, limit, offset int) ([]*domain.Photo, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	filter := bson.M{
		"match_code": matchCode,
		"deleted_at": bson.M{"$exists": false},
	}
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get photos by match code", zap.Error(err), zap.String("match_code", matchCode))
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

// GetByMatchCodeAndDate retrieves photos by match code and date
func (r *PhotoRepositoryNew) GetByMatchCodeAndDate(ctx context.Context, matchCode string, date time.Time) ([]*domain.Photo, error) {
	// Get start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Second)

	filter := bson.M{
		"match_code": matchCode,
		"date": bson.M{
			"$gte": startOfDay,
			"$lte": endOfDay,
		},
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get photos by match code and date", zap.Error(err))
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

// DeleteByMatchCode deletes all photos for a match code (for unmatch)
func (r *PhotoRepositoryNew) DeleteByMatchCode(ctx context.Context, matchCode string) error {
	filter := bson.M{
		"match_code": matchCode,
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete photos by match code", zap.Error(err))
		return fmt.Errorf("failed to delete photos by match code: %w", err)
	}

	return nil
}

// Update updates a photo
func (r *PhotoRepositoryNew) Update(ctx context.Context, id primitive.ObjectID, photo *domain.Photo) error {
	photo.UpdatedAt = time.Now()

	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": photo,
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update photo", zap.Error(err))
		return fmt.Errorf("failed to update photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}

// Delete soft deletes a photo
func (r *PhotoRepositoryNew) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to delete photo", zap.Error(err))
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}

// SearchByMatchCode searches photos by match code and query
func (r *PhotoRepositoryNew) SearchByMatchCode(ctx context.Context, matchCode string, query string, limit, offset int) ([]*domain.Photo, error) {
	filter := bson.M{
		"match_code": matchCode,
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{query}}},
		},
		"deleted_at": bson.M{"$exists": false},
	}

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

// Restore restores a soft-deleted photo
func (r *PhotoRepositoryNew) Restore(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": true},
	}

	update := bson.M{
		"$unset": bson.M{"deleted_at": ""},
		"$set":   bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to restore photo", zap.Error(err))
		return fmt.Errorf("failed to restore photo: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("photo not found or not deleted")
	}

	return nil
}

// HardDelete permanently deletes a photo
func (r *PhotoRepositoryNew) HardDelete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to hard delete photo", zap.Error(err))
		return fmt.Errorf("failed to hard delete photo: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("photo not found")
	}

	return nil
}

// ListDeleted retrieves soft-deleted photos by match code
func (r *PhotoRepositoryNew) ListDeleted(ctx context.Context, matchCode string, limit, offset int) ([]*domain.Photo, error) {
	filter := bson.M{
		"match_code": matchCode,
		"deleted_at": bson.M{"$exists": true},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "deleted_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to list deleted photos", zap.Error(err))
		return nil, fmt.Errorf("failed to list deleted photos: %w", err)
	}
	defer cursor.Close(ctx)

	var photos []*domain.Photo
	if err := cursor.All(ctx, &photos); err != nil {
		r.logger.Error("Failed to decode photos", zap.Error(err))
		return nil, fmt.Errorf("failed to decode photos: %w", err)
	}

	return photos, nil
}
