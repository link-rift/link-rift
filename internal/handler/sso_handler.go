package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/middleware"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type SSOHandler struct {
	ssoService service.SSOService
	logger     *zap.Logger
}

func NewSSOHandler(ssoService service.SSOService, logger *zap.Logger) *SSOHandler {
	return &SSOHandler{ssoService: ssoService, logger: logger}
}

// RegisterAdminRoutes registers workspace-scoped SSO admin endpoints.
func (h *SSOHandler) RegisterAdminRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	sso := wsScoped.Group("/sso")
	{
		sso.GET("", adminMw, h.GetSSOConfig)
		sso.POST("", adminMw, h.CreateSSOConfig)
		sso.PUT("", adminMw, h.UpdateSSOConfig)
		sso.DELETE("", adminMw, h.DeleteSSOConfig)
		sso.GET("/identities", adminMw, h.ListIdentities)
	}
}

// RegisterPublicRoutes registers public SSO flow endpoints (no auth required).
func (h *SSOHandler) RegisterPublicRoutes(v1 *gin.RouterGroup) {
	sso := v1.Group("/auth/sso")
	{
		sso.GET("/metadata/:workspaceId", h.GetSPMetadata)
		sso.GET("/login/:workspaceId", h.InitiateSSO)
		sso.POST("/callback/:workspaceId", h.HandleCallback)
	}
}

// Admin endpoints

func (h *SSOHandler) GetSSOConfig(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	cfg, err := h.ssoService.GetConfig(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, cfg)
}

func (h *SSOHandler) CreateSSOConfig(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.CreateSSOConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	cfg, err := h.ssoService.CreateConfig(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, cfg)
}

func (h *SSOHandler) UpdateSSOConfig(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.UpdateSSOConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	cfg, err := h.ssoService.UpdateConfig(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, cfg)
}

func (h *SSOHandler) DeleteSSOConfig(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	if err := h.ssoService.DeleteConfig(c.Request.Context(), ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "SSO configuration deleted"})
}

func (h *SSOHandler) ListIdentities(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	identities, err := h.ssoService.ListIdentities(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, identities)
}

// Public SSO flow endpoints

func (h *SSOHandler) GetSPMetadata(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("workspaceId", "invalid workspace ID"))
		return
	}

	metadata, err := h.ssoService.GetSPMetadata(c.Request.Context(), workspaceID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	c.Data(http.StatusOK, "application/xml", []byte(metadata))
}

func (h *SSOHandler) InitiateSSO(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("workspaceId", "invalid workspace ID"))
		return
	}

	authURL, err := h.ssoService.InitiateSSO(c.Request.Context(), workspaceID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func (h *SSOHandler) HandleCallback(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("workspaceId", "invalid workspace ID"))
		return
	}

	samlResponse := c.PostForm("SAMLResponse")
	if samlResponse == "" {
		httputil.RespondError(c, httputil.Validation("SAMLResponse", "missing SAML response"))
		return
	}

	identity, err := h.ssoService.HandleCallback(c.Request.Context(), workspaceID, samlResponse)
	if err != nil {
		// Redirect to frontend with error
		h.logger.Error("SSO callback failed", zap.Error(err))
		c.Redirect(http.StatusFound, "/auth/login?sso_error=authentication_failed")
		return
	}

	// In a full implementation, we would:
	// 1. Create a session and tokens for the user
	// 2. Redirect to frontend with the tokens
	// For now, return the identity data
	httputil.RespondSuccess(c, http.StatusOK, identity)
}
