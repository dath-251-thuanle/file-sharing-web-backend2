package app

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/routes"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/storage"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
)

type fileModule struct {
	routes routes.Route
}

func NewFileModule(
	cfg *config.Config,
	fileRepo repository.FileRepository,
	sharedRepo repository.SharedRepository,
	userRepo repository.UserRepository,
	storageService storage.Storage,
) Module {
	fileService := service.NewFileService(cfg, fileRepo, sharedRepo, userRepo, storageService)
	fileHandler := handlers.NewFileHandler(fileService)
	fileRoutes := routes.NewFileRoutes(fileHandler)

	return &fileModule{
		routes: fileRoutes,
	}
}

func (m *fileModule) Routes() routes.Route {
	return m.routes
}
