package handler

import (
	"strconv"

	"github.com/eralove/eralove-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

// CMSHandler handles CMS-related HTTP requests
type CMSHandler struct {
	cmsService *service.CMSService
}

// NewCMSHandler creates a new CMS handler
func NewCMSHandler(cmsService *service.CMSService) *CMSHandler {
	return &CMSHandler{
		cmsService: cmsService,
	}
}

// GetContent retrieves content from a collection
// @Summary Get content from collection
// @Description Get items from a Directus collection
// @Tags CMS
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/{collection} [get]
func (h *CMSHandler) GetContent(c *fiber.Ctx) error {
	collection := c.Params("collection")
	if collection == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "collection parameter is required",
		})
	}

	// Build filters from query params
	filters := make(map[string]string)
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}
	if sort := c.Query("sort"); sort != "" {
		filters["sort"] = sort
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	// Add any filter parameters
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		if len(keyStr) > 7 && keyStr[:7] == "filter[" {
			filters[keyStr] = string(value)
		}
	})

	result, err := h.cmsService.GetContent(collection, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetContentByID retrieves a single content item by ID
// @Summary Get content by ID
// @Description Get a single item from a Directus collection
// @Tags CMS
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Item ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/{collection}/{id} [get]
func (h *CMSHandler) GetContentByID(c *fiber.Ctx) error {
	collection := c.Params("collection")
	id := c.Params("id")

	if collection == "" || id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "collection and id parameters are required",
		})
	}

	result, err := h.cmsService.GetContentByID(collection, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// CreateContent creates new content in a collection
// @Summary Create content
// @Description Create a new item in a Directus collection
// @Tags CMS
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param body body map[string]interface{} true "Content data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/{collection} [post]
// @Security BearerAuth
func (h *CMSHandler) CreateContent(c *fiber.Ctx) error {
	collection := c.Params("collection")
	if collection == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "collection parameter is required",
		})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := h.cmsService.CreateContent(collection, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

// UpdateContent updates existing content
// @Summary Update content
// @Description Update an existing item in a Directus collection
// @Tags CMS
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Item ID"
// @Param body body map[string]interface{} true "Content data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/{collection}/{id} [patch]
// @Security BearerAuth
func (h *CMSHandler) UpdateContent(c *fiber.Ctx) error {
	collection := c.Params("collection")
	id := c.Params("id")

	if collection == "" || id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "collection and id parameters are required",
		})
	}

	var data map[string]interface{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := h.cmsService.UpdateContent(collection, id, data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// DeleteContent deletes content from a collection
// @Summary Delete content
// @Description Delete an item from a Directus collection
// @Tags CMS
// @Accept json
// @Produce json
// @Param collection path string true "Collection name"
// @Param id path string true "Item ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/{collection}/{id} [delete]
// @Security BearerAuth
func (h *CMSHandler) DeleteContent(c *fiber.Ctx) error {
	collection := c.Params("collection")
	id := c.Params("id")

	if collection == "" || id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "collection and id parameters are required",
		})
	}

	if err := h.cmsService.DeleteContent(collection, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetFiles retrieves files from Directus
// @Summary Get files
// @Description Get files from Directus
// @Tags CMS
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/files [get]
func (h *CMSHandler) GetFiles(c *fiber.Ctx) error {
	filters := make(map[string]string)
	if limit := c.Query("limit"); limit != "" {
		filters["limit"] = limit
	}
	if offset := c.Query("offset"); offset != "" {
		filters["offset"] = offset
	}

	result, err := h.cmsService.GetFiles(filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetBlogPosts retrieves blog posts
// @Summary Get blog posts
// @Description Get blog posts from Directus
// @Tags CMS
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Status filter"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/blog/posts [get]
func (h *CMSHandler) GetBlogPosts(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	status := c.Query("status", "published")

	result, err := h.cmsService.GetBlogPosts(limit, offset, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetPages retrieves pages
// @Summary Get pages
// @Description Get pages from Directus
// @Tags CMS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/pages [get]
func (h *CMSHandler) GetPages(c *fiber.Ctx) error {
	result, err := h.cmsService.GetPages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// GetSettings retrieves settings
// @Summary Get settings
// @Description Get settings from Directus
// @Tags CMS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/cms/settings [get]
func (h *CMSHandler) GetSettings(c *fiber.Ctx) error {
	result, err := h.cmsService.GetSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
