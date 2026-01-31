package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/link-rift/link-rift/internal/realtime"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/paseto"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate against allowed origins
	},
}

type WebSocketHandler struct {
	hub        *realtime.Hub
	tokenMaker paseto.Maker
	memberRepo repository.WorkspaceMemberRepository
	logger     *zap.Logger
}

func NewWebSocketHandler(
	hub *realtime.Hub,
	tokenMaker paseto.Maker,
	memberRepo repository.WorkspaceMemberRepository,
	logger *zap.Logger,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:        hub,
		tokenMaker: tokenMaker,
		memberRepo: memberRepo,
		logger:     logger,
	}
}

// RegisterRoutes registers the WebSocket endpoint.
func (h *WebSocketHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/ws/analytics", h.HandleWebSocket)
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Authenticate via query param token (WebSocket doesn't support headers)
	token := c.Query("token")
	if token == "" {
		httputil.RespondError(c, httputil.Unauthorized("token is required"))
		return
	}

	claims, err := h.tokenMaker.VerifyToken(token)
	if err != nil {
		httputil.RespondError(c, httputil.Unauthorized("invalid or expired token"))
		return
	}

	// Parse workspace ID
	wsIDStr := c.Query("workspace_id")
	if wsIDStr == "" {
		httputil.RespondError(c, httputil.Validation("workspace_id", "workspace_id is required"))
		return
	}

	workspaceID, err := uuid.Parse(wsIDStr)
	if err != nil {
		httputil.RespondError(c, httputil.Validation("workspace_id", "invalid workspace ID"))
		return
	}

	// Verify workspace membership
	_, err = h.memberRepo.Get(c.Request.Context(), workspaceID, claims.UserID)
	if err != nil {
		httputil.RespondError(c, httputil.Forbidden("not a member of this workspace"))
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := realtime.NewClient(h.hub, conn, workspaceID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
