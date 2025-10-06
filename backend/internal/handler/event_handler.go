package handler

import (
	"strconv"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	eventService domain.EventService
	validator    *validator.Validate
	i18n         *i18n.I18n
	logger       *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	eventService domain.EventService,
	validator *validator.Validate,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *EventHandler {
	return &EventHandler{
		eventService: eventService,
		validator:    validator,
		i18n:         i18n,
		logger:       logger,
	}
}

// CreateEvent handles event creation
// @Summary Create a new event
// @Description Create a new event/milestone
// @Tags events
// @Accept json
// @Produce json
// @Param request body domain.CreateEventRequest true "Event creation data"
// @Security BearerAuth
// @Success 201 {object} domain.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /events [post]
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)

	var req domain.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	event, err := h.eventService.CreateEvent(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create event",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to create event",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(event)
}

// GetEvents handles getting user events
// @Summary Get user events
// @Description Get events for the authenticated user
// @Tags events
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param partner_id query string false "Partner ID to filter shared events"
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month"
// @Security BearerAuth
// @Success 200 {object} domain.EventListResponse
// @Failure 401 {object} ErrorResponse
// @Router /events [get]
func (h *EventHandler) GetEvents(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	year, _ := strconv.Atoi(c.Query("year", "0"))
	month, _ := strconv.Atoi(c.Query("month", "0"))
	partnerIDStr := c.Query("partner_id")

	var partnerID *primitive.ObjectID
	if partnerIDStr != "" {
		if id, err := primitive.ObjectIDFromHex(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	events, total, err := h.eventService.GetUserEvents(c.Context(), userID, partnerID, year, month, page, limit)
	if err != nil {
		h.logger.Error("Failed to get events",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get events",
			Message: err.Error(),
		})
	}

	return c.JSON(domain.EventListResponse{
		Events: events,
		Total:  total,
		Page:   page,
		Limit:  limit,
	})
}

// GetEvent handles getting a specific event
// @Summary Get event by ID
// @Description Get a specific event by its ID
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Security BearerAuth
// @Success 200 {object} domain.EventResponse
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /events/{id} [get]
func (h *EventHandler) GetEvent(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	eventID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid event ID",
			Message: "Event ID must be a valid ObjectID",
		})
	}

	event, err := h.eventService.GetEvent(c.Context(), eventID, userID)
	if err != nil {
		h.logger.Error("Failed to get event",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "Event not found",
			Message: err.Error(),
		})
	}

	return c.JSON(event)
}

// UpdateEvent handles event updates
// @Summary Update event
// @Description Update event information
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param request body domain.UpdateEventRequest true "Event update data"
// @Security BearerAuth
// @Success 200 {object} domain.EventResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /events/{id} [put]
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	eventID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid event ID",
			Message: "Event ID must be a valid ObjectID",
		})
	}

	var req domain.UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	event, err := h.eventService.UpdateEvent(c.Context(), eventID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to update event",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update event",
			Message: err.Error(),
		})
	}

	return c.JSON(event)
}

// DeleteEvent handles event deletion
// @Summary Delete event
// @Description Delete an event
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Security BearerAuth
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	eventID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid event ID",
			Message: "Event ID must be a valid ObjectID",
		})
	}

	err = h.eventService.DeleteEvent(c.Context(), eventID, userID)
	if err != nil {
		h.logger.Error("Failed to delete event",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete event",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
