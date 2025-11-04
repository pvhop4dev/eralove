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
	// TODO: Uncomment when repositories are implemented
	// ProvideEventRepository,
	// ProvideMessageRepository,
	// ProvideMatchRequestRepository,
)

// ProvideUserRepository provides a user repository (PostgreSQL)
func ProvideUserRepository(db *database.PostgresDB, logger *zap.Logger) domain.UserRepository {
	return NewUserRepositoryPostgres(db, logger)
}

// ProvidePhotoRepository provides a photo repository (PostgreSQL)
func ProvidePhotoRepository(db *database.PostgresDB, logger *zap.Logger) domain.PhotoRepository {
	return NewPhotoRepositoryPostgres(db, logger)
}
