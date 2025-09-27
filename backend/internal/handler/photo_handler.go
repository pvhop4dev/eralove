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
	userID := getUserIDFromContext(c)

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "File is required",
			Message: "Please provide a photo file",
		})
	}

	// Parse form data
	req := domain.CreatePhotoRequest{
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Tags:        parseCommaSeparatedTags(c.FormValue("tags")),
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	// Create photo
	photo, err := h.photoService.CreatePhoto(c.Context(), userID, &req, file)
	if err != nil {
		h.logger.Error("Failed to create photo", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to create photo",
			Message: err.Error(),
		})
	}

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

	photos, total, err := h.photoService.GetUserPhotos(c.Context(), userID, partnerID, page, limit)
	if err != nil {
		h.logger.Error("Failed to get photos", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to get photos",
			Message: err.Error(),
		})
	}

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
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: "Photo ID must be a valid ObjectID",
		})
	}

	photo, err := h.photoService.GetPhoto(c.Context(), photoID, userID)
	if err != nil {
		h.logger.Error("Failed to get photo", zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error:   "Photo not found",
			Message: err.Error(),
		})
	}

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
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: "Photo ID must be a valid ObjectID",
		})
	}

	var req domain.UpdatePhotoRequest
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

	photo, err := h.photoService.UpdatePhoto(c.Context(), photoID, userID, &req)
	if err != nil {
		h.logger.Error("Failed to update photo", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to update photo",
			Message: err.Error(),
		})
	}

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
	userID := getUserIDFromContext(c)
	photoID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "Invalid photo ID",
			Message: "Photo ID must be a valid ObjectID",
		})
	}

	err = h.photoService.DeletePhoto(c.Context(), photoID, userID)
	if err != nil {
		h.logger.Error("Failed to delete photo", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "Failed to delete photo",
			Message: err.Error(),
		})
	}

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
