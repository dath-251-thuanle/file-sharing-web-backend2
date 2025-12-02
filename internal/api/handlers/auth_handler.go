package handlers

import (
	"net/http"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
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

	user, token, err := ah.auth_service.Login(input.Email, input.Password)
	if err != nil {
		err.Export(ctx)
		return
	}

	if user.EnableTOTP {
		ctx.JSON(http.StatusOK, gin.H{
			"requireTOTP": user.EnableTOTP,
			"id":          user.Id,
			"cid":         token,
			"message":     "TOTP verification required",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": token,
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
		err.Export(ctx)
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "User logged out", nil)
}

func getUserIDFromContext(c *gin.Context) (string, bool) {
	userObj, exists := c.Get("user")
	if !exists {
		return "", false
	}

	claims, ok := userObj.(*jwt.Claims)
	if !ok {
		return "", false
	}

	return claims.UserID, true
}

func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	if authErr, exists := c.Get("authError"); exists {
		switch authErr {
		case "required":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
		case "invalid":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired access token",
			})
		}
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or expired access token",
		})
		return
	}

	resp, err := h.auth_service.SetupTOTP(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "TOTP secret generated",
		"totpSetup": resp,
	})
}

type VerifyTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	if authErr, exists := c.Get("authError"); exists {
		switch authErr {
		case "required":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
		case "invalid":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired access token",
			})
		}
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or expired access token",
		})
		return
	}

	var req VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid TOTP code",
			"message": "The provided code is incorrect or expired",
		})
		return
	}

	okVerify, err := h.auth_service.VerifyTOTP(userID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !okVerify {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or expired access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "TOTP verified successfully",
		"totpEnabled": true,
	})
}

func (ah *AuthHandler) LoginTOTP(ctx *gin.Context) {
	var input domain.LoginTOTPInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	user, accessToken, err := ah.auth_service.LoginTOTP(input.ID, input.TOTPCode)
	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
		"user": gin.H{
			"id":       user.Id,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}
