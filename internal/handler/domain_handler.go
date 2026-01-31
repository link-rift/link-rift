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

type DomainHandler struct {
	domainService service.DomainService
	logger        *zap.Logger
}

func NewDomainHandler(domainService service.DomainService, logger *zap.Logger) *DomainHandler {
	return &DomainHandler{domainService: domainService, logger: logger}
}

func (h *DomainHandler) RegisterRoutes(wsScoped *gin.RouterGroup, editorMw gin.HandlerFunc) {
	domains := wsScoped.Group("/domains")
	{
		domains.GET("", h.ListDomains)
		domains.GET("/:id", h.GetDomain)
		domains.GET("/:id/dns-records", h.GetDNSRecords)

		domains.POST("", editorMw, h.AddDomain)
		domains.POST("/:id/verify", editorMw, h.VerifyDomain)
		domains.DELETE("/:id", editorMw, h.RemoveDomain)
	}
}

func (h *DomainHandler) AddDomain(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.CreateDomainInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	domain, err := h.domainService.AddDomain(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, domain)
}

func (h *DomainHandler) ListDomains(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	domains, err := h.domainService.ListDomains(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, domains)
}

func (h *DomainHandler) GetDomain(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid domain ID"))
		return
	}

	domain, err := h.domainService.GetDomain(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, domain)
}

func (h *DomainHandler) VerifyDomain(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid domain ID"))
		return
	}

	domain, err := h.domainService.VerifyDomain(c.Request.Context(), id, ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, domain)
}

func (h *DomainHandler) RemoveDomain(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid domain ID"))
		return
	}

	if err := h.domainService.RemoveDomain(c.Request.Context(), id, ws.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "domain deleted successfully"})
}

func (h *DomainHandler) GetDNSRecords(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid domain ID"))
		return
	}

	records, err := h.domainService.GetDNSRecords(c.Request.Context(), id)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, records)
}
