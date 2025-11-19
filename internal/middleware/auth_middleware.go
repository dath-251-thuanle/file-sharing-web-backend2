package middleware

import (
	"net/http"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/gin-gonic/gin"
)

var (
	jwtService jwt.TokenService
	authRepo   repository.AuthRepository
)

func InitAuthMiddleware(service jwt.TokenService, repo repository.AuthRepository) {
	jwtService = service
	authRepo = repo
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		isBlacklisted, _ := authRepo.IsTokenBlacklisted(tokenString)
		if isBlacklisted {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token has been revoked",
			})
			return
		}

		claims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		ctx.Set("user", claims)
		ctx.Set("userID", claims.UserID)
		ctx.Next()
	}
}
