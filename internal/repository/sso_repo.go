package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type SSORepository interface {
	// Config
	GetConfig(ctx context.Context, workspaceID uuid.UUID) (*models.SSOConfig, error)
	CreateConfig(ctx context.Context, params sqlc.CreateSSOConfigParams) (*models.SSOConfig, error)
	UpdateConfig(ctx context.Context, params sqlc.UpdateSSOConfigParams) (*models.SSOConfig, error)
	DeleteConfig(ctx context.Context, workspaceID uuid.UUID) error

	// Identity
	GetIdentityByExternalID(ctx context.Context, workspaceID uuid.UUID, externalID string) (*models.SSOIdentity, error)
	GetIdentityByUserID(ctx context.Context, userID, workspaceID uuid.UUID) (*models.SSOIdentity, error)
	CreateIdentity(ctx context.Context, params sqlc.CreateSSOIdentityParams) (*models.SSOIdentity, error)
	UpdateIdentityLastLogin(ctx context.Context, userID, workspaceID uuid.UUID, email, name string, attrs json.RawMessage) error
	ListIdentities(ctx context.Context, workspaceID uuid.UUID) ([]*models.SSOIdentity, error)
	DeleteIdentity(ctx context.Context, id uuid.UUID) error
}

type ssoRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewSSORepository(queries *sqlc.Queries, logger *zap.Logger) SSORepository {
	return &ssoRepository{queries: queries, logger: logger}
}

func (r *ssoRepository) GetConfig(ctx context.Context, workspaceID uuid.UUID) (*models.SSOConfig, error) {
	cfg, err := r.queries.GetSSOConfigByWorkspace(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("sso_config")
		}
		return nil, httputil.Wrap(err, "failed to get SSO config")
	}
	return models.SSOConfigFromSqlc(cfg), nil
}

func (r *ssoRepository) CreateConfig(ctx context.Context, params sqlc.CreateSSOConfigParams) (*models.SSOConfig, error) {
	cfg, err := r.queries.CreateSSOConfig(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create SSO config")
	}
	return models.SSOConfigFromSqlc(cfg), nil
}

func (r *ssoRepository) UpdateConfig(ctx context.Context, params sqlc.UpdateSSOConfigParams) (*models.SSOConfig, error) {
	cfg, err := r.queries.UpdateSSOConfig(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("sso_config")
		}
		return nil, httputil.Wrap(err, "failed to update SSO config")
	}
	return models.SSOConfigFromSqlc(cfg), nil
}

func (r *ssoRepository) DeleteConfig(ctx context.Context, workspaceID uuid.UUID) error {
	if err := r.queries.DeleteSSOConfig(ctx, workspaceID); err != nil {
		return httputil.Wrap(err, "failed to delete SSO config")
	}
	return nil
}

func (r *ssoRepository) GetIdentityByExternalID(ctx context.Context, workspaceID uuid.UUID, externalID string) (*models.SSOIdentity, error) {
	id, err := r.queries.GetSSOIdentityByExternalID(ctx, sqlc.GetSSOIdentityByExternalIDParams{
		WorkspaceID: workspaceID,
		ExternalID:  externalID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("sso_identity")
		}
		return nil, httputil.Wrap(err, "failed to get SSO identity")
	}
	return models.SSOIdentityFromSqlc(id), nil
}

func (r *ssoRepository) GetIdentityByUserID(ctx context.Context, userID, workspaceID uuid.UUID) (*models.SSOIdentity, error) {
	id, err := r.queries.GetSSOIdentityByUserID(ctx, sqlc.GetSSOIdentityByUserIDParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("sso_identity")
		}
		return nil, httputil.Wrap(err, "failed to get SSO identity")
	}
	return models.SSOIdentityFromSqlc(id), nil
}

func (r *ssoRepository) CreateIdentity(ctx context.Context, params sqlc.CreateSSOIdentityParams) (*models.SSOIdentity, error) {
	id, err := r.queries.CreateSSOIdentity(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create SSO identity")
	}
	return models.SSOIdentityFromSqlc(id), nil
}

func (r *ssoRepository) UpdateIdentityLastLogin(ctx context.Context, userID, workspaceID uuid.UUID, email, name string, attrs json.RawMessage) error {
	return r.queries.UpdateSSOIdentityLastLogin(ctx, sqlc.UpdateSSOIdentityLastLoginParams{
		UserID:        userID,
		WorkspaceID:   workspaceID,
		Email:         email,
		Name:          pgtype.Text{String: name, Valid: name != ""},
		RawAttributes: attrs,
	})
}

func (r *ssoRepository) ListIdentities(ctx context.Context, workspaceID uuid.UUID) ([]*models.SSOIdentity, error) {
	rows, err := r.queries.ListSSOIdentitiesForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list SSO identities")
	}
	identities := make([]*models.SSOIdentity, 0, len(rows))
	for _, row := range rows {
		identities = append(identities, models.SSOIdentityFromSqlc(row))
	}
	return identities, nil
}

func (r *ssoRepository) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteSSOIdentity(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to delete SSO identity")
	}
	return nil
}
