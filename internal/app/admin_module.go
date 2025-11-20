// file: internal/app/admin_module.go (Mới)
package app

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/routes"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/storage"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
)

type adminModule struct {
	routes routes.Route
}

// Cần thêm các tham số FileRepo và Storage để hỗ trợ Cleanup
func NewAdminModule(
	cfg *config.Config,
	fileRepo repository.FileRepository, // <-- THÊM
	storageService storage.Storage, // <-- THÊM
) Module {

	// Policy tĩnh: không cần Repository
	adminService := service.NewAdminService(cfg, fileRepo, storageService) // <-- CẬP NHẬT
	adminHandler := handlers.NewAdminHandler(adminService)
	adminRoutes := routes.NewAdminRoutes(adminHandler)

	return &adminModule{
		routes: adminRoutes,
	}
}

func (m *adminModule) Routes() routes.Route {
	return m.routes
}
