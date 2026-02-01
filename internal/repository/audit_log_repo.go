package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type AuditLogRepository interface {
	Create(ctx context.Context, params sqlc.CreateAuditLogParams) error
	GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*models.AuditLog, error)
	ListFiltered(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, limit, offset int32) ([]*models.AuditLog, int64, error)
	Count(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

type auditLogRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewAuditLogRepository(queries *sqlc.Queries, logger *zap.Logger) AuditLogRepository {
	return &auditLogRepository{queries: queries, logger: logger}
}

func (r *auditLogRepository) Create(ctx context.Context, params sqlc.CreateAuditLogParams) error {
	return r.queries.CreateAuditLog(ctx, params)
}

func (r *auditLogRepository) GetByID(ctx context.Context, id, workspaceID uuid.UUID) (*models.AuditLog, error) {
	a, err := r.queries.GetAuditLogByID(ctx, sqlc.GetAuditLogByIDParams{
		ID:          id,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("audit_log")
		}
		return nil, httputil.Wrap(err, "failed to get audit log")
	}
	return models.AuditLogFromSqlc(a), nil
}

func (r *auditLogRepository) ListFiltered(ctx context.Context, workspaceID uuid.UUID, filter models.AuditLogFilter, limit, offset int32) ([]*models.AuditLog, int64, error) {
	filterParams := buildFilterParams(workspaceID, filter)

	// Get count
	countParams := sqlc.CountAuditLogsFilteredParams{
		WorkspaceID:  filterParams.WorkspaceID,
		Action:       filterParams.Action,
		ResourceType: filterParams.ResourceType,
		UserID:       filterParams.UserID,
		StartDate:    filterParams.StartDate,
		EndDate:      filterParams.EndDate,
	}
	total, err := r.queries.CountAuditLogsFiltered(ctx, countParams)
	if err != nil {
		return nil, 0, httputil.Wrap(err, "failed to count audit logs")
	}

	// Get items
	filterParams.Limit = limit
	filterParams.Offset = offset
	rows, err := r.queries.ListAuditLogsFiltered(ctx, filterParams)
	if err != nil {
		return nil, 0, httputil.Wrap(err, "failed to list audit logs")
	}

	logs := make([]*models.AuditLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, models.AuditLogFromSqlc(row))
	}

	return logs, total, nil
}

func (r *auditLogRepository) Count(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	count, err := r.queries.CountAuditLogsForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to count audit logs")
	}
	return count, nil
}

func buildFilterParams(workspaceID uuid.UUID, filter models.AuditLogFilter) sqlc.ListAuditLogsFilteredParams {
	params := sqlc.ListAuditLogsFilteredParams{
		WorkspaceID: workspaceID,
	}
	if filter.Action != nil {
		params.Action = pgtype.Text{String: *filter.Action, Valid: true}
	}
	if filter.ResourceType != nil {
		params.ResourceType = pgtype.Text{String: *filter.ResourceType, Valid: true}
	}
	if filter.UserID != nil {
		params.UserID = pgtype.UUID{Bytes: *filter.UserID, Valid: true}
	}
	if filter.StartDate != nil {
		params.StartDate = pgtype.Timestamptz{Time: *filter.StartDate, Valid: true}
	}
	if filter.EndDate != nil {
		params.EndDate = pgtype.Timestamptz{Time: *filter.EndDate, Valid: true}
	}
	return params
}
