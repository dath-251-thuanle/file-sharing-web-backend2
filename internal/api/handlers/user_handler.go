package handlers

import (
	"net/http"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	user_service service.UserService
}

func NewUserHandler(user_service service.UserService) *UserHandler {
	return &UserHandler{
		user_service: user_service,
	}
}

func (uh *UserHandler) GetUserById(ctx *gin.Context) {
	id := ctx.Param("id")

	var user domain.User
	createdUser, err := uh.user_service.GetUserById(id)
	if err != nil {
		err.Export(ctx)
		return
	} else {
		user = *createdUser
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (uh *UserHandler) GetUserByEmail(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
		return
	}
	var user domain.UserResponse
	createdUser, err := uh.user_service.GetUserByEmail(email)
	if err != nil {
		err.Export(ctx)
		return
	} else {
		user = *createdUser
	}
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}
