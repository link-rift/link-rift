package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type WorkspaceMember struct {
	ID          uuid.UUID     `json:"id"`
	WorkspaceID uuid.UUID     `json:"workspace_id"`
	UserID      uuid.UUID     `json:"user_id"`
	Role        WorkspaceRole `json:"role"`
	InvitedBy   *uuid.UUID    `json:"invited_by,omitempty"`
	JoinedAt    *time.Time    `json:"joined_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

type WorkspaceMemberResponse struct {
	ID          uuid.UUID     `json:"id"`
	WorkspaceID uuid.UUID     `json:"workspace_id"`
	UserID      uuid.UUID     `json:"user_id"`
	Role        WorkspaceRole `json:"role"`
	Email       string        `json:"email"`
	Name        string        `json:"name"`
	AvatarURL   *string       `json:"avatar_url,omitempty"`
	JoinedAt    *time.Time    `json:"joined_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

type InviteMemberInput struct {
	Email string        `json:"email" binding:"required,email"`
	Role  WorkspaceRole `json:"role" binding:"required"`
}

type UpdateMemberRoleInput struct {
	Role WorkspaceRole `json:"role" binding:"required"`
}

type TransferOwnershipInput struct {
	NewOwnerID uuid.UUID `json:"new_owner_id" binding:"required"`
}

func WorkspaceMemberFromSqlc(m sqlc.WorkspaceMember) *WorkspaceMember {
	wm := &WorkspaceMember{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        WorkspaceRole(m.Role),
	}
	if m.InvitedBy.Valid {
		id := uuid.UUID(m.InvitedBy.Bytes)
		wm.InvitedBy = &id
	}
	if m.JoinedAt.Valid {
		t := m.JoinedAt.Time
		wm.JoinedAt = &t
	}
	if m.CreatedAt.Valid {
		wm.CreatedAt = m.CreatedAt.Time
	}
	return wm
}

func WorkspaceMemberResponseFromSqlcRow(r sqlc.ListWorkspaceMembersRow) *WorkspaceMemberResponse {
	resp := &WorkspaceMemberResponse{
		ID:          r.ID,
		WorkspaceID: r.WorkspaceID,
		UserID:      r.UserID,
		Role:        WorkspaceRole(r.Role),
		Email:       r.Email,
		Name:        r.UserName,
	}
	if r.AvatarUrl.Valid {
		resp.AvatarURL = &r.AvatarUrl.String
	}
	if r.JoinedAt.Valid {
		t := r.JoinedAt.Time
		resp.JoinedAt = &t
	}
	if r.CreatedAt.Valid {
		resp.CreatedAt = r.CreatedAt.Time
	}
	return resp
}
