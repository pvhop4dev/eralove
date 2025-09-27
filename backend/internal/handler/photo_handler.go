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
// @Description Upload and create a new photo
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Photo file"
// @Param title formData string false "Photo title"
// @Param description formData string false "Photo description"
// @Param tags formData string false "Photo tags (comma separated)"
// @Security BearerAuth
// @Success 201 {object} domain.PhotoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /photos [post]
func (h *PhotoHandler) CreatePhoto(c *fiber.Ctx) error {
	LogRequestStart(h.logger, c, "Create photo")
	
	userID := getUserIDFromContext(c)

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		LogRequestError(h.logger, c, "File upload failed", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "File is required",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	// Parse form data
	req := domain.CreatePhotoRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Tags:        parseCommaSeparatedTags(c.FormValue("tags")),
	}

	LogRequestParsed(h.logger, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("title", req.Title),
		zap.String("filename", file.Filename))

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Create photo", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("filename", file.Filename))

	// Create photo
	photo, err := h.photoService.CreatePhoto(c.Context(), userID, &req, file)
	if err != nil {
		LogServiceError(h.logger, err, "Create photo", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to create photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Create photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photo.ID.Hex()))

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
	partnerIDStr := c.Query("partner_id")

	var partnerID *primitive.ObjectID
	if partnerIDStr != "" {
		if id, err := primitive.ObjectIDFromHex(partnerIDStr); err == nil {
			partnerID = &id
		}
	}

	LogRequestParsed(h.logger, "Get photos", 
		zap.String("user_id", userID.Hex()),
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.String("partner_id", partnerIDStr))

	LogServiceCall(h.logger, "Get photos", 
		zap.String("user_id", userID.Hex()))

	photos, total, err := h.photoService.GetUserPhotos(c.Context(), userID, partnerID, page, limit)
	if err != nil {
		LogServiceError(h.logger, err, "Get photos", 
			zap.String("user_id", userID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get photos",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Get photos", 
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
		LogValidationError(h.logger, err, "Get photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id_param", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogServiceCall(h.logger, "Get photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	photo, err := h.photoService.GetPhoto(c.Context(), photoID, userID)
	if err != nil {
		LogServiceError(h.logger, err, "Get photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "Photo not found",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "not_found", nil),
		})
	}

	LogServiceSuccess(h.logger, "Get photo", 
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
		LogValidationError(h.logger, err, "Update photo", 
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

	LogRequestParsed(h.logger, "Update photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()),
		zap.String("title", req.Title))

	if err := h.validator.Struct(&req); err != nil {
		LogValidationError(h.logger, err, "Update photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "validation_failed", nil),
		})
	}

	LogServiceCall(h.logger, "Update photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	photo, err := h.photoService.UpdatePhoto(c.Context(), photoID, userID, &req)
	if err != nil {
		LogServiceError(h.logger, err, "Update photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Update photo", 
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
		LogValidationError(h.logger, err, "Delete photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id_param", c.Params("id")))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "invalid_request", nil),
		})
	}

	LogServiceCall(h.logger, "Delete photo", 
		zap.String("user_id", userID.Hex()),
		zap.String("photo_id", photoID.Hex()))

	err = h.photoService.DeletePhoto(c.Context(), photoID, userID)
	if err != nil {
		LogServiceError(h.logger, err, "Delete photo", 
			zap.String("user_id", userID.Hex()),
			zap.String("photo_id", photoID.Hex()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete photo",
			Message: h.i18n.Translate(c.Get("Accept-Language", "en"), "internal_error", nil),
		})
	}

	LogServiceSuccess(h.logger, "Delete photo", 
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
