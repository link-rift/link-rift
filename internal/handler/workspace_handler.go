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

type WorkspaceHandler struct {
	wsService service.WorkspaceService
	logger    *zap.Logger
}

func NewWorkspaceHandler(wsService service.WorkspaceService, logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{wsService: wsService, logger: logger}
}

// RegisterRoutes registers workspace routes under the given router group.
// wsAccessMw must be applied to workspace-scoped routes.
func (h *WorkspaceHandler) RegisterRoutes(v1 *gin.RouterGroup, authMw gin.HandlerFunc, wsAccessMw gin.HandlerFunc) {
	workspaces := v1.Group("/workspaces", authMw)
	{
		workspaces.POST("", h.CreateWorkspace)
		workspaces.GET("", h.ListWorkspaces)
	}

	ws := workspaces.Group("/:workspaceId", wsAccessMw)
	{
		ws.GET("", h.GetWorkspace)

		adminMw := middleware.RequireWorkspaceRole(models.RoleAdmin)
		ownerMw := middleware.RequireWorkspaceRole(models.RoleOwner)

		ws.PUT("", adminMw, h.UpdateWorkspace)
		ws.DELETE("", ownerMw, h.DeleteWorkspace)

		ws.GET("/members", h.ListMembers)
		ws.POST("/members", adminMw, h.InviteMember)
		ws.PUT("/members/:userId", adminMw, h.UpdateMemberRole)
		ws.DELETE("/members/:userId", adminMw, h.RemoveMember)
		ws.POST("/transfer", ownerMw, h.TransferOwnership)
	}
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	var input models.CreateWorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	ws, err := h.wsService.CreateWorkspace(c.Request.Context(), user.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, ws.ToResponse(1, models.RoleOwner))
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		httputil.RespondError(c, httputil.Unauthorized("not authenticated"))
		return
	}

	workspaces, err := h.wsService.ListWorkspaces(c.Request.Context(), user.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	responses := make([]*models.WorkspaceResponse, 0, len(workspaces))
	for _, ws := range workspaces {
		memberCount, _ := h.wsService.GetMemberCount(c.Request.Context(), ws.ID)
		member, _ := h.wsService.GetMember(c.Request.Context(), ws.ID, user.ID)
		role := models.RoleViewer
		if member != nil {
			role = member.Role
		}
		responses = append(responses, ws.ToResponse(memberCount, role))
	}

	httputil.RespondSuccess(c, http.StatusOK, responses)
}

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	member := middleware.GetWorkspaceMemberFromContext(c)
	if ws == nil || member == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	memberCount, _ := h.wsService.GetMemberCount(c.Request.Context(), ws.ID)

	httputil.RespondSuccess(c, http.StatusOK, ws.ToResponse(memberCount, member.Role))
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.UpdateWorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	updated, err := h.wsService.UpdateWorkspace(c.Request.Context(), ws.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	member := middleware.GetWorkspaceMemberFromContext(c)
	memberCount, _ := h.wsService.GetMemberCount(c.Request.Context(), ws.ID)
	role := models.RoleViewer
	if member != nil {
		role = member.Role
	}

	httputil.RespondSuccess(c, http.StatusOK, updated.ToResponse(memberCount, role))
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	user := middleware.GetUserFromContext(c)
	if ws == nil || user == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	if err := h.wsService.DeleteWorkspace(c.Request.Context(), ws.ID, user.ID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "workspace deleted successfully"})
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	if ws == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	members, err := h.wsService.ListMembers(c.Request.Context(), ws.ID)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, members)
}

func (h *WorkspaceHandler) InviteMember(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	user := middleware.GetUserFromContext(c)
	if ws == nil || user == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.InviteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	member, err := h.wsService.InviteMember(c.Request.Context(), ws.ID, user.ID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusCreated, member)
}

func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	user := middleware.GetUserFromContext(c)
	if ws == nil || user == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("userId", "invalid user ID"))
		return
	}

	var input models.UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	member, err := h.wsService.UpdateMemberRole(c.Request.Context(), ws.ID, user.ID, targetUserID, input)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, member)
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	user := middleware.GetUserFromContext(c)
	if ws == nil || user == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		httputil.RespondError(c, httputil.Validation("userId", "invalid user ID"))
		return
	}

	if err := h.wsService.RemoveMember(c.Request.Context(), ws.ID, user.ID, targetUserID); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "member removed successfully"})
}

func (h *WorkspaceHandler) TransferOwnership(c *gin.Context) {
	ws := middleware.GetWorkspaceFromContext(c)
	user := middleware.GetUserFromContext(c)
	if ws == nil || user == nil {
		httputil.RespondError(c, httputil.Forbidden("workspace access required"))
		return
	}

	var input models.TransferOwnershipInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.RespondError(c, httputil.Validation("body", err.Error()))
		return
	}

	if err := h.wsService.TransferOwnership(c.Request.Context(), ws.ID, user.ID, input); err != nil {
		httputil.RespondError(c, err)
		return
	}

	httputil.RespondSuccess(c, http.StatusOK, gin.H{"message": "ownership transferred successfully"})
}
