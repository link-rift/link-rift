package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type AuditLogService interface {
	ListAuditLogs(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, pagination models.Pagination) (*models.AuditLogListResult, error)
	GetAuditLog(ctx context.Context, id, workspaceID uuid.UUID) (*models.AuditLog, error)
	ExportAuditLogs(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, format string) ([]byte, string, error)
}

type auditLogService struct {
	auditLogRepo repository.AuditLogRepository
	licManager   *license.Manager
	logger       *zap.Logger
}

func NewAuditLogService(
	auditLogRepo repository.AuditLogRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) AuditLogService {
	return &auditLogService{
		auditLogRepo: auditLogRepo,
		licManager:   licManager,
		logger:       logger,
	}
}

func (s *auditLogService) ListAuditLogs(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, pagination models.Pagination) (*models.AuditLogListResult, error) {
	if !s.licManager.HasFeature(license.FeatureAuditLogs) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAuditLogs), string(license.TierEnterprise))
	}

	logs, total, err := s.auditLogRepo.ListFiltered(ctx, workspaceID, filter, int32(pagination.Limit), int32(pagination.Offset))
	if err != nil {
		return nil, err
	}

	return &models.AuditLogListResult{
		AuditLogs: logs,
		Total:     total,
	}, nil
}

func (s *auditLogService) GetAuditLog(ctx context.Context, id, workspaceID uuid.UUID) (*models.AuditLog, error) {
	if !s.licManager.HasFeature(license.FeatureAuditLogs) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAuditLogs), string(license.TierEnterprise))
	}

	return s.auditLogRepo.GetByID(ctx, id, workspaceID)
}

func (s *auditLogService) ExportAuditLogs(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, format string) ([]byte, string, error) {
	if !s.licManager.HasFeature(license.FeatureAuditLogs) {
		return nil, "", httputil.PaymentRequiredWithDetails(string(license.FeatureAuditLogs), string(license.TierEnterprise))
	}

	// Export up to 10000 records
	logs, _, err := s.auditLogRepo.ListFiltered(ctx, workspaceID, filter, 10000, 0)
	if err != nil {
		return nil, "", err
	}

	switch format {
	case "csv":
		return s.exportCSV(logs)
	default:
		return s.exportJSON(logs)
	}
}

func (s *auditLogService) exportJSON(logs []*models.AuditLog) ([]byte, string, error) {
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return nil, "", httputil.Wrap(err, "failed to marshal audit logs")
	}
	return data, "application/json", nil
}

func (s *auditLogService) exportCSV(logs []*models.AuditLog) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	if err := w.Write([]string{
		"id", "workspace_id", "user_id", "action", "resource_type", "resource_id",
		"ip_address", "user_agent", "created_at",
	}); err != nil {
		return nil, "", httputil.Wrap(err, "failed to write CSV header")
	}

	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}
		resourceID := ""
		if log.ResourceID != nil {
			resourceID = log.ResourceID.String()
		}

		if err := w.Write([]string{
			log.ID.String(),
			log.WorkspaceID.String(),
			userID,
			log.Action,
			log.ResourceType,
			resourceID,
			log.IPAddress,
			log.UserAgent,
			fmt.Sprintf("%s", log.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}); err != nil {
			return nil, "", httputil.Wrap(err, "failed to write CSV row")
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", httputil.Wrap(err, "failed to flush CSV")
	}

	return buf.Bytes(), "text/csv", nil
}
