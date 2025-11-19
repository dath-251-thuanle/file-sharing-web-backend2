package routes

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

type TotpRoutes struct {
	handler *handlers.TotpHandler
}

func NewTotpRoutes(handler *handlers.TotpHandler) *TotpRoutes {
	return &TotpRoutes{handler: handler}
}

func (r *TotpRoutes) Register(group *gin.RouterGroup) {
	totp := group.Group("/auth/totp")
	{
		totp.POST("/setup", r.handler.SetupTOTP)
		totp.POST("/verify", r.handler.VerifyTOTP)
	}
}
