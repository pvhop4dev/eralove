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

// MatchRequestRepository implements domain.MatchRequestRepository
type MatchRequestRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewMatchRequestRepository creates a new match request repository
func NewMatchRequestRepository(db *mongo.Database, logger *zap.Logger) domain.MatchRequestRepository {
	return &MatchRequestRepository{
		collection: db.Collection("match_requests"),
		logger:     logger,
	}
}

// Create creates a new match request
func (r *MatchRequestRepository) Create(matchRequest *domain.MatchRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, matchRequest)
	if err != nil {
		r.logger.Error("Failed to create match request", zap.Error(err))
		return fmt.Errorf("failed to create match request: %w", err)
	}

	return nil
}

// GetByID retrieves a match request by ID
func (r *MatchRequestRepository) GetByID(id primitive.ObjectID) (*domain.MatchRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var matchRequest domain.MatchRequest
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&matchRequest)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("match request not found")
		}
		r.logger.Error("Failed to get match request by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get match request: %w", err)
	}

	return &matchRequest, nil
}

// GetBySenderID retrieves match requests sent by a user
func (r *MatchRequestRepository) GetBySenderID(senderID primitive.ObjectID, limit, offset int) ([]*domain.MatchRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"sender_id": senderID}, opts)
	if err != nil {
		r.logger.Error("Failed to get match requests by sender", zap.Error(err))
		return nil, fmt.Errorf("failed to get match requests: %w", err)
	}
	defer cursor.Close(ctx)

	var matchRequests []*domain.MatchRequest
	if err := cursor.All(ctx, &matchRequests); err != nil {
		r.logger.Error("Failed to decode match requests", zap.Error(err))
		return nil, fmt.Errorf("failed to decode match requests: %w", err)
	}

	return matchRequests, nil
}

// GetByReceiverID retrieves match requests received by a user
func (r *MatchRequestRepository) GetByReceiverID(receiverID primitive.ObjectID, limit, offset int) ([]*domain.MatchRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"receiver_id": receiverID}, opts)
	if err != nil {
		r.logger.Error("Failed to get match requests by receiver", zap.Error(err))
		return nil, fmt.Errorf("failed to get match requests: %w", err)
	}
	defer cursor.Close(ctx)

	var matchRequests []*domain.MatchRequest
	if err := cursor.All(ctx, &matchRequests); err != nil {
		r.logger.Error("Failed to decode match requests", zap.Error(err))
		return nil, fmt.Errorf("failed to decode match requests: %w", err)
	}

	return matchRequests, nil
}

// GetByReceiverEmail retrieves match requests by receiver email
func (r *MatchRequestRepository) GetByReceiverEmail(email string, limit, offset int) ([]*domain.MatchRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"receiver_email": email}, opts)
	if err != nil {
		r.logger.Error("Failed to get match requests by receiver email", zap.Error(err))
		return nil, fmt.Errorf("failed to get match requests: %w", err)
	}
	defer cursor.Close(ctx)

	var matchRequests []*domain.MatchRequest
	if err := cursor.All(ctx, &matchRequests); err != nil {
		r.logger.Error("Failed to decode match requests", zap.Error(err))
		return nil, fmt.Errorf("failed to decode match requests: %w", err)
	}

	return matchRequests, nil
}

// GetPendingByReceiverID retrieves pending match requests for a user
func (r *MatchRequestRepository) GetPendingByReceiverID(receiverID primitive.ObjectID) ([]*domain.MatchRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"receiver_id": receiverID,
		"status":      domain.MatchRequestStatusPending,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to get pending match requests", zap.Error(err))
		return nil, fmt.Errorf("failed to get pending match requests: %w", err)
	}
	defer cursor.Close(ctx)

	var matchRequests []*domain.MatchRequest
	if err := cursor.All(ctx, &matchRequests); err != nil {
		r.logger.Error("Failed to decode match requests", zap.Error(err))
		return nil, fmt.Errorf("failed to decode match requests: %w", err)
	}

	return matchRequests, nil
}

// Update updates a match request
func (r *MatchRequestRepository) Update(id primitive.ObjectID, matchRequest *domain.MatchRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": matchRequest}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update match request", zap.Error(err))
		return fmt.Errorf("failed to update match request: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("match request not found")
	}

	return nil
}

// Delete deletes a match request
func (r *MatchRequestRepository) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		r.logger.Error("Failed to delete match request", zap.Error(err))
		return fmt.Errorf("failed to delete match request: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("match request not found")
	}

	return nil
}

// ExistsPendingRequest checks if there's a pending request between two users
func (r *MatchRequestRepository) ExistsPendingRequest(senderID, receiverID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"status":      domain.MatchRequestStatusPending,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to check pending request", zap.Error(err))
		return false, fmt.Errorf("failed to check pending request: %w", err)
	}

	return count > 0, nil
}
