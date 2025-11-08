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

// EventRepository implements domain.EventRepository
type EventRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *mongo.Database, logger *zap.Logger) domain.EventRepository {
	return &EventRepository{
		collection: db.Collection("events"),
		logger:     logger,
	}
}

// Create creates a new event
func (r *EventRepository) Create(event *domain.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		r.logger.Error("Failed to create event", zap.Error(err))
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(id primitive.ObjectID) (*domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var event domain.Event
	err := r.collection.FindOne(ctx, bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&event)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("event not found")
		}
		r.logger.Error("Failed to get event by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return &event, nil
}

// GetByMatchCode retrieves events for a match code
func (r *EventRepository) GetByMatchCode(matchCode string, limit, offset int) ([]*domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"match_code": matchCode,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get events by match code", zap.Error(err))
		return nil, fmt.Errorf("failed to get events by match code: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*domain.Event
	if err := cursor.All(ctx, &events); err != nil {
		r.logger.Error("Failed to decode events", zap.Error(err))
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetByMatchCodeAndDateRange retrieves events within a date range for a match code
func (r *EventRepository) GetByMatchCodeAndDateRange(matchCode string, startDate, endDate time.Time) ([]*domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"match_code": matchCode,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get events by match code and date range", zap.Error(err))
		return nil, fmt.Errorf("failed to get events by match code and date range: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*domain.Event
	if err := cursor.All(ctx, &events); err != nil {
		r.logger.Error("Failed to decode events", zap.Error(err))
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetByMatchCodeAndDate retrieves events for a specific date for a match code
func (r *EventRepository) GetByMatchCodeAndDate(matchCode string, date time.Time) ([]*domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	opts := options.Find().SetSort(bson.D{{Key: "time", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get events by match code and date", zap.Error(err))
		return nil, fmt.Errorf("failed to get events by match code and date: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*domain.Event
	if err := cursor.All(ctx, &events); err != nil {
		r.logger.Error("Failed to decode events", zap.Error(err))
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// GetUpcomingByMatchCode retrieves upcoming events for a match code
func (r *EventRepository) GetUpcomingByMatchCode(matchCode string, limit int) ([]*domain.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()

	filter := bson.M{
		"match_code": matchCode,
		"date":       bson.M{"$gte": now},
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error("Failed to get upcoming events by match code", zap.Error(err))
		return nil, fmt.Errorf("failed to get upcoming events by match code: %w", err)
	}
	defer cursor.Close(ctx)

	var events []*domain.Event
	if err := cursor.All(ctx, &events); err != nil {
		r.logger.Error("Failed to decode events", zap.Error(err))
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	return events, nil
}

// DeleteByMatchCode deletes all events for a match code (for unmatch)
func (r *EventRepository) DeleteByMatchCode(matchCode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"match_code": matchCode,
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete events by match code", zap.Error(err))
		return fmt.Errorf("failed to delete events by match code: %w", err)
	}

	return nil
}

// Update updates an event
func (r *EventRepository) Update(id primitive.ObjectID, event *domain.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":        id,
		"deleted_at": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": event,
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update event", zap.Error(err))
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// Delete soft deletes an event
func (r *EventRepository) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
		r.logger.Error("Failed to delete event", zap.Error(err))
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}
