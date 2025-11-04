package handler

import (
	"github.com/eralove/eralove-backend/internal/service"
	"github.com/google/wire"
)

// HandlerSet provides all handler dependencies
var HandlerSet = wire.NewSet(
	// ProvideUserHandler,  // TODO: Reimplement with PostgreSQL
	// ProvidePhotoHandler, // TODO: Reimplement with PostgreSQL
	// ProvideUploadHandler, // TODO: Reimplement with PostgreSQL
	ProvideCMSHandler,
	// TODO: Uncomment when services are implemented
	// ProvideEventHandler,
	// ProvideMessageHandler,
	// ProvideMatchRequestHandler,
)

// TODO: Reimplement UserHandler with PostgreSQL UUID
// ProvideUserHandler provides a user handler
// func ProvideUserHandler(
// 	userService domain.UserService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *UserHandler {
// 	return NewUserHandler(userService, validator, i18nService, logger)
// }

// TODO: Reimplement PhotoHandler with PostgreSQL UUID
// ProvidePhotoHandler provides a photo handler
// func ProvidePhotoHandler(
// 	photoService domain.PhotoService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *PhotoHandler {
// 	return NewPhotoHandler(photoService, validator, i18nService, logger)
// }

// TODO: Uncomment when services are implemented
// // ProvideEventHandler provides an event handler
// func ProvideEventHandler(
// 	eventService domain.EventService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *EventHandler {
// 	return NewEventHandler(eventService, validator, i18nService, logger)
// }

// // ProvideMessageHandler provides a message handler
// func ProvideMessageHandler(
// 	messageService domain.MessageService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *MessageHandler {
// 	return NewMessageHandler(messageService, validator, i18nService, logger)
// }

// // ProvideMatchRequestHandler provides a match request handler
// func ProvideMatchRequestHandler(
// 	matchRequestService domain.MatchRequestService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *MatchRequestHandler {
// 	return NewMatchRequestHandler(matchRequestService, validator, i18nService, logger)
// }

// TODO: Reimplement UploadHandler with PostgreSQL UUID
// ProvideUploadHandler provides an upload handler
// func ProvideUploadHandler(
// 	storageService domain.StorageService,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *UploadHandler {
// 	return NewUploadHandler(storageService, i18nService, logger)
// }

// ProvideCMSHandler provides a CMS handler
func ProvideCMSHandler(cmsService *service.CMSService) *CMSHandler {
	return NewCMSHandler(cmsService)
}
