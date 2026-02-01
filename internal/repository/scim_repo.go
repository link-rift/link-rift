package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// SCIMRepository handles SCIM token and sync log persistence.
type SCIMRepository interface {
	// Tokens
	CreateToken(ctx context.Context, params sqlc.CreateSCIMTokenParams) (*models.SCIMToken, error)
	GetTokenByHash(ctx context.Context, tokenHash string) (*models.SCIMToken, error)
	ListTokens(ctx context.Context, workspaceID uuid.UUID) ([]*models.SCIMToken, error)
	RevokeToken(ctx context.Context, id uuid.UUID) error
	UpdateTokenLastUsed(ctx context.Context, id uuid.UUID) error
	DeleteToken(ctx context.Context, id uuid.UUID) error

	// Sync Log
	CreateSyncLog(ctx context.Context, params sqlc.CreateSCIMSyncLogParams) error
	ListSyncLogs(ctx context.Context, params sqlc.ListSCIMSyncLogsParams) ([]*models.SCIMSyncLog, int64, error)
}

type scimRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

// NewSCIMRepository creates a new SCIM repository.
func NewSCIMRepository(queries *sqlc.Queries, logger *zap.Logger) SCIMRepository {
	return &scimRepository{queries: queries, logger: logger}
}

func (r *scimRepository) CreateToken(ctx context.Context, params sqlc.CreateSCIMTokenParams) (*models.SCIMToken, error) {
	token, err := r.queries.CreateSCIMToken(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create SCIM token")
	}
	return models.SCIMTokenFromSqlc(token), nil
}

func (r *scimRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*models.SCIMToken, error) {
	token, err := r.queries.GetSCIMTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("scim_token")
		}
		return nil, httputil.Wrap(err, "failed to get SCIM token")
	}
	return models.SCIMTokenFromSqlc(token), nil
}

func (r *scimRepository) ListTokens(ctx context.Context, workspaceID uuid.UUID) ([]*models.SCIMToken, error) {
	rows, err := r.queries.ListSCIMTokensForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list SCIM tokens")
	}
	tokens := make([]*models.SCIMToken, 0, len(rows))
	for _, row := range rows {
		tokens = append(tokens, models.SCIMTokenFromSqlc(row))
	}
	return tokens, nil
}

func (r *scimRepository) RevokeToken(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.RevokeSCIMToken(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to revoke SCIM token")
	}
	return nil
}

func (r *scimRepository) UpdateTokenLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateSCIMTokenLastUsed(ctx, id)
}

func (r *scimRepository) DeleteToken(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteSCIMToken(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to delete SCIM token")
	}
	return nil
}

func (r *scimRepository) CreateSyncLog(ctx context.Context, params sqlc.CreateSCIMSyncLogParams) error {
	return r.queries.CreateSCIMSyncLog(ctx, params)
}

func (r *scimRepository) ListSyncLogs(ctx context.Context, params sqlc.ListSCIMSyncLogsParams) ([]*models.SCIMSyncLog, int64, error) {
	rows, err := r.queries.ListSCIMSyncLogs(ctx, params)
	if err != nil {
		return nil, 0, httputil.Wrap(err, "failed to list SCIM sync logs")
	}
	count, err := r.queries.CountSCIMSyncLogs(ctx, params.WorkspaceID)
	if err != nil {
		return nil, 0, httputil.Wrap(err, "failed to count SCIM sync logs")
	}
	logs := make([]*models.SCIMSyncLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, models.SCIMSyncLogFromSqlc(row))
	}
	return logs, count, nil
}
