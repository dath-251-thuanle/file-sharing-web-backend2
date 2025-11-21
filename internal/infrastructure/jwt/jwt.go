package jwt

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var jwtSecretKey = []byte(utils.GetEnv("JWT_SECRET_KEY", "github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"))

const (
	AccessTokenTTL = time.Minute * 30
)

func NewJWTService() TokenService {
	return &JWTService{}
}

func (js *JWTService) GenerateAccessToken(user domain.User) (string, error) {
	// Implement token generation logic here
	claims := &Claims{
		UserID: user.Id,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "file-sharing",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

func (js *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if err != nil {
		return nil, utils.NewError("Invalid token", utils.ErrCodeUnauthorized)
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, utils.NewError("Invalid token", utils.ErrCodeUnauthorized)
	}
}
