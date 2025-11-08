package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eralove/eralove-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// EventService implements domain.EventService
type EventService struct {
	eventRepo domain.EventRepository
	userRepo  domain.UserRepository
	logger    *zap.Logger
}

// NewEventService creates a new event service
func NewEventService(
	eventRepo domain.EventRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.EventService {
	return &EventService{
		eventRepo: eventRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// CreateEvent creates a new event
func (s *EventService) CreateEvent(
	ctx context.Context,
	userID primitive.ObjectID,
	req *domain.CreateEventRequest,
) (*domain.EventResponse, error) {
	s.logger.Info("Creating event",
		zap.String("user_id", userID.Hex()),
		zap.String("title", req.Title),
		zap.String("event_type", req.EventType))

	// Get user to get match code
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.MatchCode == "" {
		s.logger.Error("User is not matched")
		return nil, fmt.Errorf("user is not matched with anyone")
	}

	// Create event
	event := &domain.Event{
		ID:             primitive.NewObjectID(),
		MatchCode:      user.MatchCode,
		CreatedBy:      userID, // Track who created this event
		Title:          req.Title,
		Description:    req.Description,
		Date:           req.Date,
		Time:           req.Time,
		Location:       req.Location,
		EventType:      req.EventType,
		IsRecurring:    req.IsRecurring,
		RecurrenceRule: req.RecurrenceRule,
		IsPrivate:      req.IsPrivate,
		Reminder:       req.Reminder,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save to database
	if err := s.eventRepo.Create(event); err != nil {
		s.logger.Error("Failed to create event", zap.Error(err))
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	s.logger.Info("Event created successfully",
		zap.String("event_id", event.ID.Hex()),
		zap.String("user_id", userID.Hex()))

	return event.ToResponse(), nil
}

// GetEvent retrieves a specific event
func (s *EventService) GetEvent(
	ctx context.Context,
	eventID, userID primitive.ObjectID,
) (*domain.EventResponse, error) {
	s.logger.Info("Getting event",
		zap.String("event_id", eventID.Hex()),
		zap.String("user_id", userID.Hex()))

	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		s.logger.Error("Failed to get event", zap.Error(err))
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Get user to verify match code
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user has access to this event
	if event.MatchCode != user.MatchCode {
		s.logger.Warn("Unauthorized access to event",
			zap.String("event_id", eventID.Hex()),
			zap.String("user_id", userID.Hex()))
		return nil, fmt.Errorf("unauthorized access to event")
	}

	return event.ToResponse(), nil
}

// GetCoupleEvents retrieves events for a couple with optional filtering
func (s *EventService) GetCoupleEvents(
	ctx context.Context,
	userID primitive.ObjectID,
	year, month, page, limit int,
) ([]*domain.EventResponse, int64, error) {
	s.logger.Info("Getting couple events",
		zap.String("user_id", userID.Hex()),
		zap.Int("year", year),
		zap.Int("month", month),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Get user to get match code
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get user: %w", err)
	}

	if user.MatchCode == "" {
		s.logger.Info("User is not matched, returning empty events")
		return []*domain.EventResponse{}, 0, nil
	}

	var events []*domain.Event

	// If year and month are specified, filter by date range
	if year > 0 && month > 0 {
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
		
		events, err = s.eventRepo.GetByMatchCodeAndDateRange(user.MatchCode, startDate, endDate)
	} else {
		// Get all couple events
		offset := (page - 1) * limit
		events, err = s.eventRepo.GetByMatchCode(user.MatchCode, limit, offset)
	}

	if err != nil {
		s.logger.Error("Failed to get couple events", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get events: %w", err)
	}

	// Convert to responses
	responses := make([]*domain.EventResponse, len(events))
	for i, event := range events {
		responses[i] = event.ToResponse()
	}

	total := int64(len(responses))

	s.logger.Info("Retrieved couple events",
		zap.String("user_id", userID.Hex()),
		zap.Int64("total", total))

	return responses, total, nil
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(
	ctx context.Context,
	eventID, userID primitive.ObjectID,
	req *domain.UpdateEventRequest,
) (*domain.EventResponse, error) {
	s.logger.Info("Updating event",
		zap.String("event_id", eventID.Hex()),
		zap.String("user_id", userID.Hex()))

	// Get existing event
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		s.logger.Error("Failed to get event for update", zap.Error(err))
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Get user to verify match code
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check ownership
	if event.MatchCode != user.MatchCode {
		s.logger.Warn("Unauthorized update attempt",
			zap.String("event_id", eventID.Hex()),
			zap.String("user_id", userID.Hex()))
		return nil, fmt.Errorf("unauthorized to update this event")
	}

	// Update fields
	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.Date != nil {
		event.Date = *req.Date
	}
	if req.Time != "" {
		event.Time = req.Time
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if req.EventType != "" {
		event.EventType = req.EventType
	}
	if req.IsRecurring != nil {
		event.IsRecurring = *req.IsRecurring
	}
	if req.RecurrenceRule != "" {
		event.RecurrenceRule = req.RecurrenceRule
	}
	if req.IsPrivate != nil {
		event.IsPrivate = *req.IsPrivate
	}
	if req.Reminder != nil {
		event.Reminder = req.Reminder
	}

	event.UpdatedAt = time.Now()

	// Save updates
	if err := s.eventRepo.Update(eventID, event); err != nil {
		s.logger.Error("Failed to update event", zap.Error(err))
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	s.logger.Info("Event updated successfully",
		zap.String("event_id", eventID.Hex()))

	return event.ToResponse(), nil
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(
	ctx context.Context,
	eventID, userID primitive.ObjectID,
) error {
	s.logger.Info("Deleting event",
		zap.String("event_id", eventID.Hex()),
		zap.String("user_id", userID.Hex()))

	// Get event to check ownership
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		s.logger.Error("Failed to get event for deletion", zap.Error(err))
		return fmt.Errorf("event not found: %w", err)
	}

	// Get user to verify match code
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check ownership
	if event.MatchCode != user.MatchCode {
		s.logger.Warn("Unauthorized delete attempt",
			zap.String("event_id", eventID.Hex()),
			zap.String("user_id", userID.Hex()))
		return fmt.Errorf("unauthorized to delete this event")
	}

	// Delete event
	if err := s.eventRepo.Delete(eventID); err != nil {
		s.logger.Error("Failed to delete event", zap.Error(err))
		return fmt.Errorf("failed to delete event: %w", err)
	}

	s.logger.Info("Event deleted successfully",
		zap.String("event_id", eventID.Hex()))

	return nil
}
