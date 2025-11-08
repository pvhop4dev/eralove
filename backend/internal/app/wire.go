//go:build wireinject
// +build wireinject

package app

import (
	"github.com/eralove/eralove-backend/internal/config"
	"github.com/eralove/eralove-backend/internal/domain"
	"github.com/eralove/eralove-backend/internal/handler"
	"github.com/eralove/eralove-backend/internal/infrastructure"
	"github.com/eralove/eralove-backend/internal/repository"
	"github.com/eralove/eralove-backend/internal/service"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ApplicationSet combines all provider sets
var ApplicationSet = wire.NewSet(
	infrastructure.InfrastructureSet,
	repository.RepositorySet,
	service.ServiceSet,
	handler.HandlerSet,
	ProvideDependencies,
	ProvideApp,
)

// InitializeApp creates a new application with all dependencies injected
func InitializeApp(cfg *config.Config, logger *zap.Logger) (*App, error) {
	wire.Build(ApplicationSet)
	return &App{}, nil
}

// ProvideDependencies creates the dependencies struct
func ProvideDependencies(
	userHandler *handler.UserHandler,
	photoHandler *handler.PhotoHandler,
	uploadHandler *handler.UploadHandler,
	storageService domain.StorageService,
	eventHandler *handler.EventHandler,
	matchRequestHandler *handler.MatchRequestHandler,
	// TODO: Add when implemented
	// messageHandler *handler.MessageHandler,
) *Dependencies {
	return &Dependencies{
		UserHandler:         userHandler,
		PhotoHandler:        photoHandler,
		UploadHandler:       uploadHandler,
		StorageService:      storageService,
		EventHandler:        eventHandler,
		MatchRequestHandler: matchRequestHandler,
		// TODO: Add when implemented
		// MessageHandler: messageHandler,
	}
}

// ProvideApp creates the main application
func ProvideApp(cfg *config.Config, logger *zap.Logger, deps *Dependencies) (*App, error) {
	return NewWithDependencies(cfg, logger, deps)
}
