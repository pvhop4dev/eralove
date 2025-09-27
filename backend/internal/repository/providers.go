package repository

import (
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/database"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// RepositorySet provides all repository dependencies
var RepositorySet = wire.NewSet(
	ProvideUserRepository,
	ProvidePhotoRepository,
)

// ProvideUserRepository provides a user repository
func ProvideUserRepository(db *database.MongoDB, logger *zap.Logger) domain.UserRepository {
	return NewUserRepository(db.Database, logger)
}

// ProvidePhotoRepository provides a photo repository
func ProvidePhotoRepository(db *database.MongoDB, logger *zap.Logger) domain.PhotoRepository {
	return NewPhotoRepository(db.Database, logger)
}
