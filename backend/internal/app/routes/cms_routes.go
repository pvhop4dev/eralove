package routes

import (
	"github.com/eralove/eralove-backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// SetupCMSRoutes sets up CMS-related routes
func SetupCMSRoutes(app *fiber.App, cmsHandler *handler.CMSHandler) {
	cms := app.Group("/api/v1/cms")

	// Public routes
	cms.Get("/blog/posts", cmsHandler.GetBlogPosts)
	cms.Get("/pages", cmsHandler.GetPages)
	cms.Get("/settings", cmsHandler.GetSettings)
	cms.Get("/files", cmsHandler.GetFiles)

	// Generic collection routes (public read)
	cms.Get("/:collection", cmsHandler.GetContent)
	cms.Get("/:collection/:id", cmsHandler.GetContentByID)

	// Protected routes (require authentication)
	// Note: Add JWT middleware here when implementing auth
	// cmsProtected := cms.Use(jwtMiddleware)
	cms.Post("/:collection", cmsHandler.CreateContent)
	cms.Patch("/:collection/:id", cmsHandler.UpdateContent)
	cms.Delete("/:collection/:id", cmsHandler.DeleteContent)
}
