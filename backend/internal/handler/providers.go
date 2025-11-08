package handler

import (
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/infrastructure/i18n"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// HandlerSet provides all handler dependencies
var HandlerSet = wire.NewSet(
	ProvideUserHandler,
	ProvidePhotoHandler,
	ProvideEventHandler,
	ProvideMatchRequestHandler,
	ProvideUploadHandler,
	// TODO: Uncomment when services are implemented
	// ProvideMessageHandler,
)

// ProvideUserHandler provides a user handler
func ProvideUserHandler(
	userService domain.UserService,
	validator *validator.Validate,
	i18nService *i18n.I18n,
	logger *zap.Logger,
) *UserHandler {
	return NewUserHandler(userService, validator, i18nService, logger)
}

// ProvidePhotoHandler provides a photo handler
func ProvidePhotoHandler(
	photoService domain.PhotoService,
	validator *validator.Validate,
	i18nService *i18n.I18n,
	logger *zap.Logger,
) *PhotoHandler {
	return NewPhotoHandler(photoService, validator, i18nService, logger)
}

// ProvideEventHandler provides an event handler
func ProvideEventHandler(
	eventService domain.EventService,
	validator *validator.Validate,
	i18nService *i18n.I18n,
	logger *zap.Logger,
) *EventHandler {
	return NewEventHandler(eventService, validator, i18nService, logger)
}

// // ProvideMessageHandler provides a message handler
// func ProvideMessageHandler(
// 	messageService domain.MessageService,
// 	validator *validator.Validate,
// 	i18nService *i18n.I18n,
// 	logger *zap.Logger,
// ) *MessageHandler {
// 	return NewMessageHandler(messageService, validator, i18nService, logger)
// }

// ProvideMatchRequestHandler provides a match request handler
func ProvideMatchRequestHandler(
	matchRequestService domain.MatchRequestService,
	validator *validator.Validate,
	i18nService *i18n.I18n,
	logger *zap.Logger,
) *MatchRequestHandler {
	return NewMatchRequestHandler(matchRequestService, validator, i18nService, logger)
}

// ProvideUploadHandler provides an upload handler
func ProvideUploadHandler(
	storageService domain.StorageService,
	i18nService *i18n.I18n,
	logger *zap.Logger,
) *UploadHandler {
	return NewUploadHandler(storageService, i18nService, logger)
}
