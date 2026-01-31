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

type SessionRepository interface {
	Create(ctx context.Context, params sqlc.CreateSessionParams) (*models.Session, error)
	GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*models.Session, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewSessionRepository(queries *sqlc.Queries, logger *zap.Logger) SessionRepository {
	return &sessionRepository{queries: queries, logger: logger}
}

func (r *sessionRepository) Create(ctx context.Context, params sqlc.CreateSessionParams) (*models.Session, error) {
	s, err := r.queries.CreateSession(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create session")
	}
	return models.SessionFromSqlc(s), nil
}

func (r *sessionRepository) GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	s, err := r.queries.GetSessionByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("session")
		}
		return nil, httputil.Wrap(err, "failed to get session")
	}
	return models.SessionFromSqlc(s), nil
}

func (r *sessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	sessions, err := r.queries.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list sessions")
	}

	result := make([]*models.Session, len(sessions))
	for i, s := range sessions {
		result[i] = models.SessionFromSqlc(s)
	}
	return result, nil
}

func (r *sessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	err := r.queries.RevokeSession(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to revoke session")
	}
	return nil
}

func (r *sessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	err := r.queries.RevokeAllUserSessions(ctx, userID)
	if err != nil {
		return httputil.Wrap(err, "failed to revoke all sessions")
	}
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	err := r.queries.DeleteExpiredSessions(ctx)
	if err != nil {
		return httputil.Wrap(err, "failed to delete expired sessions")
	}
	return nil
}
