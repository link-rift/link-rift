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

type QRHandler struct {
	qrService service.QRCodeService
	logger    *zap.Logger
}

func NewQRHandler(qrService service.QRCodeService, logger *zap.Logger) *QRHandler {
	return &QRHandler{qrService: qrService, logger: logger}
}

func (h *QRHandler) RegisterRoutes(wsScoped *gin.RouterGroup, editorMw gin.HandlerFunc) {
	links := wsScoped.Group("/links")
	{
		links.POST("/:id/qr", editorMw, h.CreateQRCode)
		links.GET("/:id/qr", h.GetQRCodeForLink)
		links.GET("/:id/qr/download", h.DownloadQRCode)
	}

	qr := wsScoped.Group("/qr")
	{
		qr.POST("/bulk", editorMw, h.BulkGenerateQRCodes)
		qr.GET("/templates", h.GetStyleTemplates)
	}
}

func (h *QRHandler) CreateQRCode(c *gin.Context) {
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

	var input models.CreateQRCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	qr, err := h.qrService.CreateQRCode(c.Request.Context(), linkID, ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, qr.ToResponse())
}

func (h *QRHandler) GetQRCodeForLink(c *gin.Context) {
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	qr, err := h.qrService.GetQRCodeForLink(c.Request.Context(), linkID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, qr.ToResponse())
}

func (h *QRHandler) DownloadQRCode(c *gin.Context) {
	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("id", "invalid link ID"))
		return
	}

	format := c.DefaultQuery("format", "png")

	data, contentType, err := h.qrService.DownloadQRCode(c.Request.Context(), linkID, format)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	filename := "qrcode." + format
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

func (h *QRHandler) BulkGenerateQRCodes(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.BulkQRCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	result, err := h.qrService.BulkGenerateQRCodes(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	// Return ZIP file
	c.Header("Content-Disposition", "attachment; filename=qr_codes.zip")
	c.Data(http.StatusOK, "application/zip", result.ZipData)
}

func (h *QRHandler) GetStyleTemplates(c *gin.Context) {
	templates := h.qrService.GetStyleTemplates()
	httputil.RespondSuccess(c, http.StatusOK, templates)
}
