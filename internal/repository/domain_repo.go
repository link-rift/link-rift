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

type DomainRepository interface {
	Create(ctx context.Context, params sqlc.CreateDomainParams) (*models.Domain, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	GetByDomain(ctx context.Context, domain string) (*models.Domain, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error)
	Update(ctx context.Context, params sqlc.UpdateDomainParams) (*models.Domain, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

type domainRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewDomainRepository(queries *sqlc.Queries, logger *zap.Logger) DomainRepository {
	return &domainRepository{queries: queries, logger: logger}
}

func (r *domainRepository) Create(ctx context.Context, params sqlc.CreateDomainParams) (*models.Domain, error) {
	d, err := r.queries.CreateDomain(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("domain")
		}
		return nil, httputil.Wrap(err, "failed to create domain")
	}
	return models.DomainFromSqlc(d), nil
}

func (r *domainRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	d, err := r.queries.GetDomainByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("domain")
		}
		return nil, httputil.Wrap(err, "failed to get domain")
	}
	return models.DomainFromSqlc(d), nil
}

func (r *domainRepository) GetByDomain(ctx context.Context, domain string) (*models.Domain, error) {
	d, err := r.queries.GetDomainByDomain(ctx, domain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("domain")
		}
		return nil, httputil.Wrap(err, "failed to get domain by name")
	}
	return models.DomainFromSqlc(d), nil
}

func (r *domainRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error) {
	rows, err := r.queries.ListDomainsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list domains")
	}

	domains := make([]*models.Domain, 0, len(rows))
	for _, row := range rows {
		domains = append(domains, models.DomainFromSqlc(row))
	}
	return domains, nil
}

func (r *domainRepository) Update(ctx context.Context, params sqlc.UpdateDomainParams) (*models.Domain, error) {
	d, err := r.queries.UpdateDomain(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("domain")
		}
		return nil, httputil.Wrap(err, "failed to update domain")
	}
	return models.DomainFromSqlc(d), nil
}

func (r *domainRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteDomain(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete domain")
	}
	return nil
}

func (r *domainRepository) GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	count, err := r.queries.GetDomainCountForWorkspace(ctx, workspaceID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to get domain count")
	}
	return count, nil
}
