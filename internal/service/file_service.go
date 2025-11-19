package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/dto"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/storage"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type FileService interface {
	UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, req *dto.UploadRequest, ownerID *string) (*domain.File, error)
	GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) (interface{}, error)
	DeleteFile(ctx context.Context, fileID string, userID string) error
	// GetFileInfo(ctx context.Context, token string) (*domain.File, error) // Cần cho download
}

type fileService struct {
	cfg        *config.Config
	fileRepo   repository.FileRepository
	sharedRepo repository.SharedRepository
	userRepo   repository.UserRepository // Cần để tìm User ID từ Email
	storage    storage.Storage
}

func NewFileService(cfg *config.Config, fr repository.FileRepository, sr repository.SharedRepository, ur repository.UserRepository, s storage.Storage) FileService {
	return &fileService{
		cfg:        cfg,
		fileRepo:   fr,
		sharedRepo: sr,
		userRepo:   ur,
		storage:    s,
	}
}

// Hàm tính toán thời gian hiệu lực
func (s *fileService) calculateValidityPeriod(req *dto.UploadRequest) (time.Time, time.Time, int, error) {
	now := time.Now().UTC()
	policy := s.cfg.Policy // Policy tĩnh

	var availableFrom, availableTo time.Time
	var validityDays int

	// 1. Tính toán availableFrom và availableTo
	if req.AvailableFrom != nil && req.AvailableTo != nil {
		availableFrom = *req.AvailableFrom
		availableTo = *req.AvailableTo
	} else if req.AvailableTo != nil {
		availableFrom = now
		availableTo = *req.AvailableTo
	} else if req.AvailableFrom != nil {
		availableFrom = *req.AvailableFrom
		availableTo = req.AvailableFrom.Add(time.Hour * 24 * time.Duration(policy.DefaultValidityDays))
	} else {
		availableFrom = now
		availableTo = now.Add(time.Hour * 24 * time.Duration(policy.DefaultValidityDays))
	}

	// 2. Validation
	// a. FROM < TO
	if availableFrom.After(availableTo) {
		return time.Time{}, time.Time{}, 0, utils.NewError("AvailableFrom cannot be after AvailableTo", utils.ErrCodeBadRequest)
	}

	duration := availableTo.Sub(availableFrom)
	validityDays = int(duration.Hours() / 24)

	// b. Khoảng cách tối thiểu/tối đa
	minDuration := time.Duration(policy.MinValidityHours) * time.Hour
	maxDuration := time.Duration(policy.MaxValidityDays) * 24 * time.Hour

	if duration < minDuration {
		return time.Time{}, time.Time{}, 0, utils.NewError(fmt.Sprintf("Validity period must be at least %d hours", policy.MinValidityHours), utils.ErrCodeBadRequest)
	}
	if duration > maxDuration {
		return time.Time{}, time.Time{}, 0, utils.NewError(fmt.Sprintf("Validity period cannot exceed %d days", policy.MaxValidityDays), utils.ErrCodeBadRequest)
	}

	return availableFrom, availableTo, validityDays, nil
}

func (s *fileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, req *dto.UploadRequest, ownerID *string) (*domain.File, error) {
	// Kiểm tra kích thước file (Sử dụng MaxFileSizeMB từ Policy)
	if fileHeader.Size > int64(s.cfg.Policy.MaxFileSizeMB)*1024*1024 {
		return nil, utils.NewError(fmt.Sprintf("File size exceeds the limit of %dMB", s.cfg.Policy.MaxFileSizeMB), utils.ErrCodeBadRequest)
	}

	// 1. Tính toán thời gian hiệu lực
	availableFrom, availableTo, validityDays, err := s.calculateValidityPeriod(req)
	if err != nil {
		return nil, err
	}

	// 2. Chuẩn bị File Metadata
	fileUUID := uuid.New().String()
	shareToken := utils.GenerateRandomString(16) // Hàm tạo token ngẫu nhiên 16 ký tự

	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, utils.WrapError(err, "Failed to hash password", utils.ErrCodeInternal)
		}
		hashStr := string(hashed)
		passwordHash = &hashStr
	}

	newFile := &domain.File{
		Id:            fileUUID,
		OwnerId:       ownerID,
		FileName:      fileHeader.Filename,
		StorageName:   fileUUID, // Tên file trên ổ đĩa sẽ là UUID
		FileSize:      fileHeader.Size,
		MimeType:      fileHeader.Header.Get("Content-Type"),
		ShareToken:    shareToken,
		IsPublic:      req.IsPublic,
		HasPassword:   passwordHash != nil,
		PasswordHash:  passwordHash,
		EnableTOTP:    req.EnableTOTP,
		AvailableFrom: availableFrom,
		AvailableTo:   availableTo,
		ValidityDays:  validityDays,
		CreatedAt:     time.Now().UTC(),
	}

	// 3. Lưu file vật lý
	_, err = s.storage.SaveFile(fileHeader, newFile.StorageName)
	if err != nil {
		return nil, utils.WrapError(err, "Failed to save file to storage", utils.ErrCodeInternal)
	}

	// 4. Lưu Metadata vào DB
	savedFile, err := s.fileRepo.CreateFile(ctx, newFile)
	if err != nil {
		// QUAN TRỌNG: Nếu lưu DB lỗi, phải xóa file đã lưu vật lý!
		s.storage.DeleteFile(newFile.StorageName)
		return nil, utils.WrapError(err, "Failed to save file metadata", utils.ErrCodeInternal)
	}

	// 5. Xử lý SharedWith
	if req.SharedWith != nil && *req.SharedWith != "" {
		var emails []string
		if err := json.Unmarshal([]byte(*req.SharedWith), &emails); err == nil {
			// Logic tìm User ID từ Email (cần UserRepository)
			userIDs := []string{"user-uuid-1", "user-uuid-2"} // Mô phỏng
			if len(userIDs) > 0 {
				s.sharedRepo.ShareFileWithUsers(ctx, savedFile.Id, userIDs) // Mô phỏng
			}
		}
	}

	return savedFile, nil
}

func (s *fileService) GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) (interface{}, error) {
	// Lấy danh sách file của user đó
	files, err := s.fileRepo.GetMyFiles(ctx, userID, params)
	if err != nil {
		return nil, utils.WrapError(err, "Failed to retrieve user files", utils.ErrCodeInternal)
	}

	// Logic tính toán Status, HoursRemaining và Summary (Mô phỏng)
	summary := domain.FileSummary{ActiveFiles: 28, PendingFiles: 5, ExpiredFiles: 9}

	// Cần thêm logic Pagination và tính toán status cho từng file

	return gin.H{
		"files":      files,
		"pagination": gin.H{"currentPage": params.Page, "totalPages": 3, "totalFiles": 42, "limit": params.Limit},
		"summary":    summary,
	}, nil
}

func (s *fileService) DeleteFile(ctx context.Context, fileID string, userID string) error {
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return utils.NewError("File not found", utils.ErrCodeNotFound)
	}

	// Kiểm tra quyền: Chỉ Owner hoặc Admin mới được xóa
	isOwner := file.OwnerId != nil && *file.OwnerId == userID
	isAnonymous := file.OwnerId == nil

	if isAnonymous || !isOwner {
		// Cần thêm kiểm tra quyền Admin tại đây
		return utils.NewError("Forbidden. Only the owner can delete the file", utils.ErrCodeUnauthorized)
	}

	// Xóa vật lý trước
	if err := s.storage.DeleteFile(file.StorageName); err != nil {
		return utils.WrapError(err, "Failed to delete file from storage", utils.ErrCodeInternal)
	}

	// Xóa metadata
	if err := s.fileRepo.DeleteFile(ctx, fileID, userID); err != nil {
		return utils.WrapError(err, "Failed to delete file metadata", utils.ErrCodeInternal)
	}

	return nil
}
