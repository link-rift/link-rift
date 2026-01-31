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

type UserRepository interface {
	Create(ctx context.Context, params sqlc.CreateUserParams) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, params sqlc.UpdateUserParams) (*models.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SetEmailVerified(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewUserRepository(queries *sqlc.Queries, logger *zap.Logger) UserRepository {
	return &userRepository{queries: queries, logger: logger}
}

func (r *userRepository) Create(ctx context.Context, params sqlc.CreateUserParams) (*models.User, error) {
	u, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httputil.AlreadyExists("user")
		}
		return nil, httputil.Wrap(err, "failed to create user")
	}
	return models.UserFromSqlc(u), nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("user")
		}
		return nil, httputil.Wrap(err, "failed to get user")
	}
	return models.UserFromSqlc(u), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("user")
		}
		return nil, httputil.Wrap(err, "failed to get user")
	}
	return models.UserFromSqlc(u), nil
}

func (r *userRepository) Update(ctx context.Context, params sqlc.UpdateUserParams) (*models.User, error) {
	u, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("user")
		}
		return nil, httputil.Wrap(err, "failed to update user")
	}
	return models.UserFromSqlc(u), nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	err := r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return httputil.Wrap(err, "failed to update password")
	}
	return nil
}

func (r *userRepository) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SetEmailVerified(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to set email verified")
	}
	return nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete user")
	}
	return nil
}
