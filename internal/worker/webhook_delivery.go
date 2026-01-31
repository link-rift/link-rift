package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	webhookDeliveryQueue  = "webhook:delivery:queue"
	maxWebhookAttempts    = 5
	maxFailuresPerDay     = 10
	retryPollInterval     = 30 * time.Second
	webhookRequestTimeout = 10 * time.Second
	maxResponseBodyLen    = 4096
)

// WebhookDeliveryProcessor processes webhook events from the Redis queue.
type WebhookDeliveryProcessor struct {
	redis       *redis.Client
	webhookRepo repository.WebhookRepository
	httpClient  *http.Client
	logger      *zap.Logger
	done        chan struct{}
}

func NewWebhookDeliveryProcessor(
	redisClient *redis.Client,
	webhookRepo repository.WebhookRepository,
	logger *zap.Logger,
) *WebhookDeliveryProcessor {
	return &WebhookDeliveryProcessor{
		redis:       redisClient,
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: webhookRequestTimeout,
		},
		logger: logger,
		done:   make(chan struct{}),
	}
}

// Start begins processing webhook delivery events.
func (p *WebhookDeliveryProcessor) Start(ctx context.Context) {
	p.logger.Info("webhook delivery processor started")

	// Start retry goroutine
	go p.retryLoop(ctx)

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("webhook delivery processor shutting down")
			return
		case <-p.done:
			return
		default:
			p.processQueue(ctx)
		}
	}
}

// Stop signals the processor to stop.
func (p *WebhookDeliveryProcessor) Stop() {
	close(p.done)
}

func (p *WebhookDeliveryProcessor) processQueue(ctx context.Context) {
	result, err := p.redis.BLPop(ctx, 2*time.Second, webhookDeliveryQueue).Result()
	if err != nil {
		if err == redis.Nil {
			return
		}
		if ctx.Err() != nil {
			return
		}
		p.logger.Error("failed to pop from webhook delivery queue", zap.Error(err))
		time.Sleep(1 * time.Second)
		return
	}

	var event models.WebhookEvent
	if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
		p.logger.Warn("failed to unmarshal webhook event", zap.Error(err))
		return
	}

	p.processEvent(ctx, &event)
}

func (p *WebhookDeliveryProcessor) processEvent(ctx context.Context, event *models.WebhookEvent) {
	// Find active webhooks for this event
	webhooks, err := p.webhookRepo.GetActiveForEvent(ctx, event.WorkspaceID, event.Event)
	if err != nil {
		p.logger.Error("failed to get active webhooks for event",
			zap.String("event", event.Event),
			zap.Error(err),
		)
		return
	}

	for _, webhook := range webhooks {
		// Build delivery payload
		payload, err := json.Marshal(map[string]any{
			"event":       event.Event,
			"workspace_id": event.WorkspaceID,
			"timestamp":   event.Timestamp,
			"data":        json.RawMessage(event.Data),
		})
		if err != nil {
			p.logger.Error("failed to marshal delivery payload", zap.Error(err))
			continue
		}

		// Create delivery record
		delivery, err := p.webhookRepo.CreateDelivery(ctx, sqlc.CreateWebhookDeliveryParams{
			WebhookID:   webhook.ID,
			Event:       event.Event,
			Payload:     payload,
			MaxAttempts: maxWebhookAttempts,
		})
		if err != nil {
			p.logger.Error("failed to create webhook delivery", zap.Error(err))
			continue
		}

		// Attempt delivery
		p.deliver(ctx, webhook, delivery, payload)
	}
}

func (p *WebhookDeliveryProcessor) deliver(ctx context.Context, webhook *models.Webhook, delivery *models.WebhookDelivery, payload []byte) {
	deliveryID := delivery.ID
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// HMAC-SHA256 signature
	signature := signPayload(webhook.Secret, payload, timestamp)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payload))
	if err != nil {
		p.logger.Error("failed to create webhook request", zap.Error(err))
		p.recordFailure(ctx, webhook.ID, deliveryID, 1, 0, "failed to create request: "+err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Linkrift-Signature", signature)
	req.Header.Set("X-Linkrift-Timestamp", timestamp)
	req.Header.Set("X-Linkrift-Event", delivery.Event)
	req.Header.Set("X-Linkrift-Delivery", deliveryID.String())
	req.Header.Set("User-Agent", "Linkrift-Webhooks/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Warn("webhook delivery failed",
			zap.String("webhook_id", webhook.ID.String()),
			zap.String("delivery_id", deliveryID.String()),
			zap.Error(err),
		)
		p.recordFailure(ctx, webhook.ID, deliveryID, 1, 0, "request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body (limited)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, int64(maxResponseBodyLen)))
	respBody := string(bodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		p.recordSuccess(ctx, webhook.ID, deliveryID, 1, int32(resp.StatusCode), respBody)
	} else {
		// Failure
		p.logger.Warn("webhook delivery received non-2xx response",
			zap.String("webhook_id", webhook.ID.String()),
			zap.Int("status", resp.StatusCode),
		)
		p.recordFailure(ctx, webhook.ID, deliveryID, 1, int32(resp.StatusCode), respBody)
	}
}

func (p *WebhookDeliveryProcessor) recordSuccess(ctx context.Context, webhookID, deliveryID uuid.UUID, attempts int32, statusCode int32, body string) {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
		ID:             deliveryID,
		ResponseStatus: pgtype.Int4{Int32: statusCode, Valid: true},
		ResponseBody:   pgtype.Text{String: body, Valid: body != ""},
		Attempts:       attempts,
		CompletedAt:    now,
	}); err != nil {
		p.logger.Error("failed to update webhook delivery", zap.Error(err))
	}

	if err := p.webhookRepo.UpdateLastTriggered(ctx, webhookID); err != nil {
		p.logger.Error("failed to update webhook last triggered", zap.Error(err))
	}
}

func (p *WebhookDeliveryProcessor) recordFailure(ctx context.Context, webhookID, deliveryID uuid.UUID, attempts int32, statusCode int32, body string) {
	respStatus := pgtype.Int4{}
	if statusCode > 0 {
		respStatus = pgtype.Int4{Int32: statusCode, Valid: true}
	}

	if err := p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
		ID:             deliveryID,
		ResponseStatus: respStatus,
		ResponseBody:   pgtype.Text{String: body, Valid: body != ""},
		Attempts:       attempts,
		CompletedAt:    pgtype.Timestamptz{}, // not completed yet if retries remain
	}); err != nil {
		p.logger.Error("failed to update webhook delivery", zap.Error(err))
	}

	if err := p.webhookRepo.IncrementFailureCount(ctx, webhookID); err != nil {
		p.logger.Error("failed to increment webhook failure count", zap.Error(err))
	}

	// Auto-disable after too many failures
	failCount, err := p.webhookRepo.CountRecentFailures(ctx, webhookID)
	if err == nil && failCount >= maxFailuresPerDay {
		p.logger.Warn("disabling webhook due to excessive failures",
			zap.String("webhook_id", webhookID.String()),
			zap.Int64("failure_count", failCount),
		)
		if err := p.webhookRepo.Disable(ctx, webhookID); err != nil {
			p.logger.Error("failed to disable webhook", zap.Error(err))
		}
	}
}

func (p *WebhookDeliveryProcessor) retryLoop(ctx context.Context) {
	ticker := time.NewTicker(retryPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.done:
			return
		case <-ticker.C:
			p.retryPendingDeliveries(ctx)
		}
	}
}

func (p *WebhookDeliveryProcessor) retryPendingDeliveries(ctx context.Context) {
	deliveries, err := p.webhookRepo.GetPendingDeliveries(ctx)
	if err != nil {
		p.logger.Error("failed to get pending webhook deliveries", zap.Error(err))
		return
	}

	for _, delivery := range deliveries {
		webhook, err := p.webhookRepo.GetByID(ctx, delivery.WebhookID)
		if err != nil {
			p.logger.Error("failed to get webhook for retry",
				zap.String("webhook_id", delivery.WebhookID.String()),
				zap.Error(err),
			)
			continue
		}

		if !webhook.IsActive {
			// Mark as completed (failed) if webhook is disabled
			now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
			p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
				ID:          delivery.ID,
				Attempts:    delivery.Attempts,
				CompletedAt: now,
				ResponseBody: pgtype.Text{String: "webhook disabled", Valid: true},
			})
			continue
		}

		p.retryDeliver(ctx, webhook, delivery)
	}
}

func (p *WebhookDeliveryProcessor) retryDeliver(ctx context.Context, webhook *models.Webhook, delivery *models.WebhookDelivery) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := signPayload(webhook.Secret, delivery.Payload, timestamp)
	attempts := delivery.Attempts + 1

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		p.recordFailure(ctx, webhook.ID, delivery.ID, attempts, 0, "failed to create request: "+err.Error())
		if attempts >= delivery.MaxAttempts {
			now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
			p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
				ID:          delivery.ID,
				Attempts:    attempts,
				CompletedAt: now,
			})
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Linkrift-Signature", signature)
	req.Header.Set("X-Linkrift-Timestamp", timestamp)
	req.Header.Set("X-Linkrift-Event", delivery.Event)
	req.Header.Set("X-Linkrift-Delivery", delivery.ID.String())
	req.Header.Set("User-Agent", "Linkrift-Webhooks/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		if attempts >= delivery.MaxAttempts {
			now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
			p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
				ID:             delivery.ID,
				ResponseBody:   pgtype.Text{String: "request failed: " + err.Error(), Valid: true},
				Attempts:       attempts,
				CompletedAt:    now,
			})
		} else {
			p.recordFailure(ctx, webhook.ID, delivery.ID, attempts, 0, "request failed: "+err.Error())
		}
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, int64(maxResponseBodyLen)))
	respBody := string(bodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		p.recordSuccess(ctx, webhook.ID, delivery.ID, attempts, int32(resp.StatusCode), respBody)
	} else {
		if attempts >= delivery.MaxAttempts {
			now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
			p.webhookRepo.UpdateDelivery(ctx, sqlc.UpdateWebhookDeliveryParams{
				ID:             delivery.ID,
				ResponseStatus: pgtype.Int4{Int32: int32(resp.StatusCode), Valid: true},
				ResponseBody:   pgtype.Text{String: respBody, Valid: respBody != ""},
				Attempts:       attempts,
				CompletedAt:    now,
			})
			p.webhookRepo.IncrementFailureCount(ctx, webhook.ID)
		} else {
			p.recordFailure(ctx, webhook.ID, delivery.ID, attempts, int32(resp.StatusCode), respBody)
		}
	}
}

func signPayload(secret string, payload []byte, timestamp string) string {
	message := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}
