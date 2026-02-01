package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	WorkspaceID  uuid.UUID       `json:"workspace_id"`
	UserID       *uuid.UUID      `json:"user_id,omitempty"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	OldValues    json.RawMessage `json:"old_values,omitempty"`
	NewValues    json.RawMessage `json:"new_values,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	IPAddress    string          `json:"ip_address,omitempty"`
	UserAgent    string          `json:"user_agent,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AuditLogFilter struct {
	Action       *string    `form:"action"`
	ResourceType *string    `form:"resource_type"`
	UserID       *uuid.UUID `form:"user_id"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
}

type AuditLogListResult struct {
	AuditLogs []*AuditLog `json:"audit_logs"`
	Total     int64       `json:"total"`
}

func AuditLogFromSqlc(a sqlc.AuditLog) *AuditLog {
	log := &AuditLog{
		ID:           a.ID,
		WorkspaceID:  a.WorkspaceID,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		IPAddress:    a.IpAddress,
	}

	if a.UserID.Valid {
		id := uuid.UUID(a.UserID.Bytes)
		log.UserID = &id
	}
	if a.ResourceID.Valid {
		id := uuid.UUID(a.ResourceID.Bytes)
		log.ResourceID = &id
	}
	if len(a.OldValues) > 0 {
		log.OldValues = json.RawMessage(a.OldValues)
	}
	if len(a.NewValues) > 0 {
		log.NewValues = json.RawMessage(a.NewValues)
	}
	if len(a.Metadata) > 0 {
		log.Metadata = json.RawMessage(a.Metadata)
	}
	if a.UserAgent.Valid {
		log.UserAgent = a.UserAgent.String
	}
	if a.CreatedAt.Valid {
		log.CreatedAt = a.CreatedAt.Time
	}

	return log
}
