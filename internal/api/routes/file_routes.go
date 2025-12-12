package routes

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/handlers"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/middleware"
	"github.com/gin-gonic/gin"
)

type FileRoutes struct {
	handler *handlers.FileHandler
}

func NewFileRoutes(handler *handlers.FileHandler) *FileRoutes {
	return &FileRoutes{
		handler: handler,
	}
}

func (fr *FileRoutes) Register(r *gin.RouterGroup) {
	files := r.Group("/files")
	optional := files.Group("/")
	optional.Use(middleware.AuthMiddlewareUpload())
	{
		optional.POST("/upload", fr.handler.UploadFile)

		optional.GET("/:shareToken", fr.handler.GetFileInfo)

		optional.GET("/:shareToken/download", fr.handler.DownloadFile)

		optional.GET("/available", fr.handler.GetAllAccessibleFiles)
	}
	protected := files.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/my", fr.handler.GetMyFiles)

		optional.GET("/:shareToken/preview", fr.handler.PreviewFile)

		// Sử dụng ID.
		protected.DELETE("/info/:id", fr.handler.DeleteFile)
		protected.GET("info/:id", fr.handler.GetFileInfoVerbose)
		protected.GET("/stats/:id", fr.handler.GetFileStats)
		protected.GET("/download-history/:id", fr.handler.GetFileDownloadHistory)
	}
}
