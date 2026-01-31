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

type WorkspaceMemberRepository interface {
	Add(ctx context.Context, params sqlc.AddWorkspaceMemberParams) (*models.WorkspaceMember, error)
	Get(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error)
	UpdateRole(ctx context.Context, params sqlc.UpdateMemberRoleParams) (*models.WorkspaceMember, error)
	Remove(ctx context.Context, workspaceID, userID uuid.UUID) error
	GetCount(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

type workspaceMemberRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewWorkspaceMemberRepository(queries *sqlc.Queries, logger *zap.Logger) WorkspaceMemberRepository {
	return &workspaceMemberRepository{queries: queries, logger: logger}
}

func (r *workspaceMemberRepository) Add(ctx context.Context, params sqlc.AddWorkspaceMemberParams) (*models.WorkspaceMember, error) {
	m, err := r.queries.AddWorkspaceMember(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("workspace member")
		}
		return nil, httputil.Wrap(err, "failed to add workspace member")
	}
	return models.WorkspaceMemberFromSqlc(m), nil
}

func (r *workspaceMemberRepository) Get(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	m, err := r.queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace member")
		}
		return nil, httputil.Wrap(err, "failed to get workspace member")
	}
	return models.WorkspaceMemberFromSqlc(m), nil
}

func (r *workspaceMemberRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
	rows, err := r.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list workspace members")
	}

	members := make([]*models.WorkspaceMemberResponse, 0, len(rows))
	for _, row := range rows {
		members = append(members, models.WorkspaceMemberResponseFromSqlcRow(row))
	}

	return members, nil
}

func (r *workspaceMemberRepository) UpdateRole(ctx context.Context, params sqlc.UpdateMemberRoleParams) (*models.WorkspaceMember, error) {
	m, err := r.queries.UpdateMemberRole(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("workspace member")
		}
		return nil, httputil.Wrap(err, "failed to update member role")
	}
	return models.WorkspaceMemberFromSqlc(m), nil
}

func (r *workspaceMemberRepository) Remove(ctx context.Context, workspaceID, userID uuid.UUID) error {
	err := r.queries.RemoveWorkspaceMember(ctx, sqlc.RemoveWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		return httputil.Wrap(err, "failed to remove workspace member")
	}
	return nil
}

func (r *workspaceMemberRepository) GetCount(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	count, err := r.queries.GetMemberCountForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get member count")
	}
	return count, nil
}
