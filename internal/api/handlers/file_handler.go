package handlers

import (
	"fmt"
	"io"
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
		utils.Response(utils.ErrCodeFileUploadRequired).Export(ctx)
		return
	}

	req.FileNameForValidation = fileHeader.Filename

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	if req.Password != nil {
		if len(*req.Password) < 8 {
			utils.ResponseMsg(utils.ErrCodeBadRequest, "Password must be at least 8 characters long").Export(ctx)
			return
		}
	}

	var userID *string
	if val, exists := ctx.Get("userID"); exists && val != "" {
		strVal := val.(string)
		userID = &strVal
	} else {
		userID = nil
	}

	if userID == nil && (!req.IsPublic || req.SharedWith != nil) {
		utils.Response(utils.ErrCodeFilePrivateNeedsAuth).Export(ctx)
		return
	}

	if req.IsPublic && req.SharedWith != nil {
		utils.Response(utils.ErrCodeFileUploadPublicWithShared).Export(ctx)
		return
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
		"isPublic":   uploadedFile.IsPublic,
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
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
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
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
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
	ident := ctx.Param("shareToken")
	userID, exists := ctx.Get("userID")
	if !exists {
		userID = ""
	}

	var file *domain.File = nil
	var err *utils.ReturnStatus = nil

	if uuid.Validate(ident) == nil {
		file, _, _, err = fh.file_service.GetFileInfoID(ctx, ident, userID.(string), false)
	} else {
		file, _, _, err = fh.file_service.GetFileInfo(ctx, ident, userID.(string), false)
	}

	if err != nil {
		err.Export(ctx)
		return
	}

	out := gin.H{
		"id":          file.Id,
		"fileName":    file.FileName,
		"shareToken":  file.ShareToken,
		"status":      file.Status,
		"isPublic":    file.IsPublic,
		"hasPassword": file.HasPassword,
		"fileSize":    file.FileSize,
		"mimeType":    file.MimeType,
	}

	//utils.ResponseSuccess(ctx, http.StatusOK, "File retrieved successfully", gin.H{"file": result})
	ctx.JSON(http.StatusOK, gin.H{
		"file": out,
	})
}

func (fh *FileHandler) GetFileInfoVerbose(ctx *gin.Context) {
	ident := ctx.Param("id")
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
		return
	}

	var file *domain.File = nil
	var owner *domain.User = nil
	var err *utils.ReturnStatus = nil
	shared := []string{}

	if uuid.Validate(ident) == nil {
		file, owner, shared, err = fh.file_service.GetFileInfoID(ctx, ident, userID.(string), true)
	} else {
		file, owner, shared, err = fh.file_service.GetFileInfo(ctx, ident, userID.(string), true)
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
		"shareLink":   fmt.Sprintf("http://localhost:8080/files/%s", file.ShareToken),
		"isPublic":    file.IsPublic,
		"hasPassword": file.HasPassword,

		"availableFrom": file.AvailableFrom,
		"availableTo":   file.AvailableTo,
		"status":        file.Status,

		"hoursRemaining": file.AvailableTo.Sub(file.AvailableFrom).Hours(),

		"createdAt": file.CreatedAt,
	}

	out["owner"] = gin.H{
		"id":       owner.Id,
		"username": owner.Username,
		"email":    owner.Email,
		"role":     owner.Role,
	}

	if shared != nil {
		out["sharedWith"] = shared
	}

	//utils.ResponseSuccess(ctx, http.StatusOK, "File retrieved successfully", gin.H{"file": result})
	ctx.JSON(http.StatusOK, gin.H{
		"file": out,
	})
}

func (fh *FileHandler) getFileData(ctx *gin.Context) (*domain.File, []byte, *utils.ReturnStatus) {
	fileToken := ctx.Param("shareToken")
	password := ctx.Query("password")
	userIDptr, exists := ctx.Get("userID")
	var userID string = ""
	if exists {
		userID = userIDptr.(string)
	}

	info, file, download_err := fh.file_service.DownloadFile(ctx, fileToken, userID, password)
	if download_err != nil {
		return nil, nil, download_err
	}

	fileBytes, readerr := io.ReadAll(file)
	if readerr != nil {
		return nil, nil, utils.ResponseMsg(utils.ErrCodeInternal, readerr.Error())
	}

	return info, fileBytes, nil
}

func (fh *FileHandler) DownloadFile(ctx *gin.Context) {
	info, file, err := fh.getFileData(ctx)
	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.Data(http.StatusOK, info.MimeType, file)
}

func (fh *FileHandler) PreviewFile(ctx *gin.Context) {
	info, file, err := fh.getFileData(ctx)
	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.Header("Content-Disposition", "inline; filename=\""+info.FileName+"\"")
	ctx.Data(http.StatusOK, info.MimeType, file)
}

func (fh *FileHandler) GetFileDownloadHistory(ctx *gin.Context) {
	fileID := ctx.Param("id")
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
		return
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
	fileID := ctx.Param("id")
	userID, exists := ctx.Get("userID")

	if !exists {
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
		return
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

func (fh *FileHandler) GetAccessibleFiles(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")

	if !exists {
		utils.Response(utils.ErrCodeBearerInvalid).Export(ctx)
		return
	}

	files, err := fh.file_service.GetAccessibleFiles(ctx, userID.(string))
	if err != nil {
		err.Export(ctx)
		return
	}

	page := utils.GetIntQuery(ctx, "page", 1)
	limit := utils.GetIntQuery(ctx, "limit", 20)
	totalPages := 1

	if len(files) != 0 {
		totalPages = (len(files) + limit) / limit
	}

	if page > totalPages {
		page = totalPages
	}

	start := limit * (page - 1)
	end := min(start+limit, len(files))

	ctx.JSON(http.StatusOK, gin.H{
		"files": files[start:end],
		"pagination": gin.H{
			"currentPage": page,
			"totalPages":  totalPages,
			"totalFiles":  len(files),
			"limit":       limit,
		},
	})
}
