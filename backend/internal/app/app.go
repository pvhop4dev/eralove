package app

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/handler"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/cache"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"github.com/eralove/eralove-backend/internal/infrastructure/email"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/eralove/eralove-backend/internal/repository"
	"github.com/eralove/eralove-backend/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	_ "github.com/eralove/eralove-backend/docs"
)

// App represents the application
type App struct {
	fiber  *fiber.App
	config *config.Config
	logger *zap.Logger
	db     *database.MongoDB
	cache  *cache.Redis
}

// Dependencies represents all application dependencies
type Dependencies struct {
	UserHandler         *handler.UserHandler
	PhotoHandler        *handler.PhotoHandler
	EventHandler        *handler.EventHandler
	MessageHandler      *handler.MessageHandler
	MatchRequestHandler *handler.MatchRequestHandler
}

// NewWithDependencies creates a new application instance with injected dependencies
func NewWithDependencies(cfg *config.Config, logger *zap.Logger, deps *Dependencies) (*App, error) {
	// Initialize database
	db, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create indexes
	if err := db.CreateIndexes(context.Background()); err != nil {
		logger.Warn("Failed to create database indexes", zap.Error(err))
	}

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
		db:     db,
		cache:  redis,
	}, nil
}

// New creates a new application instance (legacy method for backward compatibility)
func New(cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize database
	db, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create indexes
	if err := db.CreateIndexes(context.Background()); err != nil {
		logger.Warn("Failed to create database indexes", zap.Error(err))
	}

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
		IdleTimeout:  120 * time.Second,
	})

	// Initialize dependencies
	validator := validator.New()
	i18nService := i18n.NewI18n(logger)
	
	// Load translation messages
	if err := i18nService.LoadMessages("./messages"); err != nil {
		logger.Warn("Failed to load translation messages", zap.Error(err))
	}
	
	emailService := email.NewEmailService(cfg, logger)

	// Initialize auth managers
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiration, cfg.JWTRefreshExpiration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Database, logger)
	// Note: PhotoRepository is now handled by Wire dependency injection

	// Initialize services
	userService := service.NewUserService(userRepo, passwordManager, jwtManager, emailService, logger)
	// Note: PhotoService is now handled by Wire dependency injection

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, validator, i18nService, logger)

	// Setup middleware
	setupMiddleware(app, cfg, logger)

	// Setup routes
	setupRoutes(app, userHandler, jwtManager, logger)

	return &App{
		fiber:  app,
		config: cfg,
		logger: logger,
		db:     db,
		cache:  redis,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	a.logger.Info("Starting server", zap.String("port", a.config.Port))
	return a.fiber.Listen(a.config.GetPort())
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down server...")

	// Shutdown Fiber
	if err := a.fiber.ShutdownWithContext(ctx); err != nil {
		a.logger.Error("Error shutting down Fiber", zap.Error(err))
	}

	// Close database connection
	if err := a.db.Close(ctx); err != nil {
		a.logger.Error("Error closing database connection", zap.Error(err))
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
	app.Use(requestid.New())

	// Logger middleware
	if cfg.IsDevelopment() {
		app.Use(fiberlogger.New(fiberlogger.Config{
			Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
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
		// Allow all origins in development
		corsConfig.AllowOrigins = "*"
		corsConfig.AllowCredentials = false // Cannot use credentials with wildcard origin
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

// setupRoutes configures application routes
func setupRoutes(app *fiber.App, userHandler *handler.UserHandler, jwtManager *auth.JWTManager, logger *zap.Logger) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)
	auth.Post("/refresh", userHandler.RefreshToken)
	auth.Post("/logout", userHandler.Logout)
	auth.Post("/verify-email", userHandler.VerifyEmail)
	auth.Post("/resend-verification", userHandler.ResendVerificationEmail)
	auth.Post("/forgot-password", userHandler.ForgotPassword)
	auth.Post("/reset-password", userHandler.ResetPassword)

	// Protected routes (authentication required)
	protected := api.Group("/", jwtMiddleware(jwtManager, logger))

	// User routes
	users := protected.Group("/users")
	users.Get("/profile", userHandler.GetProfile)
	users.Put("/profile", userHandler.UpdateProfile)
	users.Delete("/account", userHandler.DeleteAccount)

	// Photo routes (placeholder - handlers need to be created)
	// photos := protected.Group("/photos")
	// photos.Post("/", photoHandler.CreatePhoto)
	// photos.Get("/", photoHandler.GetPhotos)
	// photos.Get("/:id", photoHandler.GetPhoto)
	// photos.Put("/:id", photoHandler.UpdatePhoto)
	// photos.Delete("/:id", photoHandler.DeletePhoto)

	// Event routes (placeholder - handlers need to be created)
	// events := protected.Group("/events")
	// events.Post("/", eventHandler.CreateEvent)
	// events.Get("/", eventHandler.GetEvents)
	// events.Get("/:id", eventHandler.GetEvent)
	// events.Put("/:id", eventHandler.UpdateEvent)
	// events.Delete("/:id", eventHandler.DeleteEvent)

	// Message routes (placeholder - handlers need to be created)
	// messages := protected.Group("/messages")
	// messages.Post("/", messageHandler.SendMessage)
	// messages.Get("/", messageHandler.GetMessages)
	// messages.Get("/conversations", messageHandler.GetConversations)
	// messages.Post("/mark-read", messageHandler.MarkAsRead)
	// messages.Delete("/:id", messageHandler.DeleteMessage)

	// Match request routes (placeholder - handlers need to be created)
	// matchRequests := protected.Group("/match-requests")
	// matchRequests.Post("/", matchRequestHandler.SendMatchRequest)
	// matchRequests.Get("/sent", matchRequestHandler.GetSentRequests)
	// matchRequests.Get("/received", matchRequestHandler.GetReceivedRequests)
	// matchRequests.Get("/:id", matchRequestHandler.GetMatchRequest)
	// matchRequests.Post("/:id/respond", matchRequestHandler.RespondToMatchRequest)
	// matchRequests.Delete("/:id", matchRequestHandler.CancelMatchRequest)
}

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

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (no authentication required)
	auth := api.Group("/auth")
	auth.Post("/register", deps.UserHandler.Register)
	auth.Post("/login", deps.UserHandler.Login)
	auth.Post("/refresh", deps.UserHandler.RefreshToken)
	auth.Post("/logout", deps.UserHandler.Logout)

	// Protected routes (authentication required)
	protected := api.Group("/", jwtMiddleware(jwtManager, logger))

	// User routes
	users := protected.Group("/users")
	users.Get("/profile", deps.UserHandler.GetProfile)
	users.Put("/profile", deps.UserHandler.UpdateProfile)
	users.Delete("/account", deps.UserHandler.DeleteAccount)

	// Photo routes (when handlers are available)
	photos := protected.Group("/photos")
	photos.Post("/", deps.PhotoHandler.CreatePhoto)
	photos.Get("/", deps.PhotoHandler.GetPhotos)
	photos.Get("/:id", deps.PhotoHandler.GetPhoto)
	photos.Put("/:id", deps.PhotoHandler.UpdatePhoto)
	photos.Delete("/:id", deps.PhotoHandler.DeletePhoto)

	// Event routes (when handlers are available)
	events := protected.Group("/events")
	events.Post("/", deps.EventHandler.CreateEvent)
	events.Get("/", deps.EventHandler.GetEvents)
	events.Get("/:id", deps.EventHandler.GetEvent)
	events.Put("/:id", deps.EventHandler.UpdateEvent)
	events.Delete("/:id", deps.EventHandler.DeleteEvent)

	// Message routes (when handlers are available)
	messages := protected.Group("/messages")
	messages.Post("/", deps.MessageHandler.SendMessage)
	messages.Get("/", deps.MessageHandler.GetMessages)
	messages.Get("/conversations", deps.MessageHandler.GetConversations)
	messages.Post("/mark-read", deps.MessageHandler.MarkAsRead)
	messages.Delete("/:id", deps.MessageHandler.DeleteMessage)

	// Match request routes (when handlers are available)
	matchRequests := protected.Group("/match-requests")
	matchRequests.Post("/", deps.MatchRequestHandler.SendMatchRequest)
	matchRequests.Get("/sent", deps.MatchRequestHandler.GetSentRequests)
	matchRequests.Get("/received", deps.MatchRequestHandler.GetReceivedRequests)
	matchRequests.Get("/:id", deps.MatchRequestHandler.GetMatchRequest)
	matchRequests.Post("/:id/respond", deps.MatchRequestHandler.RespondToMatchRequest)
	matchRequests.Delete("/:id", deps.MatchRequestHandler.CancelMatchRequest)
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
			userID, err := primitive.ObjectIDFromHex(userIDStr)
			if err != nil {
				logger.Error("Invalid user ID in token", zap.Error(err))
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid token",
				})
			}

			c.Locals("user_id", userID)
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
