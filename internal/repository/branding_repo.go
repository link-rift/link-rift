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

type BrandingRepository interface {
	Get(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceBranding, error)
	Upsert(ctx context.Context, params sqlc.UpsertWorkspaceBrandingParams) (*models.WorkspaceBranding, error)
	Delete(ctx context.Context, workspaceID uuid.UUID) error
}

type brandingRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewBrandingRepository(queries *sqlc.Queries, logger *zap.Logger) BrandingRepository {
	return &brandingRepository{queries: queries, logger: logger}
}

func (r *brandingRepository) Get(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceBranding, error) {
	b, err := r.queries.GetWorkspaceBranding(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace_branding")
		}
		return nil, httputil.Wrap(err, "failed to get workspace branding")
	}
	return models.BrandingFromSqlc(b), nil
}

func (r *brandingRepository) Upsert(ctx context.Context, params sqlc.UpsertWorkspaceBrandingParams) (*models.WorkspaceBranding, error) {
	b, err := r.queries.UpsertWorkspaceBranding(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to upsert workspace branding")
	}
	return models.BrandingFromSqlc(b), nil
}

func (r *brandingRepository) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	if err := r.queries.DeleteWorkspaceBranding(ctx, workspaceID); err != nil {
		return httputil.Wrap(err, "failed to delete workspace branding")
	}
	return nil
}
