package routes

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/middleware"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/gin-gonic/gin"
)

type Route interface {
	Register(r *gin.RouterGroup)
}

func RegisterRoutes(r *gin.Engine, authService jwt.TokenService, authRepo repository.AuthRepository, routes ...Route) {

	api := r.Group("/")

	middleware.InitAuthMiddleware(authService, authRepo)

	protected := api.Group("")

	protected.Use(
		middleware.AuthMiddleware(),
	)

	for _, route := range routes {
		switch route.(type) {
		case *AuthRoutes:
			route.Register(api)
		case *FileRoutes:
			route.Register(api)
		default:
			route.Register(protected)
		}
	}

	home := func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to File Sharing API",
			"status":  "online",
		})
	}

	r.GET("/", home)
}
