package app

import (
	"context"
	"time"

	"strings"

	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/handler"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/cache"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	_ "github.com/eralove/eralove-backend/docs"
)

// App represents the application
type App struct {
	fiber  *fiber.App
	config *config.Config
	logger *zap.Logger
	cache  *cache.Redis
}

// Dependencies represents all application dependencies
type Dependencies struct {
	// UserHandler         *handler.UserHandler         // TODO: Reimplement
	// PhotoHandler        *handler.PhotoHandler        // TODO: Reimplement
	// EventHandler        *handler.EventHandler        // TODO: Implement
	// MessageHandler      *handler.MessageHandler      // TODO: Implement
	// MatchRequestHandler *handler.MatchRequestHandler // TODO: Implement
	// UploadHandler       *handler.UploadHandler       // TODO: Reimplement
	CMSHandler *handler.CMSHandler
}

// NewWithDependencies creates a new application instance with injected dependencies
func NewWithDependencies(cfg *config.Config, logger *zap.Logger, deps *Dependencies) (*App, error) {
	// Initialize cache
	redis, err := cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, logger)
	if err != nil {
		logger.Warn("Failed to connect to Redis", zap.Error(err))
		// Continue without Redis for now
		redis = nil
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// Initialize JWT manager for middleware
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiration, cfg.JWTRefreshExpiration)

	// Setup middleware
	setupMiddleware(app, cfg, logger)

	// Setup routes with injected dependencies
	setupRoutesWithDeps(app, deps, jwtManager, logger)

	return &App{
		fiber:  app,
		config: cfg,
		logger: logger,
		cache:  redis,
	}, nil
}

// Legacy New function removed - use Wire DI via InitializeApp instead

// Run starts the application
func (a *App) Run() error {
	addr := "0.0.0.0" + a.config.GetPort()
	a.logger.Info("Starting server",
		zap.String("address", addr),
		zap.String("port", a.config.Port))
	return a.fiber.Listen(addr)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	// Shutdown Fiber
	if err := a.fiber.ShutdownWithContext(ctx); err != nil {
		a.logger.Error("Error shutting down Fiber", zap.Error(err))
	}

	// Close Redis connection
	if a.cache != nil {
		if err := a.cache.Close(); err != nil {
			a.logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}

	return nil
}

// setupMiddleware configures middleware
func setupMiddleware(app *fiber.App, cfg *config.Config, logger *zap.Logger) {
	// Request ID middleware
	app.Use(requestid.New(requestid.Config{
		Header:     "X-Request-ID",
		ContextKey: "requestid",
	}))

	// Logger middleware with trace ID
	if cfg.IsDevelopment() {
		app.Use(fiberlogger.New(fiberlogger.Config{
			Format:     "[${time}] [${locals:requestid}] ${status} - ${method} ${path} - ${latency}\n",
			TimeFormat: "2006-01-02 15:04:05",
		}))
	} else {
		app.Use(fiberlogger.New(fiberlogger.Config{
			Format:     "[${time}] [${locals:requestid}] ${ip} ${status} - ${method} ${path} - ${latency} ${error}\n",
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		}))
	}

	// Recovery middleware
	app.Use(recover.New())

	// CORS middleware
	corsConfig := cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,Access-Control-Allow-Origin",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type",
	}

	if cfg.IsDevelopment() {
		// Allow localhost and 127.0.0.1 in development
		corsConfig.AllowOrigins = "http://localhost:3000,http://localhost:5173,http://localhost:8080,http://127.0.0.1:3000,http://127.0.0.1:5173,http://127.0.0.1:8080"
		corsConfig.AllowCredentials = true
	} else {
		// Use specific origins in production
		origins := strings.Split(cfg.CORSOrigins, ",")
		corsConfig.AllowOrigins = strings.Join(origins, ",")
	}

	app.Use(cors.New(corsConfig))

	// Handle preflight requests
	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
}

// Legacy setupRoutes removed - use setupRoutesWithDeps instead

// setupRoutesWithDeps configures application routes with injected dependencies
func setupRoutesWithDeps(app *fiber.App, deps *Dependencies, jwtManager *auth.JWTManager, logger *zap.Logger) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Static file serving for uploaded files (public access - no auth required)
	// Register at app level to completely bypass all middleware
	app.Static("/api/v1/files", "./uploads")

	// API routes
	api := app.Group("/api/v1")

	// TODO: Reimplement Auth routes with PostgreSQL UUID
	// auth := api.Group("/auth")
	// auth.Post("/register", deps.UserHandler.Register)
	// auth.Post("/login", deps.UserHandler.Login)
	// auth.Post("/refresh", deps.UserHandler.RefreshToken)
	// auth.Post("/logout", deps.UserHandler.Logout)

	// Protected routes (authentication required)
	// Use specific path instead of "/" to avoid catching all routes
	protected := api.Group("")
	protected.Use(jwtMiddleware(jwtManager, logger))

	// TODO: Reimplement User routes with PostgreSQL UUID
	// users := protected.Group("/users")
	// users.Get("/profile", deps.UserHandler.GetProfile)
	// users.Put("/profile", deps.UserHandler.UpdateProfile)
	// users.Delete("/account", deps.UserHandler.DeleteAccount)

	// TODO: Reimplement Photo routes with PostgreSQL UUID
	// photos := protected.Group("/photos")
	// photos.Post("/", deps.PhotoHandler.CreatePhoto)
	// photos.Get("/", deps.PhotoHandler.GetPhotos)
	// photos.Get("/:id", deps.PhotoHandler.GetPhoto)
	// photos.Put("/:id", deps.PhotoHandler.UpdatePhoto)
	// photos.Delete("/:id", deps.PhotoHandler.DeletePhoto)

	// TODO: Uncomment when handlers are implemented
	// // Event routes
	// events := protected.Group("/events")
	// events.Post("/", deps.EventHandler.CreateEvent)
	// events.Get("/", deps.EventHandler.GetEvents)
	// events.Get("/:id", deps.EventHandler.GetEvent)
	// events.Put("/:id", deps.EventHandler.UpdateEvent)
	// events.Delete("/:id", deps.EventHandler.DeleteEvent)

	// // Message routes
	// messages := protected.Group("/messages")
	// messages.Post("/", deps.MessageHandler.SendMessage)
	// messages.Get("/", deps.MessageHandler.GetMessages)
	// messages.Get("/conversations", deps.MessageHandler.GetConversations)
	// messages.Post("/mark-read", deps.MessageHandler.MarkAsRead)
	// messages.Delete("/:id", deps.MessageHandler.DeleteMessage)

	// // Match request routes
	// matchRequests := protected.Group("/match-requests")
	// matchRequests.Post("/", deps.MatchRequestHandler.SendMatchRequest)
	// matchRequests.Get("/sent", deps.MatchRequestHandler.GetSentRequests)
	// matchRequests.Get("/received", deps.MatchRequestHandler.GetReceivedRequests)
	// matchRequests.Get("/:id", deps.MatchRequestHandler.GetMatchRequest)
	// matchRequests.Post("/:id/respond", deps.MatchRequestHandler.RespondToMatchRequest)
	// matchRequests.Delete("/:id", deps.MatchRequestHandler.CancelMatchRequest)

	// TODO: Reimplement Upload routes with PostgreSQL UUID
	// upload := protected.Group("/upload")
	// upload.Post("/", deps.UploadHandler.UploadFile)
	// upload.Post("/multiple", deps.UploadHandler.UploadMultipleFiles)
	// upload.Delete("/", deps.UploadHandler.DeleteFile)

	// CMS routes (if CMSHandler is available)
	if deps.CMSHandler != nil {
		cms := api.Group("/cms")

		// Public routes
		cms.Get("/blog/posts", deps.CMSHandler.GetBlogPosts)
		cms.Get("/pages", deps.CMSHandler.GetPages)
		cms.Get("/settings", deps.CMSHandler.GetSettings)
		cms.Get("/files", deps.CMSHandler.GetFiles)
		cms.Get("/:collection", deps.CMSHandler.GetContent)
		cms.Get("/:collection/:id", deps.CMSHandler.GetContentByID)

		// Protected CMS routes
		cmsProtected := cms.Use(jwtMiddleware(jwtManager, logger))
		cmsProtected.Post("/:collection", deps.CMSHandler.CreateContent)
		cmsProtected.Patch("/:collection/:id", deps.CMSHandler.UpdateContent)
		cmsProtected.Delete("/:collection/:id", deps.CMSHandler.DeleteContent)
	}
}

// jwtMiddleware creates JWT authentication middleware
func jwtMiddleware(jwtManager *auth.JWTManager, logger *zap.Logger) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: []byte(jwtManager.GetSecretKey()),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Warn("JWT authentication failed", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Invalid or missing token",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			// Extract user info from token and set in context
			token := c.Locals("user").(*jwt.Token)
			claims := token.Claims.(jwt.MapClaims)

			userIDStr := claims["user_id"].(string)
			// UUID string - no conversion needed
			c.Locals("user_id", userIDStr)
			c.Locals("user_email", claims["email"])
			c.Locals("user_name", claims["name"])

			return c.Next()
		},
	})
}

// errorHandler handles application errors
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"message": message,
	})
}
