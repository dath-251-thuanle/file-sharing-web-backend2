package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"slices"

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
func (s *fileService) calculateValidityPeriod(req *dto.UploadRequest) (time.Time, time.Time, int, *utils.ReturnStatus) {
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
		return time.Time{}, time.Time{}, 0, utils.ResponseMsg(utils.ErrCodeBadRequest, "AvailableFrom cannot be after AvailableTo")
	}

	duration := availableTo.Sub(availableFrom)
	validityDays = int(duration.Hours() / 24)

	// b. Khoảng cách tối thiểu/tối đa
	minDuration := time.Duration(policy.MinValidityHours) * time.Hour
	maxDuration := time.Duration(policy.MaxValidityDays) * 24 * time.Hour

	if duration < minDuration {
		return time.Time{}, time.Time{}, 0, utils.ResponseMsg(utils.ErrCodeBadRequest, fmt.Sprintf("Validity period must be at least %d hours", policy.MinValidityHours))
	}
	if duration > maxDuration {
		return time.Time{}, time.Time{}, 0, utils.ResponseMsg(utils.ErrCodeBadRequest, fmt.Sprintf("Validity period cannot exceed %d days", policy.MaxValidityDays))
	}

	return availableFrom, availableTo, validityDays, nil
}

func (s *fileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, req *dto.UploadRequest, ownerID *string) (*domain.File, *utils.ReturnStatus) {
	// Kiểm tra kích thước file (Sử dụng MaxFileSizeMB từ Policy)
	if fileHeader.Size > int64(s.cfg.Policy.MaxFileSizeMB)*1024*1024 {
		return nil, utils.ResponseMsg(utils.ErrCodeBadRequest, fmt.Sprintf("File size exceeds the limit of %dMB", s.cfg.Policy.MaxFileSizeMB))
	}

	// 1. Tính toán thời gian hiệu lực
	availableFrom, availableTo, validityDays, err := s.calculateValidityPeriod(req)
	if err.IsErr() {
		return nil, err
	}

	// 2. Chuẩn bị File Metadata
	fileUUID := uuid.New().String()
	shareToken := utils.GenerateRandomString(16) // Hàm tạo token ngẫu nhiên 16 ký tự

	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, utils.ResponseMsg(utils.ErrCodeInternal, err.Error())
		}
		hashStr := string(hashed)
		passwordHash = &hashStr
	}

	storageFileName := fileUUID
	newFile := &domain.File{
		Id:            fileUUID,
		OwnerId:       ownerID,
		FileName:      fileHeader.Filename,
		StorageName:   storageFileName, // Tên file trên ổ đĩa sẽ là UUID
		FileSize:      fileHeader.Size,
		MimeType:      fileHeader.Header.Get("Content-Type"),
		ShareToken:    shareToken,
		IsPublic:      req.IsPublic || ownerID == nil, // buộc file là public khi không xác định được owner.
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
	if err.IsErr() {
		return nil, err
	}

	// 4. Lưu Metadata vào DB
	savedFile, err := s.fileRepo.CreateFile(ctx, newFile)
	if err.IsErr() {
		// QUAN TRỌNG: Nếu lưu DB lỗi, phải xóa file đã lưu vật lý!
		s.storage.DeleteFile(newFile.StorageName)
		return nil, err
	}

	// 5. Xử lý SharedWith
	if req.SharedWith != nil && *req.SharedWith != "" {
		var emails []string
		if err := json.Unmarshal([]byte(*req.SharedWith), &emails); err == nil {
			if err := s.sharedRepo.ShareFileWithUsers(ctx, savedFile.Id, emails); err != nil {
				return nil, err
			}
		}
	}

	return savedFile, nil
}

func (s *fileService) GetMyFiles(ctx context.Context, userID string, params domain.ListFileParams) (interface{}, *utils.ReturnStatus) {
	// Lấy danh sách file của user đó
	fileSummary, err := s.fileRepo.GetFileSummary(ctx, userID)
	if err.IsErr() {
		// Log lỗi hoặc xử lý lỗi một cách nhẹ nhàng hơn nếu summary không bắt buộc
		// Trong trường hợp này, ta sẽ trả về lỗi
		return nil, err
	}
	totalFiles, err := s.fileRepo.GetTotalUserFiles(ctx, userID)
	if err.IsErr() {
		return nil, err
	}
	files, err := s.fileRepo.GetMyFiles(ctx, userID, params)
	if err.IsErr() {
		return nil, err
	}
	totalPages := 0
	if params.Limit > 0 {
		totalPages = (totalFiles + params.Limit - 1) / params.Limit
	}

	pagination := gin.H{
		"currentPage": params.Page,
		"totalPages":  totalPages,
		"totalFiles":  totalFiles,
		"limit":       params.Limit,
	}

	out := []gin.H{}

	for _, f := range files {
		out = append(out, gin.H{
			"id":        f.Id,
			"fileName":  f.FileName,
			"status":    f.Status,
			"createdAt": f.CreatedAt,
		})
	}

	// 5. Trả về kết quả với dữ liệu thực tế
	return gin.H{
		"files":      out,
		"pagination": pagination,  // Dữ liệu phân trang thực tế
		"summary":    fileSummary, // Dữ liệu summary thực tế
	}, nil
}

func (s *fileService) DeleteFile(ctx context.Context, fileID string, userID string) *utils.ReturnStatus {
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err.IsErr() {
		return err
	}

	// Kiểm tra quyền: Chỉ Owner hoặc Admin mới được xóa
	isOwner := file.OwnerId != nil && *file.OwnerId == userID
	isAnonymous := file.OwnerId == nil

	if isAnonymous || !isOwner {
		// Cần thêm kiểm tra quyền Admin tại đây
		return utils.Response(utils.ErrCodeDeleteValidationErr)
	}

	// Xóa vật lý trước
	file.StorageName = fileID // Đảm bảo đúng tên file vật lý
	if err := s.storage.DeleteFile(file.StorageName); err.IsErr() {
		return err
	}

	if err := s.fileRepo.DeleteFile(ctx, fileID, userID); err.IsErr() {
		return err
	}

	return nil
}

func (s *fileService) getFileInfo(ctx context.Context, id string, userID string, isToken bool) (*domain.File, *domain.User, []string, *utils.ReturnStatus) {
	var file *domain.File = nil
	var err *utils.ReturnStatus = nil
	if isToken {
		file, err = s.fileRepo.GetFileByToken(ctx, id)
	} else {
		file, err = s.fileRepo.GetFileByID(ctx, id)
	}

	if err.IsErr() {
		return nil, nil, nil, err
	}

	now := time.Now()

	file.Status = domain.FILE_ACTIVE

	if now.Before(file.AvailableFrom) {
		file.Status = domain.FILE_PENDING
	} else if now.After(file.AvailableTo) {
		file.Status = domain.FILE_EXPIRED
	}

	owner_ := domain.User{}
	var owner *domain.User = nil
	if s.userRepo.FindById(*file.OwnerId, &owner_) == nil {
		owner = &owner_
	}

	shareds, err := s.sharedRepo.GetUsersSharedWith(ctx, file.Id)
	if err != nil {
		return nil, nil, nil, err
	}

	if !file.IsPublic && *file.OwnerId != userID {
		if !slices.Contains(shareds.UserIds, userID) {
			return nil, nil, nil, utils.Response(utils.ErrCodeGetForbidden)
		}

		if file.Status == domain.FILE_EXPIRED {
			return nil, nil, nil, utils.ResponseArgs(utils.ErrCodeFileExpired,
				gin.H{
					"expiredAt": file.AvailableTo,
				},
			)
		}

		if file.Status == domain.FILE_PENDING {
			return nil, nil, nil, utils.ResponseArgs(utils.ErrCodeFileLocked,
				gin.H{
					"availableFrom":       file.AvailableFrom,
					"hoursUntilAvailable": file.AvailableFrom.Sub(now).Hours(),
				},
			)
		}
	}

	outShared := []string{}
	for _, id := range shareds.UserIds {
		sharedowner := domain.User{}
		serr := s.userRepo.FindById(id, &sharedowner)
		if serr != nil {
			return nil, nil, nil, utils.ResponseMsg(utils.ErrCodeInternal, "Failed to retrieve emails of shared users")
		}

		outShared = append(outShared, sharedowner.Email)
	}

	return file, owner, outShared, nil
}

func (s *fileService) GetFileInfo(ctx context.Context, token string, userID string) (*domain.File, *domain.User, []string, *utils.ReturnStatus) {
	return s.getFileInfo(ctx, token, userID, true)
}

func (s *fileService) GetFileInfoID(ctx context.Context, id string, userID string) (*domain.File, *domain.User, []string, *utils.ReturnStatus) {
	return s.getFileInfo(ctx, id, userID, false)
}

func (s *fileService) DownloadFile(ctx context.Context, token string, userID string, password string) (*domain.File, []byte, *utils.ReturnStatus) {
	fileInfo, _, _, err := s.getFileInfo(ctx, token, userID, true)

	if err.IsErr() {
		return nil, nil, err
	}

	if fileInfo.HasPassword {
		if password == "" {
			return nil, nil, utils.Response(utils.ErrCodeDownloadPasswordInvalid)
		}

		if bcrypt.CompareHashAndPassword([]byte(*fileInfo.PasswordHash), []byte(password)) != nil {
			return nil, nil, utils.Response(utils.ErrCodeDownloadPasswordInvalid)
		}
	}

	fileReader, err := s.storage.GetFile(fileInfo.Id)
	if err.IsErr() {
		return nil, nil, err
	}

	file, readerr := io.ReadAll(fileReader)
	if readerr != nil {
		return nil, nil, utils.ResponseMsg(utils.ErrCodeInternal, readerr.Error())
	}

	if err := s.fileRepo.RegisterDownload(ctx, fileInfo.Id, userID); err.IsErr() {
		return nil, nil, err
	}

	return fileInfo, file, nil
}

func (s *fileService) GetFileDownloadHistory(ctx context.Context, fileID string, userID string, pagenum, limit int) (*domain.FileDownloadHistory, *utils.ReturnStatus) {
	history, err := s.fileRepo.GetFileDownloadHistory(ctx, fileID, userID)
	if err.IsErr() {
		return nil, err
	}

	history.Pagination = domain.Pagination{
		CurrentPage:  pagenum,
		TotalPages:   (len(history.History) + limit) / limit,
		TotalRecords: len(history.History),
		Limit:        limit,
	}

	start := (len(history.History) / limit) * pagenum
	end := min(start+limit, len(history.History))
	history.History = history.History[start:end]

	for i := range history.History {
		u := &history.History[i]

		if u.UserId == nil {
			continue
		}

		if *u.UserId == "" {
			continue
		}

		user := domain.User{}
		err := s.userRepo.FindById(*u.UserId, &user)
		if err != nil {
			return nil, err
		}

		u.Downloader = &domain.Downloader{
			Username: user.Username,
			Email:    user.Email,
		}
	}

	return history, nil
}

func (s *fileService) GetFileStats(ctx context.Context, fileID string, userID string) (*domain.FileStat, *utils.ReturnStatus) {
	return s.fileRepo.GetFileStats(ctx, fileID, userID)
}
