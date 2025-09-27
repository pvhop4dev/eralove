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
