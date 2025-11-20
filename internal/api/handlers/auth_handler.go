package handlers

import (
	"net/http"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/validation"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth_service service.AuthService
}

func NewAuthHandler(auth_service service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth_service: auth_service,
	}
}

func (uh *AuthHandler) CreateUser(ctx *gin.Context) {
	var user domain.UserCreate
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Required fields are missing",
		})
		return
	}

	createdUser, err := uh.auth_service.CreateUser(user.Username, user.Password, user.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"userId":  createdUser.Id,
	})
}

func (ah *AuthHandler) Login(ctx *gin.Context) {
	var input domain.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	user, accessToken, expiresIn, err := ah.auth_service.Login(input.Email, input.Password)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	if user.EnableTOTP {
		ctx.JSON(http.StatusOK, gin.H{
			"requireTOTP": true,
			"message":     "TOTP verification required",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
		"expiresIn":   expiresIn,
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (ah *AuthHandler) Logout(ctx *gin.Context) {
	err := ah.auth_service.Logout(ctx)

	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "User logged out", nil)

}

func (ah *AuthHandler) LoginTOTP(ctx *gin.Context) {
	var input domain.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	user, accessToken, _, err := ah.auth_service.Login(input.Email, input.Password)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
		},
	})
}
