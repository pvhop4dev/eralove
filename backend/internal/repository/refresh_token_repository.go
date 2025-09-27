package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// RefreshTokenRepository handles refresh token database operations
type RefreshTokenRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *mongo.Database, logger *zap.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		collection: db.Collection("refresh_tokens"),
		logger:     logger,
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, refreshToken *model.RefreshToken) error {
	refreshToken.ID = primitive.NewObjectID()
	refreshToken.CreatedAt = time.Now()
	refreshToken.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, refreshToken)
	if err != nil {
		r.logger.Error("Failed to create refresh token", zap.Error(err))
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	r.logger.Info("Refresh token created", zap.String("token_id", refreshToken.ID.Hex()))
	return nil
}

// FindByToken finds a refresh token by token string
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("refresh token not found")
		}
		r.logger.Error("Failed to find refresh token", zap.Error(err))
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return &refreshToken, nil
}

// FindByUserID finds all refresh tokens for a user
func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.RefreshToken, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":    userID,
		"is_revoked": false,
	})
	if err != nil {
		r.logger.Error("Failed to find refresh tokens by user ID", zap.Error(err))
		return nil, fmt.Errorf("failed to find refresh tokens: %w", err)
	}
	defer cursor.Close(ctx)

	var refreshTokens []*model.RefreshToken
	if err = cursor.All(ctx, &refreshTokens); err != nil {
		r.logger.Error("Failed to decode refresh tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to decode refresh tokens: %w", err)
	}

	return refreshTokens, nil
}

// RevokeToken revokes a refresh token
func (r *RefreshTokenRepository) RevokeToken(ctx context.Context, token string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{
			"$set": bson.M{
				"is_revoked": true,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error("Failed to revoke refresh token", zap.Error(err))
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("refresh token not found")
	}

	r.logger.Info("Refresh token revoked", zap.String("token", token))
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"user_id":    userID,
			"is_revoked": false,
		},
		bson.M{
			"$set": bson.M{
				"is_revoked": true,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		r.logger.Error("Failed to revoke all user tokens", zap.Error(err))
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	r.logger.Info("All user refresh tokens revoked", zap.String("user_id", userID.Hex()))
	return nil
}

// DeleteExpiredTokens deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		r.logger.Error("Failed to delete expired tokens", zap.Error(err))
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	r.logger.Info("Expired refresh tokens deleted")
	return nil
}

// Update updates a refresh token
func (r *RefreshTokenRepository) Update(ctx context.Context, refreshToken *model.RefreshToken) error {
	refreshToken.UpdatedAt = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": refreshToken.ID},
		bson.M{"$set": refreshToken},
	)
	if err != nil {
		r.logger.Error("Failed to update refresh token", zap.Error(err))
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("refresh token not found")
	}

	return nil
}
