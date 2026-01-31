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

// RegisterRoutes registers link routes under a workspace-scoped router group.
// editorMw enforces editor+ role for write operations.
func (h *LinkHandler) RegisterRoutes(wsScoped *gin.RouterGroup, editorMw gin.HandlerFunc) {
	links := wsScoped.Group("/links")
	{
		links.GET("", h.ListLinks)
		links.GET("/:id", h.GetLink)
		links.GET("/:id/stats", h.GetQuickStats)

		links.POST("", editorMw, h.CreateLink)
		links.PUT("/:id", editorMw, h.UpdateLink)
		links.DELETE("/:id", editorMw, h.DeleteLink)
		links.POST("/bulk", editorMw, h.BulkCreateLinks)
	}
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.CreateLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	link, err := h.linkService.CreateLink(c.Request.Context(), user.ID, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, link)
}

func (h *LinkHandler) ListLinks(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
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

	result, err := h.linkService.ListLinks(c.Request.Context(), ws.ID, filter, pagination)
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
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
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

	link, err := h.linkService.UpdateLink(c.Request.Context(), id, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, link)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.linkService.DeleteLink(c.Request.Context(), id, ws.ID); err != nil {
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

	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.BulkCreateLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	links, err := h.linkService.BulkCreateLinks(c.Request.Context(), user.ID, ws.ID, input)
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
