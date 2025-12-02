package handlers

import (
	"net/http"
	"time"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/api/dto"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/service"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/validation"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	admin_service service.AdminService
}

func NewAdminHandler(admin_service service.AdminService) *AdminHandler {
	return &AdminHandler{
		admin_service: admin_service,
	}
}

func (ah *AdminHandler) GetSystemPolicy(ctx *gin.Context) {
	policy, err := ah.admin_service.GetSystemPolicy(ctx)

	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, policy)
}

func (ah *AdminHandler) UpdateSystemPolicy(ctx *gin.Context) {
	var req dto.UpdatePolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ResponseValidator(ctx, validation.HandleValidationErrors(err))
		return
	}

	updatedPolicy, err := ah.admin_service.UpdateSystemPolicy(ctx, req.ToMap())

	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "System policy updated successfully",
		"policy":  updatedPolicy,
	})
}

func (ah *AdminHandler) CleanupExpiredFiles(ctx *gin.Context) {
	deletedCount, err := ah.admin_service.CleanupExpiredFiles(ctx)
	if err != nil {
		err.Export(ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Cleanup completed",
		"deletedFiles": deletedCount,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}
