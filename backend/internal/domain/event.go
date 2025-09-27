package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event in the system
type Event struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
	PartnerID   *primitive.ObjectID `json:"partner_id,omitempty" bson:"partner_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required,min=1,max=200"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Date        time.Time          `json:"date" bson:"date"`
	Time        string             `json:"time,omitempty" bson:"time,omitempty"`
	Location    string             `json:"location,omitempty" bson:"location,omitempty"`
	EventType   string             `json:"event_type" bson:"event_type" validate:"required,oneof=anniversary date milestone celebration other"`
	IsRecurring bool               `json:"is_recurring" bson:"is_recurring"`
	RecurrenceRule string          `json:"recurrence_rule,omitempty" bson:"recurrence_rule,omitempty"`
	IsPrivate   bool               `json:"is_private" bson:"is_private"`
	Reminder    *EventReminder     `json:"reminder,omitempty" bson:"reminder,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt   *time.Time         `json:"-" bson:"deleted_at,omitempty"`
}

// EventReminder represents reminder settings for an event
type EventReminder struct {
	Enabled     bool      `json:"enabled" bson:"enabled"`
	ReminderAt  time.Time `json:"reminder_at" bson:"reminder_at"`
	Message     string    `json:"message,omitempty" bson:"message,omitempty"`
	IsNotified  bool      `json:"is_notified" bson:"is_notified"`
}

// CreateEventRequest represents the request to create a new event
type CreateEventRequest struct {
	Title          string         `json:"title" validate:"required,min=1,max=200"`
	Description    string         `json:"description,omitempty"`
	Date           time.Time      `json:"date" validate:"required"`
	Time           string         `json:"time,omitempty"`
	Location       string         `json:"location,omitempty"`
	EventType      string         `json:"event_type" validate:"required,oneof=anniversary date milestone celebration other"`
	IsRecurring    bool           `json:"is_recurring"`
	RecurrenceRule string         `json:"recurrence_rule,omitempty"`
	IsPrivate      bool           `json:"is_private"`
	Reminder       *EventReminder `json:"reminder,omitempty"`
}

// UpdateEventRequest represents the request to update an event
type UpdateEventRequest struct {
	Title          string         `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description    string         `json:"description,omitempty"`
	Date           *time.Time     `json:"date,omitempty"`
	Time           string         `json:"time,omitempty"`
	Location       string         `json:"location,omitempty"`
	EventType      string         `json:"event_type,omitempty" validate:"omitempty,oneof=anniversary date milestone celebration other"`
	IsRecurring    *bool          `json:"is_recurring,omitempty"`
	RecurrenceRule string         `json:"recurrence_rule,omitempty"`
	IsPrivate      *bool          `json:"is_private,omitempty"`
	Reminder       *EventReminder `json:"reminder,omitempty"`
}

// EventResponse represents the event response
type EventResponse struct {
	ID             primitive.ObjectID `json:"id"`
	UserID         primitive.ObjectID `json:"user_id"`
	PartnerID      *primitive.ObjectID `json:"partner_id,omitempty"`
	Title          string             `json:"title"`
	Description    string             `json:"description,omitempty"`
	Date           time.Time          `json:"date"`
	Time           string             `json:"time,omitempty"`
	Location       string             `json:"location,omitempty"`
	EventType      string             `json:"event_type"`
	IsRecurring    bool               `json:"is_recurring"`
	RecurrenceRule string             `json:"recurrence_rule,omitempty"`
	IsPrivate      bool               `json:"is_private"`
	Reminder       *EventReminder     `json:"reminder,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ToResponse converts Event to EventResponse
func (e *Event) ToResponse() *EventResponse {
	return &EventResponse{
		ID:             e.ID,
		UserID:         e.UserID,
		PartnerID:      e.PartnerID,
		Title:          e.Title,
		Description:    e.Description,
		Date:           e.Date,
		Time:           e.Time,
		Location:       e.Location,
		EventType:      e.EventType,
		IsRecurring:    e.IsRecurring,
		RecurrenceRule: e.RecurrenceRule,
		IsPrivate:      e.IsPrivate,
		Reminder:       e.Reminder,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

// EventListResponse represents a list of events response
type EventListResponse struct {
	Events []*EventResponse `json:"events"`
	Total  int64            `json:"total"`
	Page   int              `json:"page"`
	Limit  int              `json:"limit"`
}

// EventRepository defines the interface for event data access
type EventRepository interface {
	Create(event *Event) error
	GetByID(id primitive.ObjectID) (*Event, error)
	GetByUserID(userID primitive.ObjectID, limit, offset int) ([]*Event, error)
	GetByCoupleID(userID, partnerID primitive.ObjectID, limit, offset int) ([]*Event, error)
	GetByDateRange(userID primitive.ObjectID, startDate, endDate time.Time) ([]*Event, error)
	GetByDate(userID primitive.ObjectID, date time.Time) ([]*Event, error)
	GetUpcoming(userID primitive.ObjectID, limit int) ([]*Event, error)
	Update(id primitive.ObjectID, event *Event) error
	Delete(id primitive.ObjectID) error
}

// EventService defines the interface for event business logic
type EventService interface {
	CreateEvent(ctx context.Context, userID primitive.ObjectID, req *CreateEventRequest) (*EventResponse, error)
	GetEvent(ctx context.Context, eventID, userID primitive.ObjectID) (*EventResponse, error)
	GetUserEvents(ctx context.Context, userID primitive.ObjectID, partnerID *primitive.ObjectID, year, month, page, limit int) ([]*EventResponse, int64, error)
	UpdateEvent(ctx context.Context, eventID, userID primitive.ObjectID, req *UpdateEventRequest) (*EventResponse, error)
	DeleteEvent(ctx context.Context, eventID, userID primitive.ObjectID) error
}
