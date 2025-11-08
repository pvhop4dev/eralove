package handler

import (
	"strconv"
	"strings"

	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// PhotoHandler handles photo-related HTTP requests
type PhotoHandler struct {
	photoService domain.PhotoService
	validator    *validator.Validate
	i18n         *i18n.I18n
	logger       *zap.Logger
}

// NewPhotoHandler creates a new photo handler
func NewPhotoHandler(
	photoService domain.PhotoService,
	validator *validator.Validate,
	i18n *i18n.I18n,
	logger *zap.Logger,
) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
		validator:    validator,
		i18n:         i18n,
		logger:       logger,
	}
}

// CreatePhoto handles photo creation
// @Summary Create a new photo
// @Description Create a new photo with uploaded file path
// @Tags photos
// @Accept json
// @Produce json
// @Param request body domain.CreatePhotoRequest true "Photo data with file_path"
// @Security BearerAuth
// @Success 201 {object} domain.PhotoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /photos [post]
func (h *PhotoHandler) CreatePhoto(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Create photo")
	
	userID := getUserIDFromContext(c)

	// Parse JSON request
	var req domain.CreatePhotoRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Create photo")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, c, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("title", req.Title),
		zap.String("file_path", req.FilePath))

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, c, err, "Create photo", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
			Details: getValidationErrors(err),
		})
	}

	LogServiceCall(h.logger, c, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("file_path", req.FilePath))

	// Create photo
	photo, err := h.photoService.CreatePhotoWithPath(c.Context(), userID, &req)
	if err != nil {
		LogServiceError(h.logger, c, err, "Create photo", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to create photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photo.ID))

	return c.Status(fiber.StatusCreated).JSON(photo)
}

// GetPhotos handles getting user photos
// @Summary Get user photos
// @Description Get photos for the authenticated user
// @Tags photos
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param partner_id query string false "Partner ID to filter shared photos"
// @Security BearerAuth
// @Success 200 {object} domain.PhotoListResponse
// @Failure 401 {object} ErrorResponse
// @Router /photos [get]
func (h *PhotoHandler) GetPhotos(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Get photos")
	
	userID := getUserIDFromContext(c)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	LogRequestParsed(h.logger, c, "Get photos", 
		zap.String("user_id", userID.Hex()),
		zap.Int("page", page),
		zap.Int("limit", limit))

	LogServiceCall(h.logger, c, "Get photos", 
		zap.String("user_id", userID.Hex()))

	photos, total, err := h.photoService.GetCouplePhotos(c.Context(), userID, page, limit)
	if err != nil {
		LogServiceError(h.logger, c, err, "Get photos", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get photos",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Get photos", 
		zap.String("user_id", userID.Hex()),
		zap.Int64("total", total),
		zap.Int("count", len(photos)))

	return c.JSON(domain.PhotoListResponse{
		Photos: photos,
		Total:  total,
		Page:   page,
		Limit:  limit,
	})
}

// GetPhoto handles getting a specific photo
// @Summary Get photo by ID
// @Description Get a specific photo by its ID
// @Tags photos
// @Produce json
// @Param id path string true "Photo ID"
// @Security BearerAuth
// @Success 200 {object} domain.PhotoResponse
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /photos/{id} [get]
func (h *PhotoHandler) GetPhoto(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Get photo")
	
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		LogValidationError(h.logger, c, err, "Get photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id_param", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogServiceCall(h.logger, c, "Get photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	photo, err := h.photoService.GetPhoto(c.Context(), photoID, userID)
	if err != nil {
		LogServiceError(h.logger, c, err, "Get photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "Photo not found",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "not_found", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Get photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	return c.JSON(photo)
}

// UpdatePhoto handles photo updates
// @Summary Update photo
// @Description Update photo information
// @Tags photos
// @Accept json
// @Produce json
// @Param id path string true "Photo ID"
// @Param request body domain.UpdatePhotoRequest true "Photo update data"
// @Security BearerAuth
// @Success 200 {object} domain.PhotoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /photos/{id} [put]
func (h *PhotoHandler) UpdatePhoto(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Update photo")
	
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		LogValidationError(h.logger, c, err, "Update photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id_param", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	var req domain.UpdatePhotoRequest
	if err := c.BodyParser(&req); err != nil {
		LogParsingError(h.logger, err, c, "Update photo")
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid request body",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogRequestParsed(h.logger, c, "Update photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()),
		zap.String("title", req.Title))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, c, err, "Update photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, c, "Update photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	photo, err := h.photoService.UpdatePhoto(c.Context(), photoID, userID, &req)
	if err != nil {
		LogServiceError(h.logger, c, err, "Update photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Update photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	return c.JSON(photo)
}

// DeletePhoto handles photo deletion
// @Summary Delete photo
// @Description Delete a photo
// @Tags photos
// @Produce json
// @Param id path string true "Photo ID"
// @Security BearerAuth
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /photos/{id} [delete]
func (h *PhotoHandler) DeletePhoto(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Delete photo")
	
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		LogValidationError(h.logger, c, err, "Delete photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id_param", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogServiceCall(h.logger, c, "Delete photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	err = h.photoService.DeletePhoto(c.Context(), photoID, userID)
	if err != nil {
		LogServiceError(h.logger, c, err, "Delete photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, c, "Delete photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	return c.SendStatus(fiber.StatusNoContent)
}

// Helper functions
func parseCommaSeparatedTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	// Simple comma separation - could be enhanced with proper parsing
	tags := []string{}
	for _, tag := range strings.Split(tagsStr, ",") {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags
}
