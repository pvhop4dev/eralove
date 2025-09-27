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

// getActiveUserFilter returns a filter for active (non-deleted) users
func getActiveUserFilter() bson.M {
	return bson.M{
		"is_active": true,
		"deleted_at": bson.M{"$exists": false},
	}
}

// getActiveUserFilterWithCondition returns a filter for active users with additional conditions
func getActiveUserFilterWithCondition(condition bson.M) bson.M {
	filter := getActiveUserFilter()
	for k, v := range condition {
		filter[k] = v
	}
	return filter
}

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
	// Email verification defaults are set in service layer

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
	filter := getActiveUserFilterWithCondition(bson.M{"_id": id})
	
	err := r.collection.FindOne(ctx, filter).Decode(&user)
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
	filter := getActiveUserFilterWithCondition(bson.M{"email": email})
	
	err := r.collection.FindOne(ctx, filter).Decode(&user)
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

	filter := getActiveUserFilterWithCondition(bson.M{"_id": id})
	result, err := r.collection.UpdateOne(ctx, filter, update)
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
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"deleted_at": now,
			"updated_at": now,
		},
	}

	filter := getActiveUserFilterWithCondition(bson.M{"_id": id})
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info("User soft deleted successfully", 
		zap.String("user_id", id.Hex()),
		zap.Time("deleted_at", now))

	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	filter := getActiveUserFilter()
	cursor, err := r.collection.Find(ctx, filter, opts)
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

// GetByEmailVerificationToken retrieves a user by email verification token
func (r *UserRepository) GetByEmailVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	filter := getActiveUserFilterWithCondition(bson.M{
		"email_verification_token": token,
	})
	
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by email verification token", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByPasswordResetToken retrieves a user by password reset token
func (r *UserRepository) GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	filter := getActiveUserFilterWithCondition(bson.M{
		"password_reset_token": token,
	})
	
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error("Failed to get user by password reset token", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Restore restores a soft deleted user
func (r *UserRepository) Restore(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_active":  true,
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"deleted_at": "",
		},
	}

	// Find deleted user
	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": true},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to restore user", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to restore user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("deleted user not found")
	}

	r.logger.Info("User restored successfully", zap.String("user_id", id.Hex()))
	return nil
}

// HardDelete permanently deletes a user from database
func (r *UserRepository) HardDelete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error("Failed to hard delete user", zap.Error(err), zap.String("id", id.Hex()))
		return fmt.Errorf("failed to hard delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Warn("User hard deleted permanently", zap.String("user_id", id.Hex()))
	return nil
}

// ListDeleted retrieves soft deleted users with pagination
func (r *UserRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "deleted_at", Value: -1}})

	filter := bson.M{
		"deleted_at": bson.M{"$exists": true},
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to list deleted users", zap.Error(err))
		return nil, fmt.Errorf("failed to list deleted users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("Failed to decode deleted users", zap.Error(err))
		return nil, fmt.Errorf("failed to decode deleted users: %w", err)
	}

	return users, nil
}
