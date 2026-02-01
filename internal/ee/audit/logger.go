package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/internal/service"
	"go.uber.org/zap"
)

// Action constants for audit logging.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionRevoke = "revoke"
)

// ResourceType constants for audit logging.
const (
	ResourceLink      = "link"
	ResourceWorkspace = "workspace"
	ResourceMember    = "member"
	ResourceDomain    = "domain"
	ResourceAPIKey    = "api_key"
	ResourceWebhook   = "webhook"
	ResourceBioPage   = "bio_page"
	ResourceBranding  = "branding"
	ResourceSSOConfig = "sso_config"
	ResourceSCIMToken = "scim_token"
)

// AuditLogWriter is the interface for writing audit log entries to storage.
type AuditLogWriter interface {
	CreateAuditLog(ctx context.Context, arg sqlc.CreateAuditLogParams) error
}

type auditLogger struct {
	writer     AuditLogWriter
	licManager *license.Manager
	logger     *zap.Logger
}

// NewAuditLogger creates a new AuditLogger that implements service.AuditLogger.
// It no-ops when the audit_logs feature is disabled.
func NewAuditLogger(writer AuditLogWriter, licManager *license.Manager, logger *zap.Logger) service.AuditLogger {
	return &auditLogger{
		writer:     writer,
		licManager: licManager,
		logger:     logger,
	}
}

func (a *auditLogger) Log(ctx context.Context, entry service.AuditEntry) {
	if !a.licManager.HasFeature(license.FeatureAuditLogs) {
		return
	}

	oldJSON := marshalJSON(entry.OldValues)
	newJSON := marshalJSON(entry.NewValues)
	metaJSON := marshalJSON(entry.Metadata)

	params := sqlc.CreateAuditLogParams{
		WorkspaceID:  entry.WorkspaceID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		OldValues:    oldJSON,
		NewValues:    newJSON,
		Metadata:     metaJSON,
		IpAddress:    entry.IPAddress,
	}

	if entry.UserID != uuid.Nil {
		params.UserID = pgtype.UUID{Bytes: entry.UserID, Valid: true}
	}
	if entry.ResourceID != uuid.Nil {
		params.ResourceID = pgtype.UUID{Bytes: entry.ResourceID, Valid: true}
	}
	if entry.UserAgent != "" {
		params.UserAgent = pgtype.Text{String: entry.UserAgent, Valid: true}
	}

	if err := a.writer.CreateAuditLog(ctx, params); err != nil {
		a.logger.Error("failed to write audit log",
			zap.String("action", entry.Action),
			zap.String("resource_type", entry.ResourceType),
			zap.Error(err),
		)
	}
}

func marshalJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
