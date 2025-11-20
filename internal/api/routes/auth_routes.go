package routes

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	handler *handlers.AuthHandler
}

func NewAuthRoutes(handler *handlers.AuthHandler) *AuthRoutes {
	return &AuthRoutes{
		handler: handler,
	}
}

func (ur *AuthRoutes) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", ur.handler.CreateUser)
		auth.POST("/login", ur.handler.Login)
		auth.POST("/login/totp", ur.handler.LoginTOTP)
		
		auth.POST("/logout", ur.handler.Logout)
		
	}
}
