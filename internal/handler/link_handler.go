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

type LinkHandler struct {
	linkService service.LinkService
	logger      *zap.Logger
}

func NewLinkHandler(linkService service.LinkService, logger *zap.Logger) *LinkHandler {
	return &LinkHandler{linkService: linkService, logger: logger}
}

func (h *LinkHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	links := rg.Group("/links", authMw)
	{
		links.POST("", h.CreateLink)
		links.GET("", h.ListLinks)
		links.GET("/:id", h.GetLink)
		links.PUT("/:id", h.UpdateLink)
		links.DELETE("/:id", h.DeleteLink)
		links.POST("/bulk", h.BulkCreateLinks)
		links.GET("/:id/stats", h.GetQuickStats)
	}
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	var input models.CreateLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	workspaceID, err := getWorkspaceID(c)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	link, err := h.linkService.CreateLink(c.Request.Context(), user.ID, workspaceID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, link)
}

func (h *LinkHandler) ListLinks(c *gin.Context) {
	workspaceID, err := getWorkspaceID(c)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	var filter models.LinkFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		httputil.RespondError(c, httputil.Validation("query", err.Error()))
		return
	}

	var pagination models.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		httputil.RespondError(c, httputil.Validation("query", err.Error()))
		return
	}
	if pagination.Limit == 0 {
		pagination.Limit = 20
	}

	result, err := h.linkService.ListLinks(c.Request.Context(), workspaceID, filter, pagination)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondList(c, result.Links, result.Total, pagination.Limit, pagination.Offset)
}

func (h *LinkHandler) GetLink(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	link, err := h.linkService.GetLink(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, link)
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	var input models.UpdateLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	link, err := h.linkService.UpdateLink(c.Request.Context(), id, user.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, link)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.linkService.DeleteLink(c.Request.Context(), id, user.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "link deleted successfully"})
}

func (h *LinkHandler) BulkCreateLinks(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	var input models.BulkCreateLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	workspaceID, err := getWorkspaceID(c)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	links, err := h.linkService.BulkCreateLinks(c.Request.Context(), user.ID, workspaceID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, links)
}

func (h *LinkHandler) GetQuickStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	stats, err := h.linkService.GetQuickStats(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

// getWorkspaceID extracts workspace_id from query param. Will be replaced by workspace middleware in Phase 5.
func getWorkspaceID(c *gin.Context) (uuid.UUID, error) {
	wsID := c.Query("workspace_id")
	if wsID == "" {
		return uuid.Nil, httputil.Validation("workspace_id", "workspace_id query parameter is required")
	}
	id, err := uuid.Parse(wsID)
	if err != nil {
		return uuid.Nil, httputil.Validation("workspace_id", "invalid workspace_id format")
	}
	return id, nil
}
