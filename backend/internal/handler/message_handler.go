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

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	messageService domain.MessageService
	validator      *validator.Validate
	i18n           *i18n.I18n
	logger         *zap.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	messageService domain.MessageService,
	validator *validator.Validate,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		validator:      validator,
		i18n:           i18n,
		logger:         logger,
	}
}

// SendMessage handles message sending
// @Summary Send a message
// @Description Send a message to partner
// @Tags messages
// @Accept json
// @Produce json
// @Param request body domain.CreateMessageRequest true "Message data"
// @Security BearerAuth
// @Success 201 {object} domain.MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /messages [post]
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	
	var req domain.CreateMessageRequest
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

	message, err := h.messageService.SendMessage(c.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to send message",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to send message",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(message)
}

// GetMessages handles getting conversation messages
// @Summary Get conversation messages
// @Description Get messages between user and partner
// @Tags messages
// @Produce json
// @Param partner_id query string true "Partner ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security BearerAuth
// @Success 200 {object} domain.MessageListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /messages [get]
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	
	partnerIDStr := c.Query("partner_id")
	if partnerIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Partner ID is required",
			Message: "Please provide partner_id query parameter",
		})
	}

	partnerID, err := primitive.ObjectIDFromHex(partnerIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid partner ID",
			Message: "Partner ID must be a valid ObjectID",
		})
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	messages, total, err := h.messageService.GetConversation(c.Context(), userID, partnerID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get messages",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get messages",
			Message: err.Error(),
		})
	}

	return c.JSON(domain.MessageListResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// GetConversations handles getting user conversations
// @Summary Get user conversations
// @Description Get all conversations for the authenticated user
// @Tags messages
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security BearerAuth
// @Success 200 {object} domain.ConversationListResponse
// @Failure 401 {object} ErrorResponse
// @Router /messages/conversations [get]
func (h *MessageHandler) GetConversations(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	
	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	conversations, total, err := h.messageService.GetUserConversations(c.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get conversations",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get conversations",
			Message: err.Error(),
		})
	}

	return c.JSON(domain.ConversationListResponse{
		Conversations: conversations,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

// MarkAsRead handles marking messages as read
// @Summary Mark messages as read
// @Description Mark messages in a conversation as read
// @Tags messages
// @Accept json
// @Produce json
// @Param request body domain.MarkAsReadRequest true "Mark as read data"
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /messages/mark-read [post]
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	
	var req domain.MarkAsReadRequest
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

	err := h.messageService.MarkAsRead(c.Context(), userID, req.PartnerID)
	if err != nil {
		h.logger.Error("Failed to mark messages as read",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to mark messages as read",
			Message: err.Error(),
		})
	}

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Messages marked as read",
	})
}

// DeleteMessage handles message deletion
// @Summary Delete message
// @Description Delete a message (soft delete)
// @Tags messages
// @Produce json
// @Param id path string true "Message ID"
// @Security BearerAuth
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	userID := getUserIDFromContext(c)
	messageID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid message ID",
			Message: "Message ID must be a valid ObjectID",
		})
	}

	err = h.messageService.DeleteMessage(c.Context(), messageID, userID)
	if err != nil {
		h.logger.Error("Failed to delete message",
			zap.String("trace_id", getTraceID(c)),
			zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete message",
			Message: err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
