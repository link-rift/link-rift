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

type QRCodeRepository interface {
	Create(ctx context.Context, params sqlc.CreateQRCodeParams) (*models.QRCode, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.QRCode, error)
	GetByLinkID(ctx context.Context, linkID uuid.UUID) (*models.QRCode, error)
	ListForLink(ctx context.Context, linkID uuid.UUID) ([]*models.QRCode, error)
	Update(ctx context.Context, params sqlc.UpdateQRCodeParams) (*models.QRCode, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementScanCount(ctx context.Context, id uuid.UUID) error
}

type qrCodeRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewQRCodeRepository(queries *sqlc.Queries, logger *zap.Logger) QRCodeRepository {
	return &qrCodeRepository{queries: queries, logger: logger}
}

func (r *qrCodeRepository) Create(ctx context.Context, params sqlc.CreateQRCodeParams) (*models.QRCode, error) {
	q, err := r.queries.CreateQRCode(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create QR code")
	}
	return models.QRCodeFromSqlc(q), nil
}

func (r *qrCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.QRCode, error) {
	q, err := r.queries.GetQRCodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("qr_code")
		}
		return nil, httputil.Wrap(err, "failed to get QR code")
	}
	return models.QRCodeFromSqlc(q), nil
}

func (r *qrCodeRepository) GetByLinkID(ctx context.Context, linkID uuid.UUID) (*models.QRCode, error) {
	q, err := r.queries.GetQRCodeByLinkID(ctx, linkID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("qr_code")
		}
		return nil, httputil.Wrap(err, "failed to get QR code for link")
	}
	return models.QRCodeFromSqlc(q), nil
}

func (r *qrCodeRepository) ListForLink(ctx context.Context, linkID uuid.UUID) ([]*models.QRCode, error) {
	rows, err := r.queries.ListQRCodesForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list QR codes")
	}

	qrs := make([]*models.QRCode, 0, len(rows))
	for _, row := range rows {
		qrs = append(qrs, models.QRCodeFromSqlc(row))
	}
	return qrs, nil
}

func (r *qrCodeRepository) Update(ctx context.Context, params sqlc.UpdateQRCodeParams) (*models.QRCode, error) {
	q, err := r.queries.UpdateQRCode(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("qr_code")
		}
		return nil, httputil.Wrap(err, "failed to update QR code")
	}
	return models.QRCodeFromSqlc(q), nil
}

func (r *qrCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteQRCode(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to delete QR code")
	}
	return nil
}

func (r *qrCodeRepository) IncrementScanCount(ctx context.Context, id uuid.UUID) error {
	err := r.queries.IncrementQRScanCount(ctx, id)
	if err != nil {
		return httputil.Wrap(err, "failed to increment scan count")
	}
	return nil
}
