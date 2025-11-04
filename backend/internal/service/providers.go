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
	// TODO: Uncomment when services are fully implemented
	// ProvideEventService,
	// ProvideMessageService,
	// ProvideMatchRequestService,
)

// ProvideUserService provides a user service
func ProvideUserService(
	userRepo domain.UserRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
	emailService *email.EmailService,
	logger *zap.Logger,
) domain.UserService {
	return NewUserService(userRepo, passwordManager, jwtManager, emailService, logger)
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

// TODO: Uncomment when services are fully implemented
// // ProvideEventService provides an event service
// func ProvideEventService(
// 	eventRepo domain.EventRepository,
// 	logger *zap.Logger,
// ) domain.EventService {
// 	return NewEventService(eventRepo, logger)
// }

// // ProvideMessageService provides a message service
// func ProvideMessageService(
// 	messageRepo domain.MessageRepository,
// 	logger *zap.Logger,
// ) domain.MessageService {
// 	return NewMessageService(messageRepo, logger)
// }

// // ProvideMatchRequestService provides a match request service
// func ProvideMatchRequestService(
// 	matchRequestRepo domain.MatchRequestRepository,
// 	logger *zap.Logger,
// ) domain.MatchRequestService {
// 	return NewMatchRequestService(matchRequestRepo, logger)
// }
