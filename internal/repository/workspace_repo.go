package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, params sqlc.CreateWorkspaceParams) (*models.Workspace, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
	GetBySlug(ctx context.Context, slug string) (*models.Workspace, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error)
	Update(ctx context.Context, params sqlc.UpdateWorkspaceParams) (*models.Workspace, error)
	UpdateOwner(ctx context.Context, params sqlc.UpdateWorkspaceOwnerParams) (*models.Workspace, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetCountForUser(ctx context.Context, userID uuid.UUID) (int64, error)
}

type workspaceRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewWorkspaceRepository(queries *sqlc.Queries, logger *zap.Logger) WorkspaceRepository {
	return &workspaceRepository{queries: queries, logger: logger}
}

func (r *workspaceRepository) Create(ctx context.Context, params sqlc.CreateWorkspaceParams) (*models.Workspace, error) {
	w, err := r.queries.CreateWorkspace(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("workspace")
		}
		return nil, httputil.Wrap(err, "failed to create workspace")
	}
	return models.WorkspaceFromSqlc(w), nil
}

func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	w, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace")
		}
		return nil, httputil.Wrap(err, "failed to get workspace")
	}
	return models.WorkspaceFromSqlc(w), nil
}

func (r *workspaceRepository) GetBySlug(ctx context.Context, slug string) (*models.Workspace, error) {
	w, err := r.queries.GetWorkspaceBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace")
		}
		return nil, httputil.Wrap(err, "failed to get workspace by slug")
	}
	return models.WorkspaceFromSqlc(w), nil
}

func (r *workspaceRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	rows, err := r.queries.ListWorkspacesForUser(ctx, userID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list workspaces")
	}

	workspaces := make([]*models.Workspace, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, models.WorkspaceFromSqlc(row))
	}

	return workspaces, nil
}

func (r *workspaceRepository) Update(ctx context.Context, params sqlc.UpdateWorkspaceParams) (*models.Workspace, error) {
	w, err := r.queries.UpdateWorkspace(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("workspace slug")
		}
		return nil, httputil.Wrap(err, "failed to update workspace")
	}
	return models.WorkspaceFromSqlc(w), nil
}

func (r *workspaceRepository) UpdateOwner(ctx context.Context, params sqlc.UpdateWorkspaceOwnerParams) (*models.Workspace, error) {
	w, err := r.queries.UpdateWorkspaceOwner(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace")
		}
		return nil, httputil.Wrap(err, "failed to update workspace owner")
	}
	return models.WorkspaceFromSqlc(w), nil
}

func (r *workspaceRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteWorkspace(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete workspace")
	}
	return nil
}

func (r *workspaceRepository) GetCountForUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.queries.GetWorkspaceCountForUser(ctx, userID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get workspace count")
	}
	return count, nil
}
