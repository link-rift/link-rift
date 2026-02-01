package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/ee/scim"
	"github.com/link-rift/link-rift/internal/middleware"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// SCIMHandler handles admin SCIM token management and SCIM 2.0 protocol endpoints.
type SCIMHandler struct {
	scimService service.SCIMService
	logger      *zap.Logger
}

// NewSCIMHandler creates a new SCIM handler.
func NewSCIMHandler(scimService service.SCIMService, logger *zap.Logger) *SCIMHandler {
	return &SCIMHandler{scimService: scimService, logger: logger}
}

// RegisterAdminRoutes registers admin SCIM token management routes (workspace-scoped, auth required).
func (h *SCIMHandler) RegisterAdminRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	scimAdmin := wsScoped.Group("/scim", adminMw)
	{
		scimAdmin.POST("/tokens", h.CreateToken)
		scimAdmin.GET("/tokens", h.ListTokens)
		scimAdmin.DELETE("/tokens/:tokenId", h.RevokeToken)
		scimAdmin.GET("/sync-logs", h.ListSyncLogs)
	}
}

// RegisterSCIMRoutes registers public SCIM 2.0 protocol endpoints (bearer token auth via middleware).
func (h *SCIMHandler) RegisterSCIMRoutes(router *gin.Engine) {
	scimAuth := scim.AuthMiddleware(h.scimService)
	scimGroup := router.Group("/scim/v2/:workspaceId", scimAuth)
	{
		scimGroup.GET("/ServiceProviderConfig", h.GetServiceProviderConfig)
		scimGroup.GET("/ResourceTypes", h.GetResourceTypes)
		scimGroup.GET("/Schemas", h.GetSchemas)

		scimGroup.GET("/Users", h.ListUsers)
		scimGroup.GET("/Users/:userId", h.GetUser)
		scimGroup.POST("/Users", h.CreateUser)
		scimGroup.PUT("/Users/:userId", h.UpdateUser)
		scimGroup.PATCH("/Users/:userId", h.PatchUser)
		scimGroup.DELETE("/Users/:userId", h.DeleteUser)

		// Groups endpoints (minimal implementation)
		scimGroup.GET("/Groups", h.ListGroups)
	}
}

// --- Admin Token Management ---

func (h *SCIMHandler) CreateToken(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)

	var input models.CreateSCIMTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	resp, err := h.scimService.CreateToken(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, resp)
}

func (h *SCIMHandler) ListTokens(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)

	tokens, err := h.scimService.ListTokens(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, tokens)
}

func (h *SCIMHandler) RevokeToken(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)

	tokenID, err := uuid.Parse(c.Param("tokenId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("tokenId", "invalid token ID"))
		return
	}

	if err := h.scimService.RevokeToken(c.Request.Context(), ws.ID, tokenID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "token revoked"})
}

func (h *SCIMHandler) ListSyncLogs(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, total, err := h.scimService.ListSyncLogs(c.Request.Context(), ws.ID, limit, offset)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondList(c, logs, total, limit, offset)
}

// --- SCIM 2.0 Protocol Endpoints ---

func (h *SCIMHandler) getBaseURL(c *gin.Context) string {
	wsID := c.Param("workspaceId")
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/scim/v2/%s", scheme, c.Request.Host, wsID)
}

func (h *SCIMHandler) GetServiceProviderConfig(c *gin.Context) {
	baseURL := h.getBaseURL(c)
	c.JSON(http.StatusOK, scim.ServiceProviderConfig(baseURL))
}

func (h *SCIMHandler) GetResourceTypes(c *gin.Context) {
	baseURL := h.getBaseURL(c)
	c.JSON(http.StatusOK, scim.ResourceTypes(baseURL))
}

func (h *SCIMHandler) GetSchemas(c *gin.Context) {
	c.JSON(http.StatusOK, scim.SCIMListResponse{
		Schemas:      []string{scim.SchemaListResponse},
		TotalResults: 2,
		ItemsPerPage: 2,
		StartIndex:   1,
		Resources:    []byte(`[]`),
	})
}

func (h *SCIMHandler) ListUsers(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)
	filter := c.Query("filter")
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	resp, err := h.scimService.ListUsers(c.Request.Context(), wsID, filter, startIndex, count)
	if err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SCIMHandler) GetUser(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)
	userID := c.Param("userId")

	user, err := h.scimService.GetUser(c.Request.Context(), wsID, userID)
	if err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *SCIMHandler) CreateUser(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)

	var scimUser scim.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		c.JSON(http.StatusBadRequest, scim.NewSCIMError(400, err.Error()))
		return
	}

	user, err := h.scimService.CreateUser(c.Request.Context(), wsID, scimUser)
	if err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *SCIMHandler) UpdateUser(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)
	userID := c.Param("userId")

	var scimUser scim.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		c.JSON(http.StatusBadRequest, scim.NewSCIMError(400, err.Error()))
		return
	}

	user, err := h.scimService.UpdateUser(c.Request.Context(), wsID, userID, scimUser)
	if err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *SCIMHandler) PatchUser(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)
	userID := c.Param("userId")

	var patch scim.SCIMPatchOp
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, scim.NewSCIMError(400, err.Error()))
		return
	}

	user, err := h.scimService.PatchUser(c.Request.Context(), wsID, userID, patch)
	if err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *SCIMHandler) DeleteUser(c *gin.Context) {
	wsID := scim.GetWorkspaceIDFromContext(c)
	userID := c.Param("userId")

	if err := h.scimService.DeleteUser(c.Request.Context(), wsID, userID); err != nil {
		h.respondSCIMError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListGroups returns an empty list (groups are not yet fully supported).
func (h *SCIMHandler) ListGroups(c *gin.Context) {
	c.JSON(http.StatusOK, scim.SCIMListResponse{
		Schemas:      []string{scim.SchemaListResponse},
		TotalResults: 0,
		ItemsPerPage: 0,
		StartIndex:   1,
		Resources:    []byte(`[]`),
	})
}

func (h *SCIMHandler) respondSCIMError(c *gin.Context, err error) {
	status := httputil.MapToHTTPStatus(err)
	var appErr *httputil.AppError
	if errors.As(err, &appErr) {
		c.JSON(status, scim.NewSCIMError(status, appErr.Message))
		return
	}
	c.JSON(http.StatusInternalServerError, scim.NewSCIMError(500, "internal server error"))
}
