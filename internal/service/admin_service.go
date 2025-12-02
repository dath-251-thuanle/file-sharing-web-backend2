package service

import (
	"context"
	"log"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/infrastructure/storage"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/repository"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type adminService struct {
	cfg      *config.Config            // Lưu tham chiếu đến cấu hình
	fileRepo repository.FileRepository // <-- THÊM: Để truy vấn file
	storage  storage.Storage           // <-- THÊM: Để xóa file vật lý
}

func NewAdminService(cfg *config.Config, fr repository.FileRepository, s storage.Storage) AdminService {
	return &adminService{
		cfg:      cfg,
		fileRepo: fr,
		storage:  s,
	}
}

func (s *adminService) GetSystemPolicy(ctx context.Context) (*config.SystemPolicy, *utils.ReturnStatus) {
	return s.cfg.Policy, nil
}

func (s *adminService) UpdateSystemPolicy(ctx context.Context, updates map[string]any) (*config.SystemPolicy, *utils.ReturnStatus) {
	// Kiểm tra tính hợp lệ của DefaultValidityDays so với MaxValidityDays (cần convert sang int)
	if maxDaysVal, ok := updates["max_validity_days"]; ok {
		if defaultDaysVal, ok := updates["default_validity_days"]; ok {
			maxDays := maxDaysVal.(int)
			defaultDays := defaultDaysVal.(int)

			if defaultDays > maxDays {
				return nil, utils.ResponseMsg(utils.ErrCodeBadRequest, "Default validity days cannot be greater than max validity days")
			}
		}
	}

	if maxFileSizeVal, exists := updates[utils.CamelToSnake("MaxFileSizeMB")]; exists {
		s.cfg.Policy.MaxFileSizeMB = maxFileSizeVal.(int)
	}
	if minValidityHoursVal, exists := updates[utils.CamelToSnake("MinValidityHours")]; exists {
		s.cfg.Policy.MinValidityHours = minValidityHoursVal.(int)
	}
	if maxValidityDaysVal, exists := updates[utils.CamelToSnake("MaxValidityDays")]; exists {
		s.cfg.Policy.MaxValidityDays = maxValidityDaysVal.(int)
	}
	if defaultValidityDaysVal, exists := updates[utils.CamelToSnake("DefaultValidityDays")]; exists {
		s.cfg.Policy.DefaultValidityDays = defaultValidityDaysVal.(int)
	}
	if requirePasswordMinLengthVal, exists := updates[utils.CamelToSnake("RequirePasswordMinLength")]; exists {
		s.cfg.Policy.RequirePasswordMinLength = requirePasswordMinLengthVal.(int)
	}
	return s.cfg.Policy, nil
}

func (s *adminService) CleanupExpiredFiles(ctx context.Context) (int, *utils.ReturnStatus) {
	// Giả định FileRepository có hàm FindAll để lấy TẤT CẢ files
	files, err := s.fileRepo.FindAll(ctx)
	if err.IsErr() {
		return 0, err
	}

	now := time.Now().UTC()
	deletedCount := 0

	// Duyệt qua tất cả các file
	for _, file := range files {
		// 1. Kiểm tra ngày hết hạn
		if file.AvailableTo.Before(now) {

			if err := s.storage.DeleteFile(file.FileName); err.IsErr() {
				// Log lỗi nhưng tiếp tục sang file tiếp theo
				log.Printf("Cleanup Error: Failed to delete physical file %s: %v", file.FileName, err)
				continue
			}
			var ownerID string
			if file.OwnerId != nil {
				ownerID = *file.OwnerId
			} else {

				log.Printf("Cleanup Warning: Skipping metadata delete for Anonymous file %s. Requires specific repo method.", file.Id)
				continue
			}

			// Xóa file (chỉ áp dụng cho file có Owner)
			if err := s.fileRepo.DeleteFile(ctx, file.Id, ownerID); err.IsErr() {
				// Log lỗi nhưng tiếp tục
				log.Printf("Cleanup Error: Failed to delete metadata for file %s: %v", file.Id, err)
				continue
			}

			deletedCount++
		}
	}

	return deletedCount, nil
}
