package service

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/jwt"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
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
		authRepo:     authRepo,
		tokenService: tokenService,
	}
}

func (us *authService) CreateUser(username, password, email string) (*domain.User, *utils.ReturnStatus) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeInternal, "failed to hash password")
	}
	hashedUserID, err := uuid.NewRandom()
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeInternal, "failed to create UserID")
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

func (as *authService) Login(email, password string) (*domain.User, string, *utils.ReturnStatus) {
	email = utils.NormalizeString(email)
	user := &domain.User{}
	err := as.userRepo.FindByEmail(email, user)
	if err != nil {
		fmt.Println("Login failed: User not found")
		return nil, "", utils.Response(utils.ErrCodeLoginInvalid)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, "", utils.Response(utils.ErrCodeLoginInvalid)
	}

	if user.EnableTOTP {
		cid, err := uuid.NewUUID()
		if err != nil {
			return nil, "", utils.ResponseMsg(utils.ErrCodeInternal, "Failed to generate CID")
		}
		timstamp_err := as.userRepo.AddTimestamp(user.Id, cid.String())
		if timstamp_err != nil {
			return nil, "", timstamp_err
		}
		return user, cid.String(), nil
	}

	accessToken, gen_err := as.tokenService.GenerateAccessToken(*user)

	if gen_err != nil {
		fmt.Println("*utils.ReturnStatus generating access token:", err)
		return nil, "", utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("Failed to generate access token: %s", gen_err.Error()))
	}

	return user, accessToken, nil

}

func (as *authService) LoginTOTP(id, totpCode string) (*domain.User, string, *utils.ReturnStatus) {
	user := &domain.User{}
	err := as.userRepo.FindById(id, user)
	if err != nil {
		fmt.Println("Login failed: User not found")
		return nil, "", utils.ResponseMsg(utils.ErrCodeUnauthorized, "Invalid ID")
	}
	secret := user.SecretTOTP
	if !totp.Validate(totpCode, secret) {
		return nil, "", utils.ResponseMsg(utils.ErrCodeUnauthorized, "Invalid or expired TOTP code")
	}

	accessToken, gen_err := as.tokenService.GenerateAccessToken(*user)

	if gen_err != nil {
		fmt.Println("*utils.ReturnStatus generating access token:", gen_err)
		return nil, "", utils.ResponseMsg(utils.ErrCodeUnauthorized, fmt.Sprintf("Failed to generate access token: %s", gen_err.Error()))
	}

	return user, accessToken, nil

}

func (as *authService) Logout(ctx *gin.Context) *utils.ReturnStatus {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.ResponseMsg(utils.ErrCodeUnauthorized, "Missing Authorization header")
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := as.tokenService.ParseToken(accessToken)
	if err != nil {
		return utils.ResponseMsg(utils.ErrCodeUnauthorized, "Invalid access token")
	}

	return as.authRepo.BlacklistToken(
		accessToken,
		claims.ExpiresAt.Time,
	)
}

func (as *authService) SetupTOTP(userID string) (*TOTPSetupResponse, *utils.ReturnStatus) {
	const appName = "file-sharing"
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: fmt.Sprintf("user-%s", userID),
	})
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeInternal, err.Error())
	}

	secret := key.Secret()
	otpURL := key.URL()

	if err := as.authRepo.SaveSecret(userID, secret); err != nil {
		return nil, err
	}

	png, err := qrcode.Encode(otpURL, qrcode.Medium, 256)
	if err != nil {
		return nil, utils.ResponseMsg(utils.ErrCodeInternal, err.Error())
	}
	qrBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	return &TOTPSetupResponse{
		Secret: secret,
		QRCode: qrBase64,
	}, nil
}

func (as *authService) VerifyTOTP(userID string, code string) (bool, *utils.ReturnStatus) {
	secret, err := as.authRepo.GetSecret(userID)
	if err != nil {
		return false, err
	}

	valid := totp.Validate(code, secret)

	if valid {
		if err := as.authRepo.EnableTOTP(userID); err != nil {
			return true, utils.ResponseMsg(utils.ErrCodeInternal, fmt.Sprintf("verified but failed to enable status: %v", err))
		}
	}

	return valid, nil
}
