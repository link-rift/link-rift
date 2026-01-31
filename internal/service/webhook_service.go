package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type WebhookService interface {
	CreateWebhook(ctx context.Context, workspaceID uuid.UUID, input models.CreateWebhookInput) (*models.CreateWebhookResponse, error)
	ListWebhooks(ctx context.Context, workspaceID uuid.UUID) ([]*models.Webhook, error)
	GetWebhook(ctx context.Context, id, workspaceID uuid.UUID) (*models.Webhook, error)
	DeleteWebhook(ctx context.Context, id, workspaceID uuid.UUID) error
	ListDeliveries(ctx context.Context, webhookID, workspaceID uuid.UUID, limit, offset int32) ([]*models.WebhookDelivery, int64, error)
}

type webhookService struct {
	webhookRepo repository.WebhookRepository
	licManager  *license.Manager
	logger      *zap.Logger
}

func NewWebhookService(
	webhookRepo repository.WebhookRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
		licManager:  licManager,
		logger:      logger,
	}
}

func (s *webhookService) CreateWebhook(ctx context.Context, workspaceID uuid.UUID, input models.CreateWebhookInput) (*models.CreateWebhookResponse, error) {
	if !s.licManager.HasFeature(license.FeatureWebhooks) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureWebhooks), "business")
	}

	// Validate URL is HTTPS
	if !strings.HasPrefix(input.URL, "https://") {
		return nil, httputil.Validation("url", "webhook URL must use HTTPS")
	}

	// Validate events
	for _, event := range input.Events {
		if !models.IsValidWebhookEvent(event) {
			return nil, httputil.Validation("events", fmt.Sprintf("invalid event: %s", event))
		}
	}

	// Generate secret: whsec_ + 32 random hex bytes
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, httputil.Wrap(err, "failed to generate webhook secret")
	}
	secret := "whsec_" + hex.EncodeToString(rawBytes)

	params := sqlc.CreateWebhookParams{
		WorkspaceID: workspaceID,
		Url:         input.URL,
		Secret:      secret,
		Events:      input.Events,
		IsActive:    true,
	}

	webhook, err := s.webhookRepo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &models.CreateWebhookResponse{
		Webhook: webhook,
		Secret:  secret,
	}, nil
}

func (s *webhookService) ListWebhooks(ctx context.Context, workspaceID uuid.UUID) ([]*models.Webhook, error) {
	return s.webhookRepo.List(ctx, workspaceID)
}

func (s *webhookService) GetWebhook(ctx context.Context, id, workspaceID uuid.UUID) (*models.Webhook, error) {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if webhook.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("webhook does not belong to this workspace")
	}
	return webhook, nil
}

func (s *webhookService) DeleteWebhook(ctx context.Context, id, workspaceID uuid.UUID) error {
	webhook, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if webhook.WorkspaceID != workspaceID {
		return httputil.Forbidden("webhook does not belong to this workspace")
	}
	return s.webhookRepo.Delete(ctx, id)
}

func (s *webhookService) ListDeliveries(ctx context.Context, webhookID, workspaceID uuid.UUID, limit, offset int32) ([]*models.WebhookDelivery, int64, error) {
	// Verify webhook belongs to workspace
	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, 0, err
	}
	if webhook.WorkspaceID != workspaceID {
		return nil, 0, httputil.Forbidden("webhook does not belong to this workspace")
	}

	deliveries, err := s.webhookRepo.ListDeliveries(ctx, webhookID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.webhookRepo.CountDeliveries(ctx, webhookID)
	if err != nil {
		return nil, 0, err
	}

	return deliveries, total, nil
}
