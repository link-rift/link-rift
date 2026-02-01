package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

// SCIMToken represents a SCIM bearer token for a workspace.
type SCIMToken struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	TokenPrefix string     `json:"token_prefix"`
	Name        string     `json:"name"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// CreateSCIMTokenInput is the request body for creating a SCIM token.
type CreateSCIMTokenInput struct {
	Name      string  `json:"name" binding:"required,min=1,max=255"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// CreateSCIMTokenResponse includes the full token (shown only once).
type CreateSCIMTokenResponse struct {
	Token     string     `json:"token"`
	SCIMToken *SCIMToken `json:"scim_token"`
}

// SCIMSyncLog represents a SCIM sync log entry.
type SCIMSyncLog struct {
	ID           uuid.UUID       `json:"id"`
	WorkspaceID  uuid.UUID       `json:"workspace_id"`
	Operation    string          `json:"operation"`
	ResourceType string          `json:"resource_type"`
	ExternalID   string          `json:"external_id,omitempty"`
	Status       string          `json:"status"`
	Details      json.RawMessage `json:"details"`
	CreatedAt    time.Time       `json:"created_at"`
}

// SCIMTokenFromSqlc converts a sqlc ScimToken to domain model.
func SCIMTokenFromSqlc(t sqlc.ScimToken) *SCIMToken {
	token := &SCIMToken{
		ID:          t.ID,
		WorkspaceID: t.WorkspaceID,
		TokenPrefix: t.TokenPrefix,
		Name:        t.Name,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Time,
	}
	if t.LastUsedAt.Valid {
		lastUsed := t.LastUsedAt.Time
		token.LastUsedAt = &lastUsed
	}
	if t.ExpiresAt.Valid {
		exp := t.ExpiresAt.Time
		token.ExpiresAt = &exp
	}
	return token
}

// SCIMSyncLogFromSqlc converts a sqlc ScimSyncLog to domain model.
func SCIMSyncLogFromSqlc(l sqlc.ScimSyncLog) *SCIMSyncLog {
	return &SCIMSyncLog{
		ID:           l.ID,
		WorkspaceID:  l.WorkspaceID,
		Operation:    l.Operation,
		ResourceType: l.ResourceType,
		ExternalID:   l.ExternalID.String,
		Status:       l.Status,
		Details:      l.Details,
		CreatedAt:    l.CreatedAt.Time,
	}
}
