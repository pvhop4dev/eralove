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

// MatchRequestHandler handles match request-related HTTP requests
type MatchRequestHandler struct {
	matchRequestService domain.MatchRequestService
	validator           *validator.Validate
	i18n                *i18n.I18n
	logger              *zap.Logger
}

// NewMatchRequestHandler creates a new match request handler
func NewMatchRequestHandler(
	matchRequestService domain.MatchRequestService,
	validator *validator.Validate,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *MatchRequestHandler {
	return &MatchRequestHandler{
		matchRequestService: matchRequestService,
		validator:           validator,
		i18n:                i18n,
		logger:              logger,
	}
}

// SendMatchRequest handles sending match requests
// @Summary Send a match request
// @Description Send a match request to another user by email
// @Tags match-requests
// @Accept json
// @Produce json
// @Param request body domain.CreateMatchRequestRequest true "Match request data"
// @Security BearerAuth
// @Success 201 {object} domain.MatchRequestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-requests [post]
func (h *MatchRequestHandler) SendMatchRequest(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)

	var req domain.CreateMatchRequestRequest
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

	matchRequest, err := h.matchRequestService.SendMatchRequest(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to send match request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to send match request",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(matchRequest)
}

// GetSentRequests handles getting sent match requests
// @Summary Get sent match requests
// @Description Get match requests sent by the authenticated user
// @Tags match-requests
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (pending, accepted, rejected)"
// @Security BearerAuth
// @Success 200 {object} domain.MatchRequestListResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-requests/sent [get]
func (h *MatchRequestHandler) GetSentRequests(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status")

	requests, total, err := h.matchRequestService.GetSentRequests(c.Context(), userID, status, page, limit)
	if err != nil {
		h.logger.Error("Failed to get sent requests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get sent requests",
			Message: err.Error(),
		})
	}

	return c.JSON(domain.MatchRequestListResponse{
		MatchRequests: requests,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

// GetReceivedRequests handles getting received match requests
// @Summary Get received match requests
// @Description Get match requests received by the authenticated user
// @Tags match-requests
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (pending, accepted, rejected)"
// @Security BearerAuth
// @Success 200 {object} domain.MatchRequestListResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-requests/received [get]
func (h *MatchRequestHandler) GetReceivedRequests(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status")

	requests, total, err := h.matchRequestService.GetReceivedRequests(c.Context(), userID, status, page, limit)
	if err != nil {
		h.logger.Error("Failed to get received requests", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get received requests",
			Message: err.Error(),
		})
	}

	return c.JSON(domain.MatchRequestListResponse{
		MatchRequests: requests,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

// RespondToMatchRequest handles responding to match requests
// @Summary Respond to match request
// @Description Accept or reject a match request
// @Tags match-requests
// @Accept json
// @Produce json
// @Param id path string true "Match Request ID"
// @Param request body domain.RespondToMatchRequestRequest true "Response data"
// @Security BearerAuth
// @Success 200 {object} domain.MatchRequestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /match-requests/{id}/respond [post]
func (h *MatchRequestHandler) RespondToMatchRequest(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	requestID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request ID",
			Message: "Request ID must be a valid ObjectID",
		})
	}

	var req domain.RespondToMatchRequestRequest
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

	matchRequest, err := h.matchRequestService.RespondToMatchRequest(c.Context(), requestID, userID, req.Action)
	if err != nil {
		h.logger.Error("Failed to respond to match request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to respond to match request",
			Message: err.Error(),
		})
	}

	return c.JSON(matchRequest)
}

// GetMatchRequest handles getting a specific match request
// @Summary Get match request by ID
// @Description Get a specific match request by its ID
// @Tags match-requests
// @Produce json
// @Param id path string true "Match Request ID"
// @Security BearerAuth
// @Success 200 {object} domain.MatchRequestResponse
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-requests/{id} [get]
func (h *MatchRequestHandler) GetMatchRequest(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	requestID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request ID",
			Message: "Request ID must be a valid ObjectID",
		})
	}

	matchRequest, err := h.matchRequestService.GetMatchRequest(c.Context(), requestID, userID)
	if err != nil {
		h.logger.Error("Failed to get match request", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "Match request not found",
			Message: err.Error(),
		})
	}

	return c.JSON(matchRequest)
}

// CancelMatchRequest handles canceling match requests
// @Summary Cancel match request
// @Description Cancel a pending match request
// @Tags match-requests
// @Produce json
// @Param id path string true "Match Request ID"
// @Security BearerAuth
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /match-requests/{id} [delete]
func (h *MatchRequestHandler) CancelMatchRequest(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	requestID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request ID",
			Message: "Request ID must be a valid ObjectID",
		})
	}

	err = h.matchRequestService.CancelMatchRequest(c.Context(), requestID, userID)
	if err != nil {
		h.logger.Error("Failed to cancel match request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to cancel match request",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

