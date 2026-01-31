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

type APIKeyHandler struct {
	apiKeyService service.APIKeyService
	logger        *zap.Logger
}

func NewAPIKeyHandler(apiKeyService service.APIKeyService, logger *zap.Logger) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService, logger: logger}
}

func (h *APIKeyHandler) RegisterRoutes(wsScoped *gin.RouterGroup, adminMw gin.HandlerFunc) {
	apiKeys := wsScoped.Group("/api-keys")
	{
		apiKeys.GET("", h.ListAPIKeys)
		apiKeys.POST("", adminMw, h.CreateAPIKey)
		apiKeys.DELETE("/:id", adminMw, h.RevokeAPIKey)
	}
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
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

	var input models.CreateAPIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	result, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), user.ID, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, result)
}

func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	keys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, keys)
}

func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid API key ID"))
		return
	}

	if err := h.apiKeyService.RevokeAPIKey(c.Request.Context(), id, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "API key revoked successfully"})
}
