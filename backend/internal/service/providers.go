package service

import (
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/auth"
	"github.com/eralove/eralove-backend/internal/infrastructure/email"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ServiceSet provides all service dependencies
var ServiceSet = wire.NewSet(
	ProvideUserService,
	ProvidePhotoService,
	ProvideEventService,
	ProvideMatchRequestService,
	// TODO: Uncomment when services are fully implemented
	// ProvideMessageService,
)

// ProvideUserService provides a user service
func ProvideUserService(
	userRepo domain.UserRepository,
	eventRepo domain.EventRepository,
	photoRepo domain.PhotoRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	emailService *email.EmailService,
	logger *zap.Logger,
) domain.UserService {
	return NewUserService(userRepo, eventRepo, photoRepo, passwordManager, jwtManager, emailService, logger)
}

// ProvidePhotoService provides a photo service
func ProvidePhotoService(
	photoRepo domain.PhotoRepository,
	userRepo domain.UserRepository,
	storageService domain.StorageService,
	logger *zap.Logger,
) domain.PhotoService {
	return NewPhotoService(photoRepo, userRepo, storageService, logger)
}

// ProvideEventService provides an event service
func ProvideEventService(
	eventRepo domain.EventRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.EventService {
	return NewEventService(eventRepo, userRepo, logger)
}

// ProvideMessageService provides a message service
// func ProvideMessageService(
// 	messageRepo domain.MessageRepository,
// 	logger *zap.Logger,
// ) domain.MessageService {
// 	return NewMessageService(messageRepo, logger)
// }

// ProvideMatchRequestService provides a match request service
func ProvideMatchRequestService(
	matchRequestRepo domain.MatchRequestRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) domain.MatchRequestService {
	return NewMatchRequestService(matchRequestRepo, userRepo, logger)
}
