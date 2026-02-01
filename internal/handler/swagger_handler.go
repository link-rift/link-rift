package handler

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger_ui.html
var swaggerUI embed.FS

// SwaggerHandler serves OpenAPI spec and Swagger UI.
type SwaggerHandler struct {
	specData []byte
}

// NewSwaggerHandler creates a new SwaggerHandler with the given OpenAPI spec bytes.
func NewSwaggerHandler(specData []byte) *SwaggerHandler {
	return &SwaggerHandler{specData: specData}
}

// RegisterRoutes registers the Swagger UI and spec routes.
func (h *SwaggerHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/api/docs", h.serveUI)
	r.GET("/api/docs/openapi.yaml", h.serveSpec)
}

func (h *SwaggerHandler) serveUI(c *gin.Context) {
	data, err := swaggerUI.ReadFile("swagger_ui.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "swagger UI not found")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

func (h *SwaggerHandler) serveSpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/x-yaml", h.specData)
}
