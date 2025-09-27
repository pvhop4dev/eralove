package service

import (
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ServiceSet provides all service dependencies
var ServiceSet = wire.NewSet(
	ProvideUserService,
	ProvidePhotoService,
)

// ProvideUserService provides a user service
func ProvideUserService(
	userRepo domain.UserRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	logger *zap.Logger,
) domain.UserService {
	return NewUserService(userRepo, passwordManager, jwtManager, logger)
}

// ProvidePhotoService provides a photo service
func ProvidePhotoService(
	photoRepo domain.PhotoRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.PhotoService {
	return NewPhotoService(photoRepo, userRepo, logger)
}
