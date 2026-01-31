package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type LicenseHandler struct {
	manager *license.Manager
	logger  *zap.Logger
}

func NewLicenseHandler(manager *license.Manager, logger *zap.Logger) *LicenseHandler {
	return &LicenseHandler{manager: manager, logger: logger}
}

func (h *LicenseHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	lic := rg.Group("/license", authMw)
	{
		lic.GET("", h.GetLicense)
		lic.POST("", h.ActivateLicense)
		lic.DELETE("", h.DeactivateLicense)
	}
}

type activateLicenseInput struct {
	LicenseKey string `json:"license_key" binding:"required"`
}

func (h *LicenseHandler) GetLicense(c *gin.Context) {
	resp := h.manager.GetLicenseResponse()
	httputil.RespondSuccess(c, http.StatusOK, resp)
}

func (h *LicenseHandler) ActivateLicense(c *gin.Context) {
	var input activateLicenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("license_key", "license_key is required"))
		return
	}

	if err := h.manager.LoadLicense(input.LicenseKey); err != nil {
		h.logger.Warn("license activation failed", zap.Error(err))
		httputil.RespondError(c, httputil.Validation("license_key", "invalid or expired license key"))
		return
	}

	h.logger.Info("license activated",
		zap.String("tier", string(h.manager.GetTier())),
		zap.Bool("community", h.manager.IsCommunity()),
	)

	resp := h.manager.GetLicenseResponse()
	httputil.RespondSuccess(c, http.StatusOK, resp)
}

func (h *LicenseHandler) DeactivateLicense(c *gin.Context) {
	h.manager.RemoveLicense()
	h.logger.Info("license deactivated, reverted to community edition")

	resp := h.manager.GetLicenseResponse()
	httputil.RespondSuccess(c, http.StatusOK, resp)
}
