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

type LinkRepository interface {
	Create(ctx context.Context, params sqlc.CreateLinkParams) (*models.Link, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Link, error)
	GetByShortCode(ctx context.Context, shortCode string) (*models.Link, error)
	GetByURL(ctx context.Context, params sqlc.GetLinkByURLParams) (*models.Link, error)
	List(ctx context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error)
	Update(ctx context.Context, params sqlc.UpdateLinkParams) (*models.Link, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)
	IncrementClicks(ctx context.Context, id uuid.UUID) error
	IncrementUniqueClicks(ctx context.Context, id uuid.UUID) error
	GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error)
	GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

type linkRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewLinkRepository(queries *sqlc.Queries, logger *zap.Logger) LinkRepository {
	return &linkRepository{queries: queries, logger: logger}
}

func (r *linkRepository) Create(ctx context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
	l, err := r.queries.CreateLink(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("short_code")
		}
		return nil, httputil.Wrap(err, "failed to create link")
	}
	return models.LinkFromSqlc(l), nil
}

func (r *linkRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	l, err := r.queries.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("link")
		}
		return nil, httputil.Wrap(err, "failed to get link")
	}
	return models.LinkFromSqlc(l), nil
}

func (r *linkRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.Link, error) {
	l, err := r.queries.GetLinkByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("link")
		}
		return nil, httputil.Wrap(err, "failed to get link by short code")
	}
	return models.LinkFromSqlc(l), nil
}

func (r *linkRepository) GetByURL(ctx context.Context, params sqlc.GetLinkByURLParams) (*models.Link, error) {
	l, err := r.queries.GetLinkByURL(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("link")
		}
		return nil, httputil.Wrap(err, "failed to get link by URL")
	}
	return models.LinkFromSqlc(l), nil
}

func (r *linkRepository) List(ctx context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error) {
	rows, err := r.queries.ListLinksForWorkspace(ctx, params)
	if err != nil {
		return nil, 0, httputil.Wrap(err, "failed to list links")
	}

	var total int64
	links := make([]*models.Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, models.LinkFromSqlcRow(row))
		total = row.TotalCount
	}

	return links, total, nil
}

func (r *linkRepository) Update(ctx context.Context, params sqlc.UpdateLinkParams) (*models.Link, error) {
	l, err := r.queries.UpdateLink(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("link")
		}
		return nil, httputil.Wrap(err, "failed to update link")
	}
	return models.LinkFromSqlc(l), nil
}

func (r *linkRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteLink(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete link")
	}
	return nil
}

func (r *linkRepository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	exists, err := r.queries.ShortCodeExists(ctx, shortCode)
	if err != nil {
		return false, httputil.Wrap(err, "failed to check short code")
	}
	return exists, nil
}

func (r *linkRepository) IncrementClicks(ctx context.Context, id uuid.UUID) error {
	err := r.queries.IncrementLinkClicks(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to increment clicks")
	}
	return nil
}

func (r *linkRepository) IncrementUniqueClicks(ctx context.Context, id uuid.UUID) error {
	err := r.queries.IncrementLinkUniqueClicks(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to increment unique clicks")
	}
	return nil
}

func (r *linkRepository) GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
	row, err := r.queries.GetLinkQuickStats(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("link")
		}
		return nil, httputil.Wrap(err, "failed to get link stats")
	}

	stats := &models.LinkQuickStats{
		TotalClicks:  row.TotalClicks,
		UniqueClicks: row.UniqueClicks,
		Clicks24h:    row.Clicks24h,
		Clicks7d:     row.Clicks7d,
	}
	if row.CreatedAt.Valid {
		stats.CreatedAt = row.CreatedAt.Time
	}

	return stats, nil
}

func (r *linkRepository) GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	count, err := r.queries.GetLinkCountForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get link count")
	}
	return count, nil
}
