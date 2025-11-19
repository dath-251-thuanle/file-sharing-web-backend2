package dto

import "time"

// UploadRequest là DTO cho POST /api/files/upload
// Sử dụng tag 'form' vì đây là multipart/form-data
type UploadRequest struct {
	// File cần upload. GIN sẽ tự động bind *multipart.FileHeader

	IsPublic bool `form:"isPublic"` // Mặc định false

	// Sử dụng string để validate file extension
	FileNameForValidation string `form:"file_validation_placeholder" validate:"file_ext=pdf jpg png txt"`

	Password *string `form:"password" validate:"omitempty,min=6"`

	// ISO Date: YYYY-MM-DDTHH:MM:SSZ
	AvailableFrom *time.Time `form:"availableFrom" time_format:"2006-01-02T15:04:05Z"`
	AvailableTo   *time.Time `form:"availableTo" time_format:"2006-01-02T15:04:05Z"`

	// Dữ liệu JSON array được gửi dưới dạng string trong form-data
	SharedWith *string `form:"sharedWith"`

	EnableTOTP bool `form:"enableTOTP"`
}

// FileHeader là một wrapper để GIN có thể bind field "file"
// Tuy nhiên, thường thì bạn sử dụng *multipart.FileHeader trực tiếp trong Handler/Service.
// Giữ nguyên FileHeader cho mục đích cấu trúc.
