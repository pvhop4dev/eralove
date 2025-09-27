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

// UserRepository implements domain.UserRepository
type UserRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database, logger *zap.Logger) domain.UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
		logger:     logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "is_active": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("id", id.Hex()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email, "is_active": true}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, id primitive.ObjectID, user *domain.User) error {

	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": user,
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "is_active": true}, update)
	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "is_active": true}, update)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("Failed to decode users", zap.Error(err))
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}
