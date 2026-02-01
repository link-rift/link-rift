package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/middleware"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type BrandingHandler struct {
	brandingService service.BrandingService
	logger          *zap.Logger
}

func NewBrandingHandler(brandingService service.BrandingService, logger *zap.Logger) *BrandingHandler {
	return &BrandingHandler{brandingService: brandingService, logger: logger}
}

func (h *BrandingHandler) RegisterRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	branding := wsScoped.Group("/branding")
	{
		branding.GET("", adminMw, h.GetBranding)
		branding.PUT("", adminMw, h.UpdateBranding)
		branding.DELETE("", adminMw, h.DeleteBranding)
	}
}

func (h *BrandingHandler) GetBranding(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	branding, err := h.brandingService.GetBranding(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, branding)
}

func (h *BrandingHandler) UpdateBranding(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.UpdateBrandingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	branding, err := h.brandingService.UpdateBranding(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, branding)
}

func (h *BrandingHandler) DeleteBranding(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	if err := h.brandingService.DeleteBranding(c.Request.Context(), ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "branding deleted successfully"})
}
