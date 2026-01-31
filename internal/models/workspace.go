package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

// WorkspaceRole represents a member's role in a workspace.
type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleEditor WorkspaceRole = "editor"
	RoleViewer WorkspaceRole = "viewer"
)

// Level returns a numeric level for role comparison.
func (r WorkspaceRole) Level() int {
	switch r {
	case RoleOwner:
		return 4
	case RoleAdmin:
		return 3
	case RoleEditor:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

// HasPermission checks if this role meets the minimum required role.
func (r WorkspaceRole) HasPermission(minRole WorkspaceRole) bool {
	return r.Level() >= minRole.Level()
}

// IsValid returns true if the role is a known workspace role.
func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

// Permission represents an action that can be performed in a workspace.
type Permission string

const (
	PermissionView          Permission = "view"
	PermissionCreateLinks   Permission = "create_links"
	PermissionUpdateLinks   Permission = "update_links"
	PermissionDeleteLinks   Permission = "delete_links"
	PermissionUpdateSettings Permission = "update_settings"
	PermissionManageMembers Permission = "manage_members"
	PermissionDeleteWorkspace Permission = "delete_workspace"
	PermissionTransferOwnership Permission = "transfer_ownership"
)

var permissionMatrix = map[Permission]WorkspaceRole{
	PermissionView:              RoleViewer,
	PermissionCreateLinks:       RoleEditor,
	PermissionUpdateLinks:       RoleEditor,
	PermissionDeleteLinks:       RoleEditor,
	PermissionUpdateSettings:    RoleAdmin,
	PermissionManageMembers:     RoleAdmin,
	PermissionDeleteWorkspace:   RoleOwner,
	PermissionTransferOwnership: RoleOwner,
}

// CheckPermission checks if a role has a given permission.
func CheckPermission(role WorkspaceRole, perm Permission) bool {
	minRole, ok := permissionMatrix[perm]
	if !ok {
		return false
	}
	return role.HasPermission(minRole)
}

type Workspace struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	OwnerID   uuid.UUID       `json:"owner_id"`
	Plan      string          `json:"plan"`
	Settings  json.RawMessage `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type WorkspaceResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Slug            string          `json:"slug"`
	OwnerID         uuid.UUID       `json:"owner_id"`
	Plan            string          `json:"plan"`
	Settings        json.RawMessage `json:"settings"`
	MemberCount     int64           `json:"member_count"`
	CurrentUserRole WorkspaceRole   `json:"current_user_role"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateWorkspaceInput struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
	Slug string `json:"slug" binding:"required,min=1,max=100,alphanumunicode"`
}

type UpdateWorkspaceInput struct {
	Name *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Slug *string `json:"slug,omitempty" binding:"omitempty,min=1,max=100,alphanumunicode"`
}

func WorkspaceFromSqlc(w sqlc.Workspace) *Workspace {
	ws := &Workspace{
		ID:      w.ID,
		Name:    w.Name,
		Slug:    w.Slug,
		OwnerID: w.OwnerID,
		Plan:    w.Plan,
	}
	if w.Settings != nil {
		ws.Settings = w.Settings
	}
	if w.CreatedAt.Valid {
		ws.CreatedAt = w.CreatedAt.Time
	}
	if w.UpdatedAt.Valid {
		ws.UpdatedAt = w.UpdatedAt.Time
	}
	return ws
}

func (w *Workspace) ToResponse(memberCount int64, currentUserRole WorkspaceRole) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:              w.ID,
		Name:            w.Name,
		Slug:            w.Slug,
		OwnerID:         w.OwnerID,
		Plan:            w.Plan,
		Settings:        w.Settings,
		MemberCount:     memberCount,
		CurrentUserRole: currentUserRole,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	}
}
