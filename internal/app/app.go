package app

import (
	"database/sql"
	"log"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/routes"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/database"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/gin-gonic/gin"
)

type Module interface {
	Routes() routes.Route
}

type Application struct {
	config  *config.Config
	router  *gin.Engine
	modules []Module
}

type ModuleContext struct {
	DB *sql.DB
}

func NewApplication(cfg *config.Config) *Application {

	r := gin.Default()

	if err := database.InitDB(); err != nil {
		log.Fatalf("unable to connnect to db: %v", err)
	}

	ctx := &ModuleContext{
		DB: database.DB,
	}

	tokenService := jwt.NewJWTService()
	authRepo := repository.NewAuthRepository(database.DB)

	modules := []Module{
		NewUserModule(ctx),
		NewAuthModule(ctx, tokenService),
		NewAdminModule(cfg),
	}

	routes.RegisterRoutes(r, tokenService, authRepo, getModuleRoutes(modules)...)

	return &Application{
		config:  cfg,
		router:  r,
		modules: modules,
	}
}

func (a *Application) Run() error {
	if a.config.ServerAddress == "" {
		a.config.ServerAddress = ":8080"
	}

	log.Printf(" Server is running at http://localhost%s\n", a.config.ServerAddress)
	return a.router.Run(a.config.ServerAddress)
}

func getModuleRoutes(modules []Module) []routes.Route {
	routeList := make([]routes.Route, len(modules))
	for i, module := range modules {
		routeList[i] = module.Routes()
	}

	return routeList
}
