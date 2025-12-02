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
	}
	protected := files.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/my", fr.handler.GetMyFiles)

		protected.DELETE("/:id", fr.handler.DeleteFile)
		protected.GET("/:ident", fr.handler.GetFileInfo)

		protected.GET("/:ident/stats", fr.handler.GetFileStats)
		protected.GET("/:ident/download", fr.handler.DownloadFile)
		protected.GET("/:ident/download-history", fr.handler.GetFileDownloadHistory)
	}
}
