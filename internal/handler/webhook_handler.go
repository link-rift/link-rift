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

type WebhookHandler struct {
	webhookService service.WebhookService
	logger         *zap.Logger
}

func NewWebhookHandler(webhookService service.WebhookService, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService, logger: logger}
}

func (h *WebhookHandler) RegisterRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	webhooks := wsScoped.Group("/webhooks")
	{
		webhooks.GET("", h.ListWebhooks)
		webhooks.POST("", adminMw, h.CreateWebhook)
		webhooks.DELETE("/:id", adminMw, h.DeleteWebhook)
		webhooks.GET("/:id/deliveries", h.ListDeliveries)
	}
}

func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.CreateWebhookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	result, err := h.webhookService.CreateWebhook(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, result)
}

func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	webhooks, err := h.webhookService.ListWebhooks(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, webhooks)
}

func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid webhook ID"))
		return
	}

	if err := h.webhookService.DeleteWebhook(c.Request.Context(), id, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "webhook deleted successfully"})
}

func (h *WebhookHandler) ListDeliveries(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	webhookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid webhook ID"))
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

	deliveries, total, err := h.webhookService.ListDeliveries(
		c.Request.Context(), webhookID, ws.ID,
		int32(pagination.Limit), int32(pagination.Offset),
	)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondList(c, deliveries, total, pagination.Limit, pagination.Offset)
}
