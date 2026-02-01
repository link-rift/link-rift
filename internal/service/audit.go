package service

import (
	"context"

	"github.com/google/uuid"
)

// AuditLogger is a service-level interface for audit logging,
// decoupled from the ee/audit package to avoid circular imports.
type AuditLogger interface {
	Log(ctx context.Context, entry AuditEntry)
}

// AuditEntry mirrors audit.AuditEntry for use within the service layer.
type AuditEntry struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Action       string
	ResourceType string
	ResourceID   uuid.UUID
	OldValues    any
	NewValues    any
	Metadata     map[string]any
	IPAddress    string
	UserAgent    string
}

// noopAuditLogger is the default no-op implementation.
type noopAuditLogger struct{}

func (noopAuditLogger) Log(_ context.Context, _ AuditEntry) {}
