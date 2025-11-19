package service

import (
	"context"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type adminService struct {
	// policyRepo repository.PolicyRepository (Bỏ)
	cfg *config.Config // Lưu tham chiếu đến cấu hình
}

func NewAdminService(cfg *config.Config) AdminService {
	return &adminService{
		cfg: cfg,
	}
}

func (s *adminService) GetSystemPolicy(ctx context.Context) (*config.SystemPolicy, error) {
	return s.cfg.Policy, nil
}

func (s *adminService) UpdateSystemPolicy(ctx context.Context, updates map[string]any) (*config.SystemPolicy, error) {
	// Kiểm tra tính hợp lệ của DefaultValidityDays so với MaxValidityDays (cần convert sang int)
	if maxDaysVal, ok := updates["max_validity_days"]; ok {
		if defaultDaysVal, ok := updates["default_validity_days"]; ok {
			maxDays := maxDaysVal.(int)
			defaultDays := defaultDaysVal.(int)

			if defaultDays > maxDays {
				return nil, utils.NewError("Default validity days cannot be greater than max validity days", utils.ErrCodeBadRequest)
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

func (s *adminService) CleanupExpiredFiles(ctx context.Context) (int, error) {
	// Logic xóa file hết hạn...
	return 32, nil
}
