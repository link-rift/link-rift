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

type WebhookRepository interface {
	Create(ctx context.Context, params sqlc.CreateWebhookParams) (*models.Webhook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]*models.Webhook, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveForEvent(ctx context.Context, workspaceID uuid.UUID, event string) ([]*models.Webhook, error)
	IncrementFailureCount(ctx context.Context, id uuid.UUID) error
	UpdateLastTriggered(ctx context.Context, id uuid.UUID) error
	Disable(ctx context.Context, id uuid.UUID) error
	CreateDelivery(ctx context.Context, params sqlc.CreateWebhookDeliveryParams) (*models.WebhookDelivery, error)
	ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit, offset int32) ([]*models.WebhookDelivery, error)
	CountDeliveries(ctx context.Context, webhookID uuid.UUID) (int64, error)
	UpdateDelivery(ctx context.Context, params sqlc.UpdateWebhookDeliveryParams) error
	GetPendingDeliveries(ctx context.Context) ([]*models.WebhookDelivery, error)
	CountRecentFailures(ctx context.Context, webhookID uuid.UUID) (int64, error)
}

type webhookRepository struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewWebhookRepository(queries *sqlc.Queries, logger *zap.Logger) WebhookRepository {
	return &webhookRepository{queries: queries, logger: logger}
}

func (r *webhookRepository) Create(ctx context.Context, params sqlc.CreateWebhookParams) (*models.Webhook, error) {
	w, err := r.queries.CreateWebhook(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create webhook")
	}
	return models.WebhookFromSqlc(w), nil
}

func (r *webhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	w, err := r.queries.GetWebhookByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httputil.NotFound("webhook")
		}
		return nil, httputil.Wrap(err, "failed to get webhook")
	}
	return models.WebhookFromSqlc(w), nil
}

func (r *webhookRepository) List(ctx context.Context, workspaceID uuid.UUID) ([]*models.Webhook, error) {
	webhooks, err := r.queries.ListWebhooksForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list webhooks")
	}
	result := make([]*models.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		result = append(result, models.WebhookFromSqlc(w))
	}
	return result, nil
}

func (r *webhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DeleteWebhook(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to delete webhook")
	}
	return nil
}

func (r *webhookRepository) GetActiveForEvent(ctx context.Context, workspaceID uuid.UUID, event string) ([]*models.Webhook, error) {
	webhooks, err := r.queries.GetActiveWebhooksForEvent(ctx, sqlc.GetActiveWebhooksForEventParams{
		WorkspaceID: workspaceID,
		Event:       event,
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to get active webhooks for event")
	}
	result := make([]*models.Webhook, 0, len(webhooks))
	for _, w := range webhooks {
		result = append(result, models.WebhookFromSqlc(w))
	}
	return result, nil
}

func (r *webhookRepository) IncrementFailureCount(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.IncrementWebhookFailureCount(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to increment webhook failure count")
	}
	return nil
}

func (r *webhookRepository) UpdateLastTriggered(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.UpdateWebhookLastTriggered(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to update webhook last triggered")
	}
	return nil
}

func (r *webhookRepository) Disable(ctx context.Context, id uuid.UUID) error {
	if err := r.queries.DisableWebhook(ctx, id); err != nil {
		return httputil.Wrap(err, "failed to disable webhook")
	}
	return nil
}

func (r *webhookRepository) CreateDelivery(ctx context.Context, params sqlc.CreateWebhookDeliveryParams) (*models.WebhookDelivery, error) {
	d, err := r.queries.CreateWebhookDelivery(ctx, params)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create webhook delivery")
	}
	return models.WebhookDeliveryFromSqlc(d), nil
}

func (r *webhookRepository) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit, offset int32) ([]*models.WebhookDelivery, error) {
	deliveries, err := r.queries.ListWebhookDeliveries(ctx, sqlc.ListWebhookDeliveriesParams{
		WebhookID: webhookID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to list webhook deliveries")
	}
	result := make([]*models.WebhookDelivery, 0, len(deliveries))
	for _, d := range deliveries {
		result = append(result, models.WebhookDeliveryFromSqlc(d))
	}
	return result, nil
}

func (r *webhookRepository) CountDeliveries(ctx context.Context, webhookID uuid.UUID) (int64, error) {
	count, err := r.queries.CountWebhookDeliveries(ctx, webhookID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to count webhook deliveries")
	}
	return count, nil
}

func (r *webhookRepository) UpdateDelivery(ctx context.Context, params sqlc.UpdateWebhookDeliveryParams) error {
	if err := r.queries.UpdateWebhookDelivery(ctx, params); err != nil {
		return httputil.Wrap(err, "failed to update webhook delivery")
	}
	return nil
}

func (r *webhookRepository) GetPendingDeliveries(ctx context.Context) ([]*models.WebhookDelivery, error) {
	deliveries, err := r.queries.GetPendingWebhookDeliveries(ctx)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to get pending webhook deliveries")
	}
	result := make([]*models.WebhookDelivery, 0, len(deliveries))
	for _, d := range deliveries {
		result = append(result, models.WebhookDeliveryFromSqlc(d))
	}
	return result, nil
}

func (r *webhookRepository) CountRecentFailures(ctx context.Context, webhookID uuid.UUID) (int64, error) {
	count, err := r.queries.CountRecentWebhookFailures(ctx, webhookID)
	if err != nil {
		return 0, httputil.Wrap(err, "failed to count recent webhook failures")
	}
	return count, nil
}
