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
	"github.com/google/uuid"
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

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.ResponseMsg(utils.ErrCodeUploadBadRequest, "File is required").Export(ctx)
		return
	}

	req.FileNameForValidation = fileHeader.Filename

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	var userID *string
	if val, exists := ctx.Get("userID"); exists && val != "" {
		strVal := val.(string)
		userID = &strVal
	} else {
		userID = nil
	}

	uploadedFile, berr := fh.file_service.UploadFile(ctx, fileHeader, &req, userID)
	if berr != nil {
		berr.Export(ctx)
		return
	}

	response := gin.H{
		"id":         uploadedFile.Id,
		"fileName":   uploadedFile.FileName,
		"shareToken": uploadedFile.ShareToken,
	}

	//utils.ResponseSuccess(ctx, http.StatusCreated, "File uploaded successfully", gin.H{"file": response})
	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"file":    response,
		"message": "File uploaded successfully",
	})
}

func (fh *FileHandler) DeleteFile(ctx *gin.Context) {
	fileID := ctx.Param("id")

	if uuid.Validate(fileID) != nil {
		utils.ResponseMsg(utils.ErrCodeBadRequest, "Invalid ID provided").Export(ctx)
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Response(utils.ErrCodeUnauthorized).Export(ctx)
		return
	}

	err := fh.file_service.DeleteFile(ctx, fileID, userID.(string))

	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
		"fileId":  fileID,
	})
}

func (fh *FileHandler) GetMyFiles(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.ResponseError(ctx, utils.NewError("Unauthorized access", utils.ErrCodeUnauthorized))
		return
	}

	status := ctx.DefaultQuery("status", "all")
	page := utils.GetIntQuery(ctx, "page", 1)
	limit := utils.GetIntQuery(ctx, "limit", 20)
	sortBy := ctx.DefaultQuery("sortBy", "createdAt")
	order := ctx.DefaultQuery("order", "desc")

	params := domain.ListFileParams{
		Status: strings.ToLower(status),
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  strings.ToLower(order),
	}

	result, err := fh.file_service.GetMyFiles(ctx, userID.(string), params)

	if err != nil {
		err.Export(ctx)
		return
	}

	//utils.ResponseSuccess(ctx, http.StatusOK, "User files retrieved successfully", gin.H{"file": result})
	ctx.JSON(http.StatusOK, result)
}

func (fh *FileHandler) GetFileInfo(ctx *gin.Context) {
	ident := ctx.Param("ident")
	userID, exists := ctx.Get("userID")
	if !exists {
		userID = nil
	}

	var file *domain.File = nil
	var owner *domain.User = nil
	var err *utils.ReturnStatus = nil
	shared := []string{}

	if uuid.Validate(ident) == nil {
		file, owner, shared, err = fh.file_service.GetFileInfoID(ctx, ident, userID.(string))
	} else {
		file, owner, shared, err = fh.file_service.GetFileInfo(ctx, ident, userID.(string))
	}

	if err != nil {
		err.Export(ctx)
		return
	}

	out := gin.H{
		"id":          file.Id,
		"fileName":    file.FileName,
		"fileSize":    file.FileSize,
		"mimeType":    file.MimeType,
		"shareToken":  file.ShareToken,
		"shareLink":   fmt.Sprintf("http://localhost:8080/api/files/%s", file.ShareToken),
		"isPublic":    file.IsPublic,
		"hasPassword": file.HasPassword,

		"availableFrom": file.AvailableFrom,
		"availableTo":   file.AvailableTo,
		"status":        file.Status,

		"hoursRemaining": file.AvailableTo.Sub(file.AvailableFrom).Hours(),

		"createdAt": file.CreatedAt,
	}

	if owner != nil {
		out["owner"] = gin.H{
			"id":       owner.Id,
			"username": owner.Username,
			"email":    owner.Email,
			"role":     owner.Role,
		}
	}

	if shared != nil {
		out["sharedWith"] = shared
	}

	//utils.ResponseSuccess(ctx, http.StatusOK, "File retrieved successfully", gin.H{"file": result})
	ctx.JSON(http.StatusOK, gin.H{
		"file": out,
	})
}

func (fh *FileHandler) DownloadFile(ctx *gin.Context) {
	fileToken := ctx.Param("ident")
	password := ctx.Query("password")
	userID, exists := ctx.Get("userID")
	if !exists {
		userID = nil
	}

	info, file, download_err := fh.file_service.DownloadFile(ctx, fileToken, userID.(string), password)
	if download_err != nil {
		download_err.Export(ctx)
		return
	}

	ctx.Data(http.StatusOK, info.MimeType, file)
}

func (fh *FileHandler) GetFileDownloadHistory(ctx *gin.Context) {
	fileID := ctx.Param("ident")
	userID, exists := ctx.Get("userID")
	if !exists {
		userID = nil
	}

	page := utils.GetIntQuery(ctx, "page", 1)
	limit := utils.GetIntQuery(ctx, "limit", 20)
	if limit == 0 {
		utils.ResponseError(ctx, utils.NewError("Limit must not be 0", utils.ErrCodeBadRequest))
		return
	}

	history, download_err := fh.file_service.GetFileDownloadHistory(ctx, fileID, userID.(string), page, limit)
	if download_err != nil {
		download_err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, history)
}

func (fh *FileHandler) GetFileStats(ctx *gin.Context) {
	fileID := ctx.Param("ident")
	userID, exists := ctx.Get("userID")

	if !exists {
		userID = nil
	}

	if uuid.Validate(fileID) != nil {
		utils.Response(utils.ErrCodeFileNotFound).Export(ctx)
		return
	}

	stats, err := fh.file_service.GetFileStats(ctx, fileID, userID.(string))
	if err != nil {
		err.Export(ctx)
		return
	}

	out := gin.H{
		"fileId":   stats.FileId,
		"fileName": stats.FileName,
		"statistics": gin.H{
			"downloadCount":     stats.TotalDownloadCount,
			"uniqueDownloaders": stats.UserDownloadCount,
			"lastDownloadedAt":  stats.LastDownloadedAt,
			"createdAt":         stats.CreatedAt,
		},
	}

	ctx.JSON(http.StatusOK, out)
}
