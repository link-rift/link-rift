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

type BioPageRepository interface {
	// Bio Pages
	Create(ctx context.Context, params sqlc.CreateBioPageParams) (*models.BioPage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.BioPage, error)
	GetBySlug(ctx context.Context, slug string) (*models.BioPage, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]*models.BioPage, error)
	Update(ctx context.Context, params sqlc.UpdateBioPageParams) (*models.BioPage, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error)

	// Bio Page Links
	CreateLink(ctx context.Context, params sqlc.CreateBioPageLinkParams) (*models.BioPageLink, error)
	GetLinkByID(ctx context.Context, id uuid.UUID) (*models.BioPageLink, error)
	ListLinks(ctx context.Context, bioPageID uuid.UUID) ([]*models.BioPageLink, error)
	UpdateLink(ctx context.Context, params sqlc.UpdateBioPageLinkParams) (*models.BioPageLink, error)
	DeleteLink(ctx context.Context, id uuid.UUID) error
	UpdateLinkPosition(ctx context.Context, params sqlc.UpdateBioPageLinkPositionParams) error
	IncrementLinkClickCount(ctx context.Context, id uuid.UUID) error
	GetMaxLinkPosition(ctx context.Context, bioPageID uuid.UUID) (int32, error)
}

type bioPageRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewBioPageRepository(queries *sqlc.Queries, logger *zap.Logger) BioPageRepository {
	return &bioPageRepository{queries: queries, logger: logger}
}

// Bio Pages

func (r *bioPageRepository) Create(ctx context.Context, params sqlc.CreateBioPageParams) (*models.BioPage, error) {
	b, err := r.queries.CreateBioPage(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("bio page slug")
		}
		return nil, httputil.Wrap(err, "failed to create bio page")
	}
	return models.BioPageFromSqlc(b), nil
}

func (r *bioPageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.BioPage, error) {
	b, err := r.queries.GetBioPageByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("bio page")
		}
		return nil, httputil.Wrap(err, "failed to get bio page")
	}
	return models.BioPageFromSqlc(b), nil
}

func (r *bioPageRepository) GetBySlug(ctx context.Context, slug string) (*models.BioPage, error) {
	b, err := r.queries.GetBioPageBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("bio page")
		}
		return nil, httputil.Wrap(err, "failed to get bio page by slug")
	}
	return models.BioPageFromSqlc(b), nil
}

func (r *bioPageRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*models.BioPage, error) {
	rows, err := r.queries.ListBioPagesForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list bio pages")
	}

	pages := make([]*models.BioPage, 0, len(rows))
	for _, row := range rows {
		pages = append(pages, models.BioPageFromSqlc(row))
	}
	return pages, nil
}

func (r *bioPageRepository) Update(ctx context.Context, params sqlc.UpdateBioPageParams) (*models.BioPage, error) {
	b, err := r.queries.UpdateBioPage(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("bio page")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("bio page slug")
		}
		return nil, httputil.Wrap(err, "failed to update bio page")
	}
	return models.BioPageFromSqlc(b), nil
}

func (r *bioPageRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteBioPage(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete bio page")
	}
	return nil
}

func (r *bioPageRepository) GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	count, err := r.queries.GetBioPageCountForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get bio page count")
	}
	return count, nil
}

// Bio Page Links

func (r *bioPageRepository) CreateLink(ctx context.Context, params sqlc.CreateBioPageLinkParams) (*models.BioPageLink, error) {
	l, err := r.queries.CreateBioPageLink(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create bio page link")
	}
	return models.BioPageLinkFromSqlc(l), nil
}

func (r *bioPageRepository) GetLinkByID(ctx context.Context, id uuid.UUID) (*models.BioPageLink, error) {
	l, err := r.queries.GetBioPageLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("bio page link")
		}
		return nil, httputil.Wrap(err, "failed to get bio page link")
	}
	return models.BioPageLinkFromSqlc(l), nil
}

func (r *bioPageRepository) ListLinks(ctx context.Context, bioPageID uuid.UUID) ([]*models.BioPageLink, error) {
	rows, err := r.queries.ListBioPageLinks(ctx, bioPageID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list bio page links")
	}

	links := make([]*models.BioPageLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, models.BioPageLinkFromSqlc(row))
	}
	return links, nil
}

func (r *bioPageRepository) UpdateLink(ctx context.Context, params sqlc.UpdateBioPageLinkParams) (*models.BioPageLink, error) {
	l, err := r.queries.UpdateBioPageLink(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("bio page link")
		}
		return nil, httputil.Wrap(err, "failed to update bio page link")
	}
	return models.BioPageLinkFromSqlc(l), nil
}

func (r *bioPageRepository) DeleteLink(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteBioPageLink(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete bio page link")
	}
	return nil
}

func (r *bioPageRepository) UpdateLinkPosition(ctx context.Context, params sqlc.UpdateBioPageLinkPositionParams) error {
	err := r.queries.UpdateBioPageLinkPosition(ctx, params)
	if err != nil {
		return httputil.Wrap(err, "failed to update link position")
	}
	return nil
}

func (r *bioPageRepository) IncrementLinkClickCount(ctx context.Context, id uuid.UUID) error {
	err := r.queries.IncrementBioPageLinkClickCount(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to increment link click count")
	}
	return nil
}

func (r *bioPageRepository) GetMaxLinkPosition(ctx context.Context, bioPageID uuid.UUID) (int32, error) {
	pos, err := r.queries.GetMaxBioPageLinkPosition(ctx, bioPageID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get max link position")
	}
	return pos, nil
}
