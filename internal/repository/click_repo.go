package repository

import (
	"context"

	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type ClickRepository interface {
	Insert(ctx context.Context, params sqlc.InsertClickParams) error
	GetByLinkID(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error)
}

type clickRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewClickRepository(queries *sqlc.Queries, logger *zap.Logger) ClickRepository {
	return &clickRepository{queries: queries, logger: logger}
}

func (r *clickRepository) Insert(ctx context.Context, params sqlc.InsertClickParams) error {
	err := r.queries.InsertClick(ctx, params)
	if err != nil {
		return httputil.Wrap(err, "failed to insert click")
	}
	return nil
}

func (r *clickRepository) GetByLinkID(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error) {
	rows, err := r.queries.GetClicksByLinkID(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to get clicks")
	}

	clicks := make([]*models.Click, 0, len(rows))
	for _, row := range rows {
		clicks = append(clicks, models.ClickFromSqlc(row))
	}

	return clicks, nil
}
