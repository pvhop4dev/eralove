package infrastructure

import (
	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/cache"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"github.com/eralove/eralove-backend/internal/infrastructure/email"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/eralove/eralove-backend/internal/infrastructure/storage"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// InfrastructureSet provides all infrastructure dependencies
var InfrastructureSet = wire.NewSet(
	ProvideValidator,
	ProvideI18n,
	ProvidePasswordManager,
	ProvideJWTManager,
	ProvideEmailService,
	ProvideMongoDB,
	ProvideRedis,
	ProvideStorageService,
)

// ProvideValidator provides a validator instance
func ProvideValidator() *validator.Validate {
	return validator.New()
}

// ProvideI18n provides an i18n service
func ProvideI18n(logger *zap.Logger) *i18n.I18n {
	i18nService := i18n.NewI18n(logger)
	
	// Load translation messages from messages directory
	// Try multiple paths to handle different working directories
	paths := []string{
		"./messages",
		"../messages",
		"./backend/messages",
		"messages",
	}
	
	var loaded bool
	for _, path := range paths {
		if err := i18nService.LoadMessages(path); err == nil {
			logger.Info("Translation messages loaded successfully", zap.String("path", path))
			loaded = true
			break
		}
	}
	
	if !loaded {
		logger.Warn("Failed to load i18n messages from any path")
	}
	
	return i18nService
}

// ProvidePasswordManager provides a password manager
func ProvidePasswordManager() *auth.PasswordManager {
	return auth.NewPasswordManager()
}

// ProvideJWTManager provides a JWT manager
func ProvideJWTManager(cfg *config.Config) *auth.JWTManager {
	return auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiration, cfg.JWTRefreshExpiration)
}

// ProvideMongoDB provides a MongoDB connection
func ProvideMongoDB(cfg *config.Config, logger *zap.Logger) (*database.MongoDB, error) {
	return database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName, logger)
}

// ProvideRedis provides a Redis connection
func ProvideRedis(cfg *config.Config, logger *zap.Logger) (*cache.Redis, error) {
	return cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, logger)
}

// ProvideEmailService provides an email service
func ProvideEmailService(cfg *config.Config, logger *zap.Logger) *email.EmailService {
	return email.NewEmailService(cfg, logger)
}

// ProvideStorageService provides a storage service
func ProvideStorageService(cfg *config.Config, logger *zap.Logger) (domain.StorageService, error) {
	// Create storage configuration from config
	storageConfig := &domain.StorageConfig{
		Provider:        cfg.StorageProvider,
		Region:          cfg.StorageRegion,
		Bucket:          cfg.StorageBucket,
		AccessKeyID:     cfg.StorageAccessKeyID,
		SecretAccessKey: cfg.StorageSecretKey,
		Endpoint:        cfg.StorageEndpoint,
		UseSSL:          cfg.StorageUseSSL,
		BaseURL:         cfg.StorageBaseURL,
	}

	// Create storage factory
	factory := storage.NewFactory(logger)

	// Create and return storage service
	return factory.CreateStorage(storageConfig)
}
