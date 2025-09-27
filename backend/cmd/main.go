package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eralove/eralove-backend/internal/app"
	"github.com/eralove/eralove-backend/internal/config"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

// @title           EraLove API
// @version         1.0
// @description     EraLove is a couple's love journey tracking application. This API provides endpoints for user authentication, photo management, event tracking, private messaging, and match requests between couples.
// @termsOfService  https://eralove.com/terms
// @contact.name   EraLove API Support
// @contact.url    https://eralove.com/support
// @contact.email  support@eralove.com
// @license.name   MIT
// @license.url    https://opensource.org/licenses/MIT
// @host           localhost:8080
// @BasePath       /api/v1
// @schemes        http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	zap.ReplaceGlobals(logger)

	// Create application with Wire DI
	application, err := app.InitializeApp(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Start server in a separate goroutine
	go func() {
		if err := application.Run(); err != nil {
			logger.Fatal("Error running server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	// Initialize zerolog for structured logging
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if cfg.Environment != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return logger, nil
}
