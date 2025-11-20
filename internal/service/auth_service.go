package service

import (
	"errors"
	"fmt"
	"strings"
	"encoding/base64"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

type authService struct {
	userRepo     repository.UserRepository
	authRepo     repository.AuthRepository
	tokenService jwt.TokenService
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, tokenService jwt.TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		authRepo:     authRepo,
		tokenService: tokenService,
	}
}

func (us *authService) CreateUser(username, password, email, role string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, utils.WrapError(err, "failed to hash password", utils.ErrCodeInternal)
	}
	user := &domain.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Role:     role,
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

func (as *authService) SetupTOTP(userID string) (*TOTPSetupResponse, error) {
	const appName = "FileSharingApp"
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: fmt.Sprintf("user-%s", userID),
	})
	if err != nil {
		return nil, err
	}

	secret := key.Secret()
	otpURL := key.URL()

	if err := as.authRepo.SaveSecret(userID, secret); err != nil {
		return nil, err
	}

	png, err := qrcode.Encode(otpURL, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}
	qrBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	return &TOTPSetupResponse{
		Secret: secret,
		QRCode: qrBase64,
	}, nil
}

func (as *authService) VerifyTOTP(userID string, code string) (bool, error) {
	secret, err := as.authRepo.GetSecret(userID)
	if err != nil {
		return false, err
	}

	valid := totp.Validate(code, secret)

	if valid {
		if err := as.authRepo.EnableTOTP(userID); err != nil {
			return true, fmt.Errorf("verified but failed to enable status: %v", err)
		}
	}

	return valid, nil
}