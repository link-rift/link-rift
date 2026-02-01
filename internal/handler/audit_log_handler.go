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

type AuditLogHandler struct {
	auditLogService service.AuditLogService
	logger          *zap.Logger
}

func NewAuditLogHandler(auditLogService service.AuditLogService, logger *zap.Logger) *AuditLogHandler {
	return &AuditLogHandler{auditLogService: auditLogService, logger: logger}
}

func (h *AuditLogHandler) RegisterRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	auditLogs := wsScoped.Group("/audit-logs")
	{
		auditLogs.GET("", adminMw, h.ListAuditLogs)
		auditLogs.GET("/export", adminMw, h.ExportAuditLogs)
		auditLogs.GET("/:id", adminMw, h.GetAuditLog)
	}
}

func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var filter models.AuditLogFilter
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

	result, err := h.auditLogService.ListAuditLogs(c.Request.Context(), ws.ID, filter, pagination)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondList(c, result.AuditLogs, result.Total, pagination.Limit, pagination.Offset)
}

func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid audit log ID"))
		return
	}

	log, err := h.auditLogService.GetAuditLog(c.Request.Context(), id, ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, log)
}

func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var filter models.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		httputil.RespondError(c, httputil.Validation("query", err.Error()))
		return
	}

	format := c.DefaultQuery("format", "json")

	data, contentType, err := h.auditLogService.ExportAuditLogs(c.Request.Context(), ws.ID, filter, format)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	ext := "json"
	if format == "csv" {
		ext = "csv"
	}

	c.Header("Content-Disposition", "attachment; filename=audit-logs."+ext)
	c.Data(http.StatusOK, contentType, data)
}
