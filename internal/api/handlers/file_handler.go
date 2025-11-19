package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/dto"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/domain"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/validation"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	file_service service.FileService
}

func NewFileHandler(file_service service.FileService) *FileHandler {
	return &FileHandler{
		file_service: file_service,
	}
}

func (fh *FileHandler) UploadFile(ctx *gin.Context) {
	var req dto.UploadRequest

	// 1. Lấy file từ request form
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		// Trả về lỗi nếu không tìm thấy file
		utils.ResponseError(ctx, utils.NewError("File is required for upload", utils.ErrCodeBadRequest))
		return
	}

	// 2. Set FileNameForValidation để Validation có thể kiểm tra extension
	req.FileNameForValidation = fileHeader.Filename

	// 3. Bind các trường dữ liệu khác
	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	// 4. Lấy UserID từ context (Cho phép Anonymous Upload)
	var userID *string
	if val, exists := ctx.Get("userID"); exists && val != "" {
		strVal := val.(string)
		userID = &strVal
	} else {
		// Nếu không tìm thấy userID, đây là ANONYMOUS UPLOAD.
		// Set userID = nil và TIẾP TỤC xử lý.
		utils.ResponseError(ctx, utils.NewError("Authentication required for upload", utils.ErrCodeUnauthorized))
		return // Dừng request

	}

	// 5. Xử lý upload
	uploadedFile, err := fh.file_service.UploadFile(ctx, fileHeader, &req, userID)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	// 6. Chuẩn bị Response DTO
	response := gin.H{
		"id":            uploadedFile.Id,
		"fileName":      uploadedFile.FileName,
		"fileSize":      uploadedFile.FileSize,
		"shareToken":    uploadedFile.ShareToken,
		"shareLink":     fmt.Sprintf("http://%s/f/%s", ctx.Request.Host, uploadedFile.ShareToken),
		"isPublic":      uploadedFile.IsPublic,
		"hasPassword":   uploadedFile.HasPassword,
		"availableFrom": uploadedFile.AvailableFrom,
		"availableTo":   uploadedFile.AvailableTo,
		"validityDays":  uploadedFile.ValidityDays,
		"enableTOTP":    uploadedFile.EnableTOTP,
		"createdAt":     uploadedFile.CreatedAt,
	}

	// Trả về 201 Created
	utils.ResponseSuccess(ctx, http.StatusCreated, "File uploaded successfully", gin.H{"file": response})
}

func (fh *FileHandler) DeleteFile(ctx *gin.Context) {
	fileID := ctx.Param("id")

	// Lấy UserID (bắt buộc phải login để xóa file)
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.ResponseError(ctx, utils.NewError("Unauthorized access", utils.ErrCodeUnauthorized))
		return
	}

	err := fh.file_service.DeleteFile(ctx, fileID, userID.(string))

	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "File deleted successfully", gin.H{"fileId": fileID})
}

func (fh *FileHandler) GetMyFiles(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.ResponseError(ctx, utils.NewError("Unauthorized access", utils.ErrCodeUnauthorized))
		return
	}

	// Lấy tham số query
	status := ctx.DefaultQuery("status", "all")
	page := utils.GetIntQuery(ctx, "page", 1)
	limit := utils.GetIntQuery(ctx, "limit", 20)
	sortBy := ctx.DefaultQuery("sortBy", "createdAt")
	order := ctx.DefaultQuery("order", "desc")

	// Ánh xạ tham số vào domain.ListFileParams
	params := domain.ListFileParams{
		Status: strings.ToLower(status),
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  strings.ToLower(order),
	}

	result, err := fh.file_service.GetMyFiles(ctx, userID.(string), params)

	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "User files retrieved successfully", result)
}
