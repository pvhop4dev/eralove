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
	ProvideEventRepository,
	ProvideMatchRequestRepository,
	// TODO: Uncomment when repositories are implemented
	// ProvideMessageRepository,
)

// ProvideUserRepository provides a user repository
func ProvideUserRepository(db *database.MongoDB, logger *zap.Logger) domain.UserRepository {
	return NewUserRepository(db.Database, logger)
}

// ProvidePhotoRepository provides a photo repository
func ProvidePhotoRepository(db *database.MongoDB, logger *zap.Logger) domain.PhotoRepository {
	return NewPhotoRepositoryWithMatchCode(db.Database, logger)
}

// ProvideEventRepository provides an event repository
func ProvideEventRepository(db *database.MongoDB, logger *zap.Logger) domain.EventRepository {
	return NewEventRepository(db.Database, logger)
}

// // ProvideMessageRepository provides a message repository
// func ProvideMessageRepository(db *database.MongoDB, logger *zap.Logger) domain.MessageRepository {
// 	return NewMessageRepository(db.Database, logger)
// }

// ProvideMatchRequestRepository provides a match request repository
func ProvideMatchRequestRepository(db *database.MongoDB, logger *zap.Logger) domain.MatchRequestRepository {
	return NewMatchRequestRepository(db.Database, logger)
}
