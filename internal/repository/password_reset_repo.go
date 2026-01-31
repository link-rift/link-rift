package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, params sqlc.CreatePasswordResetParams) (sqlc.PasswordReset, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (sqlc.PasswordReset, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type passwordResetRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewPasswordResetRepository(queries *sqlc.Queries, logger *zap.Logger) PasswordResetRepository {
	return &passwordResetRepository{queries: queries, logger: logger}
}

func (r *passwordResetRepository) Create(ctx context.Context, params sqlc.CreatePasswordResetParams) (sqlc.PasswordReset, error) {
	pr, err := r.queries.CreatePasswordReset(ctx, params)
	if err != nil {
		return sqlc.PasswordReset{}, httputil.Wrap(err, "failed to create password reset")
	}
	return pr, nil
}

func (r *passwordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (sqlc.PasswordReset, error) {
	pr, err := r.queries.GetPasswordResetByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PasswordReset{}, httputil.NotFound("password reset")
		}
		return sqlc.PasswordReset{}, httputil.Wrap(err, "failed to get password reset")
	}
	return pr, nil
}

func (r *passwordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	err := r.queries.MarkPasswordResetUsed(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to mark password reset used")
	}
	return nil
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	err := r.queries.DeleteExpiredPasswordResets(ctx)
	if err != nil {
		return httputil.Wrap(err, "failed to delete expired password resets")
	}
	return nil
}
