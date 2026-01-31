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

type BioPageHandler struct {
	bioPageService service.BioPageService
	logger         *zap.Logger
}

func NewBioPageHandler(bioPageService service.BioPageService, logger *zap.Logger) *BioPageHandler {
	return &BioPageHandler{bioPageService: bioPageService, logger: logger}
}

func (h *BioPageHandler) RegisterRoutes(wsScoped *gin.RouterGroup, editorMw gin.HandlerFunc) {
	bioPages := wsScoped.Group("/bio-pages")
	{
		bioPages.GET("", h.ListBioPages)
		bioPages.GET("/:id", h.GetBioPage)
		bioPages.GET("/:id/links", h.ListLinks)

		bioPages.POST("", editorMw, h.CreateBioPage)
		bioPages.PUT("/:id", editorMw, h.UpdateBioPage)
		bioPages.DELETE("/:id", editorMw, h.DeleteBioPage)
		bioPages.POST("/:id/publish", editorMw, h.PublishBioPage)
		bioPages.POST("/:id/unpublish", editorMw, h.UnpublishBioPage)
		bioPages.POST("/:id/links", editorMw, h.AddLink)
		bioPages.PUT("/:id/links/:linkId", editorMw, h.UpdateLink)
		bioPages.DELETE("/:id/links/:linkId", editorMw, h.DeleteLink)
		bioPages.POST("/:id/links/reorder", editorMw, h.ReorderLinks)
	}

	themes := wsScoped.Group("/bio-themes")
	{
		themes.GET("", h.ListThemes)
		themes.GET("/:themeId", h.GetTheme)
	}
}

func (h *BioPageHandler) RegisterPublicRoutes(router *gin.Engine) {
	router.GET("/b/:slug", h.GetPublicPage)
	router.POST("/b/:slug/click/:linkId", h.TrackLinkClick)
}

// Bio Page CRUD

func (h *BioPageHandler) CreateBioPage(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.CreateBioPageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	page, err := h.bioPageService.CreateBioPage(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, page)
}

func (h *BioPageHandler) ListBioPages(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	pages, err := h.bioPageService.ListBioPages(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, pages)
}

func (h *BioPageHandler) GetBioPage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	page, err := h.bioPageService.GetBioPage(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, page)
}

func (h *BioPageHandler) UpdateBioPage(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	var input models.UpdateBioPageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	page, err := h.bioPageService.UpdateBioPage(c.Request.Context(), id, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, page)
}

func (h *BioPageHandler) DeleteBioPage(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	if err := h.bioPageService.DeleteBioPage(c.Request.Context(), id, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "bio page deleted successfully"})
}

// Publish

func (h *BioPageHandler) PublishBioPage(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	page, err := h.bioPageService.PublishBioPage(c.Request.Context(), id, ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, page)
}

func (h *BioPageHandler) UnpublishBioPage(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	page, err := h.bioPageService.UnpublishBioPage(c.Request.Context(), id, ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, page)
}

// Links

func (h *BioPageHandler) ListLinks(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	links, err := h.bioPageService.ListLinks(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, links)
}

func (h *BioPageHandler) AddLink(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	var input models.CreateBioPageLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	link, err := h.bioPageService.AddLink(c.Request.Context(), pageID, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, link)
}

func (h *BioPageHandler) UpdateLink(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	linkID, err := uuid.Parse(c.Param("linkId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("linkId", "invalid link ID"))
		return
	}

	var input models.UpdateBioPageLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	link, err := h.bioPageService.UpdateLink(c.Request.Context(), pageID, linkID, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, link)
}

func (h *BioPageHandler) DeleteLink(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	linkID, err := uuid.Parse(c.Param("linkId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("linkId", "invalid link ID"))
		return
	}

	if err := h.bioPageService.DeleteLink(c.Request.Context(), pageID, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "link deleted successfully"})
}

func (h *BioPageHandler) ReorderLinks(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid bio page ID"))
		return
	}

	var input models.ReorderBioLinksInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	if err := h.bioPageService.ReorderLinks(c.Request.Context(), pageID, ws.ID, input); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "links reordered successfully"})
}

// Themes

func (h *BioPageHandler) ListThemes(c *gin.Context) {
	themes := h.bioPageService.ListThemes()
	httputil.RespondSuccess(c, http.StatusOK, themes)
}

func (h *BioPageHandler) GetTheme(c *gin.Context) {
	themeID := c.Param("themeId")
	theme, err := h.bioPageService.GetTheme(themeID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, theme)
}

// Public

func (h *BioPageHandler) GetPublicPage(c *gin.Context) {
	slug := c.Param("slug")

	page, err := h.bioPageService.GetPublicPage(c.Request.Context(), slug)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, page)
}

func (h *BioPageHandler) TrackLinkClick(c *gin.Context) {
	linkID, err := uuid.Parse(c.Param("linkId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("linkId", "invalid link ID"))
		return
	}

	// Fire-and-forget
	if err := h.bioPageService.TrackLinkClick(c.Request.Context(), linkID); err != nil {
		h.logger.Warn("failed to track bio link click", zap.Error(err))
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"tracked": true})
}
