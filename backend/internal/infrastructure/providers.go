package infrastructure

import (
	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/cache"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
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
	ProvideMongoDB,
	ProvideRedis,
)

// ProvideValidator provides a validator instance
func ProvideValidator() *validator.Validate {
	return validator.New()
}

// ProvideI18n provides an i18n service
func ProvideI18n(logger *zap.Logger) *i18n.I18n {
	return i18n.NewI18n(logger)
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
