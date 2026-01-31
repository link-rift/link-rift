package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/middleware"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	linkService      service.LinkService
	logger           *zap.Logger
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService, linkService service.LinkService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		linkService:      linkService,
		logger:           logger,
	}
}

// RegisterRoutes registers analytics routes under a workspace-scoped group.
func (h *AnalyticsHandler) RegisterRoutes(wsScoped *gin.RouterGroup) {
	analytics := wsScoped.Group("/analytics")
	{
		analytics.GET("/links/:id", h.GetLinkStats)
		analytics.GET("/links/:id/timeseries", h.GetTimeSeries)
		analytics.GET("/links/:id/referrers", h.GetReferrers)
		analytics.GET("/links/:id/countries", h.GetCountries)
		analytics.GET("/links/:id/devices", h.GetDevices)
		analytics.GET("/links/:id/browsers", h.GetBrowsers)
		analytics.GET("/workspace", h.GetWorkspaceStats)
		analytics.GET("/export", h.ExportData)
	}
}

func (h *AnalyticsHandler) GetLinkStats(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	stats, err := h.analyticsService.GetLinkStats(c.Request.Context(), linkID, dr)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetTimeSeries(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	interval := h.parseInterval(c)

	points, err := h.analyticsService.GetTimeSeries(c.Request.Context(), linkID, interval, dr)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, points)
}

func (h *AnalyticsHandler) GetReferrers(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	limit := h.parseLimit(c)

	stats, err := h.analyticsService.GetTopReferrers(c.Request.Context(), linkID, dr, limit)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetCountries(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	limit := h.parseLimit(c)

	stats, err := h.analyticsService.GetTopCountries(c.Request.Context(), linkID, dr, limit)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetDevices(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)

	breakdown, err := h.analyticsService.GetDeviceBreakdown(c.Request.Context(), linkID, dr)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, breakdown)
}

func (h *AnalyticsHandler) GetBrowsers(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	limit := h.parseLimit(c)

	stats, err := h.analyticsService.GetBrowserBreakdown(c.Request.Context(), linkID, dr, limit)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

func (h *AnalyticsHandler) GetWorkspaceStats(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	dr := h.parseDateRange(c)

	stats, err := h.analyticsService.GetWorkspaceStats(c.Request.Context(), ws.ID, dr)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, stats)
}

func (h *AnalyticsHandler) ExportData(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	linkIDStr := c.Query("link_id")
	if linkIDStr == "" {
		httputil.RespondError(c, httputil.Validation("link_id", "link_id is required"))
		return
	}

	linkID, err := uuid.Parse(linkIDStr)
	if err != nil {
		httputil.RespondError(c, httputil.Validation("link_id", "invalid link ID"))
		return
	}

	if err := h.verifyLinkOwnership(c, linkID, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	dr := h.parseDateRange(c)
	format := models.AnalyticsExportFormat(c.DefaultQuery("format", "csv"))

	data, contentType, err := h.analyticsService.ExportLinkData(c.Request.Context(), linkID, dr, format)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	filename := "analytics-export"
	if format == models.ExportCSV {
		filename += ".csv"
	} else {
		filename += ".json"
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

// verifyLinkOwnership checks that the link belongs to the workspace.
func (h *AnalyticsHandler) verifyLinkOwnership(c *gin.Context, linkID, workspaceID uuid.UUID) error {
	link, err := h.linkService.GetLink(c.Request.Context(), linkID)
	if err != nil {
		return err
	}
	if link.WorkspaceID != workspaceID {
		return httputil.Forbidden("link does not belong to this workspace")
	}
	return nil
}

func (h *AnalyticsHandler) parseDateRange(c *gin.Context) models.DateRange {
	if preset := c.Query("range"); preset != "" {
		return models.DateRangeFromPreset(preset)
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	now := time.Now().UTC()
	dr := models.DateRange{
		Start: now.Add(-7 * 24 * time.Hour),
		End:   now,
	}

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			dr.Start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			dr.End = t
		}
	}

	return dr
}

func (h *AnalyticsHandler) parseInterval(c *gin.Context) models.TimeSeriesInterval {
	switch c.DefaultQuery("interval", "day") {
	case "hour":
		return models.IntervalHour
	case "week":
		return models.IntervalWeek
	case "month":
		return models.IntervalMonth
	default:
		return models.IntervalDay
	}
}

func (h *AnalyticsHandler) parseLimit(c *gin.Context) int {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}
