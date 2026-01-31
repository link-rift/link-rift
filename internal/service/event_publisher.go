package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const webhookDeliveryQueue = "webhook:delivery:queue"

// EventPublisher publishes webhook events to the delivery queue.
type EventPublisher interface {
	Publish(ctx context.Context, event string, workspaceID uuid.UUID, data any) error
}

type redisEventPublisher struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewEventPublisher creates a new Redis-backed event publisher.
func NewEventPublisher(redisClient *redis.Client, logger *zap.Logger) EventPublisher {
	return &redisEventPublisher{
		redis:  redisClient,
		logger: logger,
	}
}

func (p *redisEventPublisher) Publish(ctx context.Context, event string, workspaceID uuid.UUID, data any) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		p.logger.Warn("failed to marshal webhook event data",
			zap.String("event", event),
			zap.Error(err),
		)
		return err
	}

	webhookEvent := models.WebhookEvent{
		Event:       event,
		WorkspaceID: workspaceID,
		Timestamp:   time.Now().UTC(),
		Data:        dataJSON,
	}

	eventJSON, err := json.Marshal(webhookEvent)
	if err != nil {
		p.logger.Warn("failed to marshal webhook event",
			zap.String("event", event),
			zap.Error(err),
		)
		return err
	}

	if err := p.redis.RPush(ctx, webhookDeliveryQueue, eventJSON).Err(); err != nil {
		p.logger.Warn("failed to publish webhook event",
			zap.String("event", event),
			zap.Error(err),
		)
		return err
	}

	p.logger.Debug("published webhook event",
		zap.String("event", event),
		zap.String("workspace_id", workspaceID.String()),
	)

	return nil
}

// noopEventPublisher is a no-op publisher for when webhooks are not configured.
type noopEventPublisher struct{}

// NewNoopEventPublisher returns a publisher that does nothing.
func NewNoopEventPublisher() EventPublisher {
	return &noopEventPublisher{}
}

func (p *noopEventPublisher) Publish(_ context.Context, _ string, _ uuid.UUID, _ any) error {
	return nil
}
