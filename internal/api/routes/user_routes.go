package routes

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	handler *handlers.UserHandler
}

func NewUserRoutes(handler *handlers.UserHandler) *UserRoutes {
	return &UserRoutes{
		handler: handler,
	}
}

func (ur *UserRoutes) Register(r *gin.RouterGroup) {
	users := r.Group("/user")
	{
		users.GET("/:id", ur.handler.GetUserById)
		users.GET("", ur.handler.GetUserById)
	}
}
