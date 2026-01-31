package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/pkg/httputil"
)

const (
	contextKeyWorkspace       = "workspace"
	contextKeyWorkspaceMember = "workspace_member"
)

// RequireWorkspaceAccess extracts the workspace ID from the :workspaceId path param,
// verifies the workspace exists and the user is a member, then injects both into context.
func RequireWorkspaceAccess(wsRepo repository.WorkspaceRepository, memberRepo repository.WorkspaceMemberRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "not authenticated",
				},
			})
			return
		}

		wsIDStr := c.Param("workspaceId")
		wsID, err := uuid.Parse(wsIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "VALIDATION_ERROR",
					Message: "invalid workspace ID",
				},
			})
			return
		}

		ws, err := wsRepo.GetByID(c.Request.Context(), wsID)
		if err != nil {
			status := httputil.MapToHTTPStatus(err)
			var appErr *httputil.AppError
			if ok := err.(*httputil.AppError); ok != nil {
				appErr = ok
			}
			if appErr != nil {
				c.AbortWithStatusJSON(status, httputil.Response{
					Success: false,
					Error: &httputil.ErrorBody{
						Code:    appErr.Code,
						Message: appErr.Message,
					},
				})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, httputil.Response{
					Success: false,
					Error: &httputil.ErrorBody{
						Code:    "INTERNAL_ERROR",
						Message: "failed to get workspace",
					},
				})
			}
			return
		}

		member, err := memberRepo.Get(c.Request.Context(), wsID, user.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: "you are not a member of this workspace",
				},
			})
			return
		}

		c.Set(contextKeyWorkspace, ws)
		c.Set(contextKeyWorkspaceMember, member)
		c.Next()
	}
}

// RequireWorkspaceRole checks that the current member has at least the specified role.
func RequireWorkspaceRole(minRole models.WorkspaceRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		member := GetWorkspaceMemberFromContext(c)
		if member == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: "workspace access required",
				},
			})
			return
		}

		if !member.Role.HasPermission(minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: "insufficient workspace permissions",
				},
			})
			return
		}

		c.Next()
	}
}

// GetWorkspaceFromContext returns the workspace from the gin context.
func GetWorkspaceFromContext(c *gin.Context) *models.Workspace {
	val, exists := c.Get(contextKeyWorkspace)
	if !exists {
		return nil
	}
	ws, ok := val.(*models.Workspace)
	if !ok {
		return nil
	}
	return ws
}

// GetWorkspaceMemberFromContext returns the workspace member from the gin context.
func GetWorkspaceMemberFromContext(c *gin.Context) *models.WorkspaceMember {
	val, exists := c.Get(contextKeyWorkspaceMember)
	if !exists {
		return nil
	}
	member, ok := val.(*models.WorkspaceMember)
	if !ok {
		return nil
	}
	return member
}
