package service

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/dto"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/gin-gonic/gin"
)

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qrCode"`
}

type UserService interface {
	GetUserById(id string) (*domain.UserResponse, *utils.ReturnStatus)
	GetUserByEmail(email string) (*domain.UserResponse, *utils.ReturnStatus)
}

type AuthService interface {
	CreateUser(username, password, email string) (*domain.User, *utils.ReturnStatus)
	Login(email, password string) (user *domain.User, accessToken string, err *utils.ReturnStatus)
	SetupTOTP(userID string) (*TOTPSetupResponse, *utils.ReturnStatus)
	VerifyTOTP(userID string, code string) (bool, *utils.ReturnStatus)
	Logout(ctx *gin.Context) *utils.ReturnStatus
	LoginTOTP(email, totpCode string) (*domain.User, string, *utils.ReturnStatus)
}

type FileService interface {
	UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, req *dto.UploadRequest, ownerID *string) (*domain.File, *utils.ReturnStatus)
	GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) (interface{}, *utils.ReturnStatus)
	DeleteFile(ctx context.Context, fileID string, userID string) *utils.ReturnStatus
	GetFileInfo(ctx context.Context, token string, userID string, verbose bool) (*domain.File, *domain.User, []string, *utils.ReturnStatus)
	GetFileInfoID(ctx context.Context, token string, userID string, verbose bool) (*domain.File, *domain.User, []string, *utils.ReturnStatus)
	DownloadFile(ctx context.Context, token string, userID string, password string) (*domain.File, io.Reader, *utils.ReturnStatus)
	GetFileDownloadHistory(ctx context.Context, fileID string, userID string, pagenum, limit int) (*domain.FileDownloadHistory, *utils.ReturnStatus)
	GetFileStats(ctx context.Context, fileID string, userID string) (*domain.FileStat, *utils.ReturnStatus)
	GetAllAccessibleFiles(ctx context.Context, userID *string) ([]dto.AccessibleFile, *utils.ReturnStatus)
}

type AdminService interface {
	GetSystemPolicy(ctx context.Context) (*config.SystemPolicy, *utils.ReturnStatus)
	UpdateSystemPolicy(ctx context.Context, updates map[string]any) (*config.SystemPolicy, *utils.ReturnStatus)
	CleanupExpiredFiles(ctx context.Context) (int, *utils.ReturnStatus)
}
