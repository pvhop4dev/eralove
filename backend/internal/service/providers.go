package service

import (
	"github.com/eralove/eralove-backend/internal/infrastructure/directus"
	"github.com/google/wire"
)

// ServiceSet provides all service dependencies
var ServiceSet = wire.NewSet(
	// ProvideUserService,  // TODO: Reimplement with PostgreSQL
	// ProvidePhotoService, // TODO: Reimplement with PostgreSQL
	ProvideCMSService,
	// TODO: Uncomment when services are fully implemented
	// ProvideEventService,
	// ProvideMessageService,
	// ProvideMatchRequestService,
)

// TODO: Reimplement UserService with PostgreSQL UUID
// ProvideUserService provides a user service
// func ProvideUserService(
// 	userRepo domain.UserRepository,
// 	passwordManager *auth.PasswordManager,
// 	jwtManager *auth.JWTManager,
// 	emailService *email.EmailService,
// 	logger *zap.Logger,
// ) domain.UserService {
// 	return NewUserService(userRepo, passwordManager, jwtManager, emailService, logger)
// }

// TODO: Reimplement PhotoService with PostgreSQL UUID
// ProvidePhotoService provides a photo service
// func ProvidePhotoService(
// 	photoRepo domain.PhotoRepository,
// 	userRepo domain.UserRepository,
// 	storageService domain.StorageService,
// 	logger *zap.Logger,
// ) domain.PhotoService {
// 	return NewPhotoService(photoRepo, userRepo, storageService, logger)
// }

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

// ProvideCMSService provides a CMS service
func ProvideCMSService(directusClient *directus.Client) *CMSService {
	return NewCMSService(directusClient)
}
