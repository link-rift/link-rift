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

type APIKeyRepository interface {
	Create(ctx context.Context, params sqlc.CreateAPIKeyParams) (*models.APIKey, error)
	GetByPrefix(ctx context.Context, prefix string) (*models.APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]*models.APIKey, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

type apiKeyRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewAPIKeyRepository(queries *sqlc.Queries, logger *zap.Logger) APIKeyRepository {
	return &apiKeyRepository{queries: queries, logger: logger}
}

func (r *apiKeyRepository) Create(ctx context.Context, params sqlc.CreateAPIKeyParams) (*models.APIKey, error) {
	k, err := r.queries.CreateAPIKey(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create API key")
	}
	return models.APIKeyFromSqlc(k), nil
}

func (r *apiKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	k, err := r.queries.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("api_key")
		}
		return nil, httputil.Wrap(err, "failed to get API key by prefix")
	}
	return models.APIKeyFromSqlc(k), nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	k, err := r.queries.GetAPIKeyByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("api_key")
		}
		return nil, httputil.Wrap(err, "failed to get API key")
	}
	return models.APIKeyFromSqlc(k), nil
}

func (r *apiKeyRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*models.APIKey, error) {
	wsID := pgtype.UUID{Bytes: workspaceID, Valid: true}
	keys, err := r.queries.ListAPIKeysForWorkspace(ctx, wsID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list API keys")
	}

	result := make([]*models.APIKey, 0, len(keys))
	for _, k := range keys {
		result = append(result, models.APIKeyFromSqlc(k))
	}
	return result, nil
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RevokeAPIKey(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to revoke API key")
	}
	return nil
}

func (r *apiKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	err := r.queries.UpdateAPIKeyLastUsed(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to update API key last used")
	}
	return nil
}
