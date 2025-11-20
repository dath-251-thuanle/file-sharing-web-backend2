package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo     repository.UserRepository
	authRepo     repository.AuthRepository
	tokenService jwt.TokenService
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, tokenService jwt.TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
		authRepo:     authRepo,
	}
}

func (us *authService) CreateUser(username, password, email string) (*domain.User, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, utils.WrapError(err, "failed to hash password", utils.ErrCodeInternal)
	}
	hashedUserID, err := uuid.NewRandom()
	if err != nil {
		return nil, utils.WrapError(err, "failed to create UserID", utils.ErrCodeInternal)
	}
	//TODO: add username and email uniqueness check
	user := &domain.User{
		Id:         hashedUserID.String(),
		Username:   username,
		Password:   string(hashedPassword),
		Email:      email,
		Role:       "user",
		EnableTOTP: false,
		SecretTOTP: "",
	}
	return us.authRepo.Create(user)
}

func (as *authService) Login(email, password string) (*domain.User, string, int, error) {
	email = utils.NormalizeString(email)
	user := &domain.User{}
	err := as.userRepo.FindByEmail(email, user)
	if err != nil {
		fmt.Println("Login failed: User not found")
		return nil, "", 0, utils.NewError("Invalid email or password", utils.ErrCodeUnauthorized)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", 0, errors.New("invalid email or password")
	}

	accessToken, err := as.tokenService.GenerateAccessToken(*user)

	if err != nil {
		fmt.Println("Error generating access token:", err)
		return nil, "", 0, utils.NewError("Failed to generate access token", utils.ErrCodeInternal)
	}

	return user, accessToken, int(jwt.AccessTokenTTL.Seconds()), nil

}

func (as *authService) Logout(ctx *gin.Context) error {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.NewError("Missing Authorization header", utils.ErrCodeUnauthorized)
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := as.tokenService.ParseToken(accessToken)
	if err != nil {
		return utils.NewError("Invalid access token", utils.ErrCodeUnauthorized)
	}

	return as.authRepo.BlacklistToken(
		accessToken,
		claims.ExpiresAt.Time,
	)
}
