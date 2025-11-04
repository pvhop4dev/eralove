package app

import (
	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/handler"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"github.com/eralove/eralove-backend/internal/infrastructure/directus"
	"github.com/eralove/eralove-backend/internal/service"
	"go.uber.org/zap"
)

// ProvidePostgresDB provides PostgreSQL database connection
func ProvidePostgresDB(cfg *config.Config, logger *zap.Logger) (*database.PostgresDB, error) {
	pgConfig := database.PostgresConfig{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		Database: cfg.PostgresDB,
		SSLMode:  cfg.PostgresSSLMode,
	}

	db, err := database.NewPostgresDB(pgConfig)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to PostgreSQL",
		zap.String("host", cfg.PostgresHost),
		zap.String("database", cfg.PostgresDB))

	return db, nil
}

// ProvideDirectusClient provides Directus client
func ProvideDirectusClient(cfg *config.Config, logger *zap.Logger) (*directus.Client, error) {
	directusConfig := &directus.Config{
		URL:         cfg.DirectusURL,
		AdminEmail:  cfg.DirectusAdminEmail,
		AdminPass:   cfg.DirectusAdminPass,
		StaticToken: cfg.DirectusStaticToken,
	}

	client, err := directus.NewClientFromConfig(directusConfig)
	if err != nil {
		logger.Error("Failed to create Directus client", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to Directus",
		zap.String("url", cfg.DirectusURL))

	return client, nil
}

// ProvideCMSService provides CMS service
func ProvideCMSService(directusClient *directus.Client) *service.CMSService {
	return service.NewCMSService(directusClient)
}

// ProvideCMSHandler provides CMS handler
func ProvideCMSHandler(cmsService *service.CMSService) *handler.CMSHandler {
	return handler.NewCMSHandler(cmsService)
}
