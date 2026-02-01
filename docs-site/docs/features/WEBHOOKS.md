# Webhooks

> Last Updated: 2025-01-24

Linkrift provides a robust webhook system that enables real-time notifications for various events, allowing integrations with external systems and automations.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Supported Events](#supported-events)
- [Payload Formats](#payload-formats)
- [Retry with Exponential Backoff](#retry-with-exponential-backoff)
- [HMAC Signature Verification](#hmac-signature-verification)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

The webhook system provides:

- **Real-time event notifications** for link clicks, creation, and other events
- **Reliable delivery** with exponential backoff retry logic
- **Security** through HMAC-SHA256 signature verification
- **Flexible filtering** by event types
- **Detailed logging** for debugging and monitoring

## Architecture

```
                                    ┌─────────────────────────────────────────┐
                                    │         Event Source Services           │
                                    │   (Redirect, Links, Analytics, etc.)    │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       │ Publish Events
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │           Event Bus (Redis)              │
                                    │         webhook.events channel          │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       │ Subscribe
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │          Webhook Dispatcher              │
                                    │    (Event Processing & Delivery)         │
                                    └──────────────────┬──────────────────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        │                        │
                              ▼                        ▼                        ▼
               ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
               │   Webhook Worker     │  │   Webhook Worker     │  │    Retry Queue       │
               │   (HTTP Delivery)    │  │   (HTTP Delivery)    │  │   (Failed Events)    │
               └──────────────────────┘  └──────────────────────┘  └──────────────────────┘
                              │                        │                        │
                              └────────────────────────┼────────────────────────┘
                                                       │
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │         Customer Endpoints               │
                                    │     (https://customer.com/webhook)       │
                                    └─────────────────────────────────────────┘
```

---

## Supported Events

```go
// internal/webhooks/events.go
package webhooks

// EventType defines webhook event types
type EventType string

const (
	// Link events
	EventLinkCreated     EventType = "link.created"
	EventLinkUpdated     EventType = "link.updated"
	EventLinkDeleted     EventType = "link.deleted"
	EventLinkClicked     EventType = "link.clicked"
	EventLinkExpired     EventType = "link.expired"

	// QR Code events
	EventQRCreated       EventType = "qr.created"
	EventQRScanned       EventType = "qr.scanned"

	// Bio Page events
	EventBioPageCreated  EventType = "biopage.created"
	EventBioPageUpdated  EventType = "biopage.updated"
	EventBioPageViewed   EventType = "biopage.viewed"

	// Domain events
	EventDomainAdded     EventType = "domain.added"
	EventDomainVerified  EventType = "domain.verified"
	EventDomainRemoved   EventType = "domain.removed"
	EventDomainSSLIssued EventType = "domain.ssl_issued"

	// Team events
	EventMemberInvited   EventType = "team.member_invited"
	EventMemberJoined    EventType = "team.member_joined"
	EventMemberRemoved   EventType = "team.member_removed"
	EventMemberRoleChanged EventType = "team.member_role_changed"

	// Billing events
	EventSubscriptionCreated  EventType = "subscription.created"
	EventSubscriptionUpdated  EventType = "subscription.updated"
	EventSubscriptionCanceled EventType = "subscription.canceled"
	EventPaymentSucceeded     EventType = "payment.succeeded"
	EventPaymentFailed        EventType = "payment.failed"
	EventUsageLimitReached    EventType = "usage.limit_reached"
)

// EventCategories groups events by category
var EventCategories = map[string][]EventType{
	"links": {
		EventLinkCreated,
		EventLinkUpdated,
		EventLinkDeleted,
		EventLinkClicked,
		EventLinkExpired,
	},
	"qr": {
		EventQRCreated,
		EventQRScanned,
	},
	"biopages": {
		EventBioPageCreated,
		EventBioPageUpdated,
		EventBioPageViewed,
	},
	"domains": {
		EventDomainAdded,
		EventDomainVerified,
		EventDomainRemoved,
		EventDomainSSLIssued,
	},
	"team": {
		EventMemberInvited,
		EventMemberJoined,
		EventMemberRemoved,
		EventMemberRoleChanged,
	},
	"billing": {
		EventSubscriptionCreated,
		EventSubscriptionUpdated,
		EventSubscriptionCanceled,
		EventPaymentSucceeded,
		EventPaymentFailed,
		EventUsageLimitReached,
	},
}

// AllEvents returns all available event types
func AllEvents() []EventType {
	var events []EventType
	for _, categoryEvents := range EventCategories {
		events = append(events, categoryEvents...)
	}
	return events
}
```

---

## Payload Formats

```go
// internal/webhooks/payload.go
package webhooks

import (
	"encoding/json"
	"time"
)

// WebhookPayload represents the standard webhook payload structure
type WebhookPayload struct {
	ID          string          `json:"id"`           // Unique event ID
	Type        EventType       `json:"type"`         // Event type
	WorkspaceID string          `json:"workspace_id"` // Workspace that owns the resource
	CreatedAt   time.Time       `json:"created_at"`   // Event timestamp
	Data        json.RawMessage `json:"data"`         // Event-specific data
}

// LinkCreatedPayload is sent when a link is created
type LinkCreatedPayload struct {
	LinkID      string    `json:"link_id"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	Title       string    `json:"title,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// LinkClickedPayload is sent when a link is clicked
type LinkClickedPayload struct {
	LinkID      string    `json:"link_id"`
	ShortCode   string    `json:"short_code"`
	ClickID     string    `json:"click_id"`
	Timestamp   time.Time `json:"timestamp"`

	// Visitor information
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	City        string `json:"city,omitempty"`
	Region      string `json:"region,omitempty"`

	// Device information
	DeviceType  string `json:"device_type,omitempty"`
	Browser     string `json:"browser,omitempty"`
	OS          string `json:"os,omitempty"`

	// Traffic source
	Referer     string `json:"referer,omitempty"`
	RefererHost string `json:"referer_host,omitempty"`

	// UTM parameters
	UTMSource   string `json:"utm_source,omitempty"`
	UTMMedium   string `json:"utm_medium,omitempty"`
	UTMCampaign string `json:"utm_campaign,omitempty"`

	// Classification
	IsBot       bool   `json:"is_bot"`
	IsUnique    bool   `json:"is_unique"`
}

// LinkUpdatedPayload is sent when a link is updated
type LinkUpdatedPayload struct {
	LinkID        string                 `json:"link_id"`
	ShortCode     string                 `json:"short_code"`
	Changes       map[string]ChangeValue `json:"changes"`
	UpdatedBy     string                 `json:"updated_by"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ChangeValue represents a before/after change
type ChangeValue struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// DomainVerifiedPayload is sent when a domain is verified
type DomainVerifiedPayload struct {
	DomainID   string    `json:"domain_id"`
	Domain     string    `json:"domain"`
	VerifiedAt time.Time `json:"verified_at"`
}

// TeamMemberJoinedPayload is sent when a team member joins
type TeamMemberJoinedPayload struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	InvitedBy string    `json:"invited_by"`
	JoinedAt  time.Time `json:"joined_at"`
}

// UsageLimitReachedPayload is sent when usage limits are reached
type UsageLimitReachedPayload struct {
	LimitType   string  `json:"limit_type"` // clicks, links, etc.
	Limit       int64   `json:"limit"`
	CurrentUsage int64  `json:"current_usage"`
	Percentage  float64 `json:"percentage"`
	Period      string  `json:"period"`
}

// Example payloads for documentation
var ExamplePayloads = map[EventType]interface{}{
	EventLinkClicked: WebhookPayload{
		ID:          "evt_123456789",
		Type:        EventLinkClicked,
		WorkspaceID: "ws_abc123",
		CreatedAt:   time.Now(),
		Data: json.RawMessage(`{
			"link_id": "lnk_xyz789",
			"short_code": "abc123",
			"click_id": "clk_456def",
			"timestamp": "2025-01-24T12:00:00Z",
			"country": "United States",
			"country_code": "US",
			"city": "San Francisco",
			"device_type": "mobile",
			"browser": "Chrome",
			"os": "iOS",
			"referer_host": "twitter.com",
			"is_bot": false,
			"is_unique": true
		}`),
	},
}
```

---

## Retry with Exponential Backoff

```go
// internal/webhooks/dispatcher.go
package webhooks

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

	"github.com/redis/go-redis/v9"
	"github.com/link-rift/link-rift/internal/db"
)

// Webhook represents a registered webhook endpoint
type Webhook struct {
	ID           string      `json:"id" db:"id"`
	WorkspaceID  string      `json:"workspace_id" db:"workspace_id"`
	URL          string      `json:"url" db:"url"`
	Secret       string      `json:"-" db:"secret"`
	Events       []EventType `json:"events" db:"events"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	Description  string      `json:"description,omitempty" db:"description"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// WebhookDelivery represents a delivery attempt
type WebhookDelivery struct {
	ID           string        `json:"id" db:"id"`
	WebhookID    string        `json:"webhook_id" db:"webhook_id"`
	EventID      string        `json:"event_id" db:"event_id"`
	EventType    EventType     `json:"event_type" db:"event_type"`
	Payload      string        `json:"payload" db:"payload"`
	StatusCode   int           `json:"status_code" db:"status_code"`
	Response     string        `json:"response,omitempty" db:"response"`
	Duration     time.Duration `json:"duration_ms" db:"duration_ms"`
	Attempts     int           `json:"attempts" db:"attempts"`
	NextRetryAt  *time.Time    `json:"next_retry_at,omitempty" db:"next_retry_at"`
	DeliveredAt  *time.Time    `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt     *time.Time    `json:"failed_at,omitempty" db:"failed_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffMultiplier float64     `json:"backoff_multiplier"`
}

// DefaultRetryConfig provides sensible defaults
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:       5,
	InitialDelay:      30 * time.Second,
	MaxDelay:          24 * time.Hour,
	BackoffMultiplier: 2.0,
}

// Dispatcher handles webhook delivery
type Dispatcher struct {
	repo         *db.WebhookRepository
	deliveryRepo *db.DeliveryRepository
	redis        *redis.Client
	httpClient   *http.Client
	retryConfig  RetryConfig
	workers      int
	quit         chan struct{}
}

// NewDispatcher creates a new webhook dispatcher
func NewDispatcher(
	repo *db.WebhookRepository,
	deliveryRepo *db.DeliveryRepository,
	redis *redis.Client,
	retryConfig RetryConfig,
	workers int,
) *Dispatcher {
	return &Dispatcher{
		repo:         repo,
		deliveryRepo: deliveryRepo,
		redis:        redis,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		retryConfig: retryConfig,
		workers:     workers,
		quit:        make(chan struct{}),
	}
}

// Start begins processing webhook events
func (d *Dispatcher) Start(ctx context.Context) error {
	// Subscribe to event channel
	pubsub := d.redis.Subscribe(ctx, "webhook.events")

	// Start workers
	for i := 0; i < d.workers; i++ {
		go d.worker(ctx)
	}

	// Start retry processor
	go d.retryProcessor(ctx)

	// Process incoming events
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			var event WebhookPayload
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}
			go d.processEvent(ctx, &event)

		case <-d.quit:
			pubsub.Close()
			return nil
		}
	}
}

func (d *Dispatcher) processEvent(ctx context.Context, event *WebhookPayload) {
	// Find webhooks subscribed to this event
	webhooks, err := d.repo.GetByWorkspaceAndEvent(ctx, event.WorkspaceID, event.Type)
	if err != nil {
		return
	}

	for _, webhook := range webhooks {
		if !webhook.IsActive {
			continue
		}

		// Create delivery record
		delivery := &WebhookDelivery{
			WebhookID:  webhook.ID,
			EventID:    event.ID,
			EventType:  event.Type,
			Payload:    string(mustMarshal(event)),
			Attempts:   0,
			CreatedAt:  time.Now(),
		}

		if err := d.deliveryRepo.Create(ctx, delivery); err != nil {
			continue
		}

		// Attempt delivery
		d.deliver(ctx, webhook, delivery)
	}
}

func (d *Dispatcher) deliver(ctx context.Context, webhook *Webhook, delivery *WebhookDelivery) {
	delivery.Attempts++

	// Prepare request
	payload := []byte(delivery.Payload)
	timestamp := time.Now().Unix()
	signature := d.signPayload(webhook.Secret, payload, timestamp)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		d.handleDeliveryFailure(ctx, webhook, delivery, 0, err.Error())
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Linkrift-Webhook/1.0")
	req.Header.Set("X-Linkrift-Event", string(delivery.EventType))
	req.Header.Set("X-Linkrift-Delivery", delivery.ID)
	req.Header.Set("X-Linkrift-Signature", signature)
	req.Header.Set("X-Linkrift-Timestamp", fmt.Sprintf("%d", timestamp))

	// Execute request
	start := time.Now()
	resp, err := d.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		d.handleDeliveryFailure(ctx, webhook, delivery, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*10)) // Max 10KB

	delivery.StatusCode = resp.StatusCode
	delivery.Response = string(body)
	delivery.Duration = duration

	// Check for success (2xx status codes)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		now := time.Now()
		delivery.DeliveredAt = &now
		delivery.NextRetryAt = nil
		d.deliveryRepo.Update(ctx, delivery)
		return
	}

	// Handle failure
	d.handleDeliveryFailure(ctx, webhook, delivery, resp.StatusCode, string(body))
}

func (d *Dispatcher) handleDeliveryFailure(
	ctx context.Context,
	webhook *Webhook,
	delivery *WebhookDelivery,
	statusCode int,
	errorMsg string,
) {
	delivery.StatusCode = statusCode
	delivery.Response = errorMsg

	if delivery.Attempts >= d.retryConfig.MaxAttempts {
		// Max attempts reached, mark as failed
		now := time.Now()
		delivery.FailedAt = &now
		delivery.NextRetryAt = nil
		d.deliveryRepo.Update(ctx, delivery)

		// Optionally disable webhook after too many failures
		d.checkWebhookHealth(ctx, webhook)
		return
	}

	// Calculate next retry time with exponential backoff
	delay := d.calculateBackoff(delivery.Attempts)
	nextRetry := time.Now().Add(delay)
	delivery.NextRetryAt = &nextRetry
	d.deliveryRepo.Update(ctx, delivery)

	// Add to retry queue
	d.scheduleRetry(ctx, delivery, delay)
}

func (d *Dispatcher) calculateBackoff(attempt int) time.Duration {
	delay := float64(d.retryConfig.InitialDelay)
	for i := 1; i < attempt; i++ {
		delay *= d.retryConfig.BackoffMultiplier
	}

	if delay > float64(d.retryConfig.MaxDelay) {
		delay = float64(d.retryConfig.MaxDelay)
	}

	return time.Duration(delay)
}

func (d *Dispatcher) scheduleRetry(ctx context.Context, delivery *WebhookDelivery, delay time.Duration) {
	// Use Redis sorted set for scheduling
	score := float64(time.Now().Add(delay).Unix())
	d.redis.ZAdd(ctx, "webhook:retry", redis.Z{
		Score:  score,
		Member: delivery.ID,
	})
}

func (d *Dispatcher) retryProcessor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.processRetries(ctx)
		case <-d.quit:
			return
		}
	}
}

func (d *Dispatcher) processRetries(ctx context.Context) {
	now := float64(time.Now().Unix())

	// Get deliveries due for retry
	deliveryIDs, err := d.redis.ZRangeByScore(ctx, "webhook:retry", &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil || len(deliveryIDs) == 0 {
		return
	}

	for _, deliveryID := range deliveryIDs {
		// Remove from retry queue
		d.redis.ZRem(ctx, "webhook:retry", deliveryID)

		// Get delivery and webhook
		delivery, err := d.deliveryRepo.GetByID(ctx, deliveryID)
		if err != nil {
			continue
		}

		webhook, err := d.repo.GetByID(ctx, delivery.WebhookID)
		if err != nil || !webhook.IsActive {
			continue
		}

		// Retry delivery
		d.deliver(ctx, webhook, delivery)
	}
}

func (d *Dispatcher) checkWebhookHealth(ctx context.Context, webhook *Webhook) {
	// Count recent failures
	failures, _ := d.deliveryRepo.CountRecentFailures(ctx, webhook.ID, 24*time.Hour)

	// Disable webhook if too many failures
	if failures >= 10 {
		webhook.IsActive = false
		d.repo.Update(ctx, webhook)

		// TODO: Notify workspace owner
	}
}

func (d *Dispatcher) worker(ctx context.Context) {
	// Worker pool for parallel delivery
	// Implementation depends on your concurrency model
}

// Stop gracefully shuts down the dispatcher
func (d *Dispatcher) Stop() {
	close(d.quit)
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
```

---

## HMAC Signature Verification

```go
// internal/webhooks/signature.go
package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	SignatureVersion   = "v1"
	SignatureHeader    = "X-Linkrift-Signature"
	TimestampHeader    = "X-Linkrift-Timestamp"
	TimestampTolerance = 5 * time.Minute // Reject events older than 5 minutes
)

// signPayload creates an HMAC-SHA256 signature
func (d *Dispatcher) signPayload(secret string, payload []byte, timestamp int64) string {
	// Create signed payload: timestamp.payload
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))

	// Generate HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Return versioned signature
	return fmt.Sprintf("%s=%s", SignatureVersion, signature)
}

// SignatureVerifier helps verify webhook signatures
type SignatureVerifier struct {
	secret string
}

// NewSignatureVerifier creates a new signature verifier
func NewSignatureVerifier(secret string) *SignatureVerifier {
	return &SignatureVerifier{secret: secret}
}

// Verify validates a webhook signature
func (sv *SignatureVerifier) Verify(payload []byte, signature string, timestampStr string) error {
	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	// Check timestamp freshness
	eventTime := time.Unix(timestamp, 0)
	if time.Since(eventTime) > TimestampTolerance {
		return fmt.Errorf("timestamp too old")
	}

	// Parse signature version
	if len(signature) < 4 || signature[:3] != "v1=" {
		return fmt.Errorf("invalid signature format")
	}
	providedSig := signature[3:]

	// Generate expected signature
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(sv.secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison
	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// GenerateSecret generates a secure webhook secret
func GenerateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(bytes), nil
}
```

### Client-Side Verification Examples

```go
// Example: Go client webhook verification
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Get signature headers
	signature := r.Header.Get("X-Linkrift-Signature")
	timestampStr := r.Header.Get("X-Linkrift-Timestamp")

	// Verify signature
	if err := verifySignature(body, signature, timestampStr, "whsec_your_secret"); err != nil {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Process webhook
	// ...

	w.WriteHeader(http.StatusOK)
}

func verifySignature(payload []byte, signature, timestampStr, secret string) error {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return err
	}

	// Check timestamp (reject if older than 5 minutes)
	if time.Since(time.Unix(timestamp, 0)) > 5*time.Minute {
		return fmt.Errorf("timestamp too old")
	}

	// Parse signature
	if len(signature) < 4 || signature[:3] != "v1=" {
		return fmt.Errorf("invalid signature format")
	}
	providedSig := signature[3:]

	// Compute expected signature
	signedPayload := fmt.Sprintf("%d.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
```

```typescript
// Example: Node.js/TypeScript webhook verification
import crypto from 'crypto';
import express from 'express';

const app = express();

app.post('/webhook', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['x-linkrift-signature'] as string;
  const timestamp = req.headers['x-linkrift-timestamp'] as string;
  const secret = process.env.WEBHOOK_SECRET!;

  try {
    verifySignature(req.body, signature, timestamp, secret);
  } catch (err) {
    console.error('Webhook verification failed:', err);
    return res.status(401).send('Invalid signature');
  }

  const event = JSON.parse(req.body.toString());

  // Process event
  switch (event.type) {
    case 'link.clicked':
      handleLinkClicked(event.data);
      break;
    // ... handle other events
  }

  res.status(200).send('OK');
});

function verifySignature(
  payload: Buffer,
  signature: string,
  timestampStr: string,
  secret: string
): void {
  const timestamp = parseInt(timestampStr, 10);

  // Check timestamp (5 minute tolerance)
  const now = Math.floor(Date.now() / 1000);
  if (now - timestamp > 300) {
    throw new Error('Timestamp too old');
  }

  // Parse signature
  if (!signature.startsWith('v1=')) {
    throw new Error('Invalid signature format');
  }
  const providedSig = signature.slice(3);

  // Compute expected signature
  const signedPayload = `${timestamp}.${payload.toString()}`;
  const expectedSig = crypto
    .createHmac('sha256', secret)
    .update(signedPayload)
    .digest('hex');

  if (!crypto.timingSafeEqual(
    Buffer.from(providedSig),
    Buffer.from(expectedSig)
  )) {
    throw new Error('Signature mismatch');
  }
}
```

---

## API Endpoints

```go
// internal/api/handlers/webhooks.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/webhooks"
)

// WebhookHandler handles webhook API requests
type WebhookHandler struct {
	service *webhooks.Service
}

// RegisterRoutes registers webhook routes
func (h *WebhookHandler) RegisterRoutes(app *fiber.App) {
	wh := app.Group("/api/v1/webhooks")

	wh.Get("/", h.ListWebhooks)
	wh.Post("/", h.CreateWebhook)
	wh.Get("/:id", h.GetWebhook)
	wh.Put("/:id", h.UpdateWebhook)
	wh.Delete("/:id", h.DeleteWebhook)
	wh.Post("/:id/test", h.TestWebhook)
	wh.Post("/:id/rotate-secret", h.RotateSecret)
	wh.Get("/:id/deliveries", h.ListDeliveries)
	wh.Get("/:id/deliveries/:deliveryId", h.GetDelivery)
	wh.Post("/:id/deliveries/:deliveryId/retry", h.RetryDelivery)

	// Events reference
	wh.Get("/events", h.ListEventTypes)
}

// CreateWebhookRequest represents a webhook creation request
type CreateWebhookRequest struct {
	URL         string              `json:"url" validate:"required,url"`
	Events      []webhooks.EventType `json:"events" validate:"required,min=1"`
	Description string              `json:"description"`
}

// CreateWebhook creates a new webhook
// @Summary Create a webhook
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param body body CreateWebhookRequest true "Webhook details"
// @Success 201 {object} WebhookResponse
// @Router /api/v1/webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *fiber.Ctx) error {
	var req CreateWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	// Generate secret
	secret, err := webhooks.GenerateSecret()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate secret",
		})
	}

	webhook, err := h.service.Create(c.Context(), &webhooks.Webhook{
		WorkspaceID: workspaceID,
		URL:         req.URL,
		Secret:      secret,
		Events:      req.Events,
		Description: req.Description,
		IsActive:    true,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return webhook with secret (only shown once)
	return c.Status(fiber.StatusCreated).JSON(WebhookResponse{
		Webhook: webhook,
		Secret:  secret, // Only returned on creation
	})
}

// TestWebhook sends a test event to the webhook
// @Summary Test a webhook
// @Tags Webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} TestResult
// @Router /api/v1/webhooks/{id}/test [post]
func (h *WebhookHandler) TestWebhook(c *fiber.Ctx) error {
	webhookID := c.Params("id")

	result, err := h.service.SendTestEvent(c.Context(), webhookID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

// RotateSecret generates a new secret for a webhook
// @Summary Rotate webhook secret
// @Tags Webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} RotateSecretResponse
// @Router /api/v1/webhooks/{id}/rotate-secret [post]
func (h *WebhookHandler) RotateSecret(c *fiber.Ctx) error {
	webhookID := c.Params("id")

	newSecret, err := h.service.RotateSecret(c.Context(), webhookID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"secret": newSecret,
	})
}

// ListDeliveries returns webhook delivery history
// @Summary List webhook deliveries
// @Tags Webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} WebhookDelivery
// @Router /api/v1/webhooks/{id}/deliveries [get]
func (h *WebhookHandler) ListDeliveries(c *fiber.Ctx) error {
	webhookID := c.Params("id")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	deliveries, total, err := h.service.ListDeliveries(c.Context(), webhookID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// RetryDelivery retries a failed delivery
// @Summary Retry a failed delivery
// @Tags Webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Param deliveryId path string true "Delivery ID"
// @Success 200 {object} WebhookDelivery
// @Router /api/v1/webhooks/{id}/deliveries/{deliveryId}/retry [post]
func (h *WebhookHandler) RetryDelivery(c *fiber.Ctx) error {
	deliveryID := c.Params("deliveryId")

	delivery, err := h.service.RetryDelivery(c.Context(), deliveryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(delivery)
}

// ListEventTypes returns all available event types
// @Summary List event types
// @Tags Webhooks
// @Produce json
// @Success 200 {object} map[string][]EventType
// @Router /api/v1/webhooks/events [get]
func (h *WebhookHandler) ListEventTypes(c *fiber.Ctx) error {
	return c.JSON(webhooks.EventCategories)
}

// Response types
type WebhookResponse struct {
	*webhooks.Webhook
	Secret string `json:"secret,omitempty"` // Only on creation
}

type TestResult struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Duration   int64  `json:"duration_ms"`
	Response   string `json:"response,omitempty"`
	Error      string `json:"error,omitempty"`
}
```

---

## React Components

### Webhook Management

```typescript
// src/components/webhooks/WebhookManager.tsx
import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { webhooksApi, Webhook, EventType, WebhookDelivery } from '@/api/webhooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { AlertCircle, CheckCircle, Clock, RefreshCw, Eye, Copy } from 'lucide-react';

export const WebhookManager: React.FC = () => {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const { data: webhooks, isLoading } = useQuery({
    queryKey: ['webhooks'],
    queryFn: webhooksApi.list,
  });

  if (isLoading) return <div>Loading...</div>;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Webhooks</h2>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button>Create Webhook</Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Create Webhook</DialogTitle>
            </DialogHeader>
            <CreateWebhookForm onSuccess={() => {
              setIsCreateOpen(false);
              queryClient.invalidateQueries({ queryKey: ['webhooks'] });
            }} />
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-4">
        {webhooks?.map((webhook) => (
          <WebhookCard key={webhook.id} webhook={webhook} />
        ))}
      </div>

      {webhooks?.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No webhooks configured. Create one to start receiving events.
        </div>
      )}
    </div>
  );
};

// Webhook Card Component
interface WebhookCardProps {
  webhook: Webhook;
}

const WebhookCard: React.FC<WebhookCardProps> = ({ webhook }) => {
  const queryClient = useQueryClient();
  const [showDeliveries, setShowDeliveries] = useState(false);

  const toggleMutation = useMutation({
    mutationFn: () => webhooksApi.update(webhook.id, { is_active: !webhook.is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['webhooks'] }),
  });

  const testMutation = useMutation({
    mutationFn: () => webhooksApi.test(webhook.id),
  });

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <code className="text-sm bg-muted px-2 py-1 rounded">
                {webhook.url}
              </code>
              {webhook.is_active ? (
                <Badge variant="success">Active</Badge>
              ) : (
                <Badge variant="secondary">Inactive</Badge>
              )}
            </div>
            <p className="text-sm text-muted-foreground">
              {webhook.description || 'No description'}
            </p>
            <div className="flex flex-wrap gap-1">
              {webhook.events.map((event) => (
                <Badge key={event} variant="outline" className="text-xs">
                  {event}
                </Badge>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Switch
              checked={webhook.is_active}
              onCheckedChange={() => toggleMutation.mutate()}
              disabled={toggleMutation.isPending}
            />
            <Button
              variant="outline"
              size="sm"
              onClick={() => testMutation.mutate()}
              disabled={testMutation.isPending}
            >
              {testMutation.isPending ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                'Test'
              )}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowDeliveries(!showDeliveries)}
            >
              <Eye className="w-4 h-4 mr-1" />
              Deliveries
            </Button>
          </div>
        </div>

        {testMutation.data && (
          <div className={`mt-4 p-3 rounded-lg ${
            testMutation.data.success ? 'bg-green-50' : 'bg-red-50'
          }`}>
            {testMutation.data.success ? (
              <div className="flex items-center gap-2 text-green-700">
                <CheckCircle className="w-4 h-4" />
                <span>Test successful ({testMutation.data.duration_ms}ms)</span>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-red-700">
                <AlertCircle className="w-4 h-4" />
                <span>Test failed: {testMutation.data.error}</span>
              </div>
            )}
          </div>
        )}

        {showDeliveries && (
          <div className="mt-4 border-t pt-4">
            <DeliveryHistory webhookId={webhook.id} />
          </div>
        )}
      </CardContent>
    </Card>
  );
};

// Create Webhook Form
interface CreateWebhookFormProps {
  onSuccess: () => void;
}

const CreateWebhookForm: React.FC<CreateWebhookFormProps> = ({ onSuccess }) => {
  const [url, setUrl] = useState('');
  const [description, setDescription] = useState('');
  const [selectedEvents, setSelectedEvents] = useState<EventType[]>([]);
  const [newSecret, setNewSecret] = useState<string | null>(null);

  const { data: eventCategories } = useQuery({
    queryKey: ['webhook-events'],
    queryFn: webhooksApi.getEventTypes,
  });

  const createMutation = useMutation({
    mutationFn: webhooksApi.create,
    onSuccess: (data) => {
      setNewSecret(data.secret);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate({ url, description, events: selectedEvents });
  };

  const toggleEvent = (event: EventType) => {
    setSelectedEvents((prev) =>
      prev.includes(event)
        ? prev.filter((e) => e !== event)
        : [...prev, event]
    );
  };

  if (newSecret) {
    return (
      <div className="space-y-4">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="font-medium text-yellow-800 mb-2">
            Save Your Webhook Secret
          </h3>
          <p className="text-sm text-yellow-700 mb-4">
            This secret will only be shown once. Save it securely - you will need
            it to verify webhook signatures.
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-white px-3 py-2 rounded border font-mono text-sm">
              {newSecret}
            </code>
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigator.clipboard.writeText(newSecret)}
            >
              <Copy className="w-4 h-4" />
            </Button>
          </div>
        </div>
        <Button onClick={onSuccess} className="w-full">
          Done
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium">Endpoint URL</label>
        <Input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://your-app.com/webhook"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Description (optional)</label>
        <Input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="E.g., Production webhook for click tracking"
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Events</label>
        <div className="border rounded-lg p-4 space-y-4 max-h-64 overflow-auto">
          {eventCategories && Object.entries(eventCategories).map(([category, events]) => (
            <div key={category}>
              <h4 className="text-sm font-medium capitalize mb-2">{category}</h4>
              <div className="flex flex-wrap gap-2">
                {events.map((event) => (
                  <button
                    key={event}
                    type="button"
                    onClick={() => toggleEvent(event)}
                    className={`px-2 py-1 text-xs rounded-full border transition-colors ${
                      selectedEvents.includes(event)
                        ? 'bg-primary text-primary-foreground border-primary'
                        : 'bg-background hover:bg-muted'
                    }`}
                  >
                    {event}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      <Button
        type="submit"
        className="w-full"
        disabled={createMutation.isPending || selectedEvents.length === 0}
      >
        {createMutation.isPending ? 'Creating...' : 'Create Webhook'}
      </Button>
    </form>
  );
};

// Delivery History Component
interface DeliveryHistoryProps {
  webhookId: string;
}

const DeliveryHistory: React.FC<DeliveryHistoryProps> = ({ webhookId }) => {
  const { data, isLoading } = useQuery({
    queryKey: ['webhook-deliveries', webhookId],
    queryFn: () => webhooksApi.listDeliveries(webhookId),
  });

  const retryMutation = useMutation({
    mutationFn: (deliveryId: string) => webhooksApi.retryDelivery(webhookId, deliveryId),
  });

  if (isLoading) return <div>Loading deliveries...</div>;

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-medium">Recent Deliveries</h4>
      {data?.deliveries.map((delivery: WebhookDelivery) => (
        <div
          key={delivery.id}
          className="flex items-center justify-between p-2 bg-muted rounded text-sm"
        >
          <div className="flex items-center gap-2">
            {delivery.delivered_at ? (
              <CheckCircle className="w-4 h-4 text-green-500" />
            ) : delivery.failed_at ? (
              <AlertCircle className="w-4 h-4 text-red-500" />
            ) : (
              <Clock className="w-4 h-4 text-yellow-500" />
            )}
            <span>{delivery.event_type}</span>
            <span className="text-muted-foreground">
              {delivery.status_code || 'Pending'}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground">
              {new Date(delivery.created_at).toLocaleString()}
            </span>
            {delivery.failed_at && !delivery.delivered_at && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => retryMutation.mutate(delivery.id)}
                disabled={retryMutation.isPending}
              >
                Retry
              </Button>
            )}
          </div>
        </div>
      ))}
    </div>
  );
};
```

### Webhook API Client

```typescript
// src/api/webhooks.ts
import { apiClient } from './client';

export type EventType =
  | 'link.created'
  | 'link.updated'
  | 'link.deleted'
  | 'link.clicked'
  | 'link.expired'
  | 'qr.created'
  | 'qr.scanned'
  | 'biopage.created'
  | 'biopage.updated'
  | 'biopage.viewed'
  | 'domain.added'
  | 'domain.verified'
  | 'domain.removed'
  | 'team.member_invited'
  | 'team.member_joined'
  | 'team.member_removed';

export interface Webhook {
  id: string;
  url: string;
  events: EventType[];
  description?: string;
  is_active: boolean;
  created_at: string;
}

export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event_id: string;
  event_type: EventType;
  status_code?: number;
  response?: string;
  duration_ms: number;
  attempts: number;
  delivered_at?: string;
  failed_at?: string;
  created_at: string;
}

export interface TestResult {
  success: boolean;
  status_code: number;
  duration_ms: number;
  response?: string;
  error?: string;
}

export const webhooksApi = {
  list: async (): Promise<Webhook[]> => {
    const response = await apiClient.get<Webhook[]>('/api/v1/webhooks');
    return response.data;
  },

  create: async (data: {
    url: string;
    events: EventType[];
    description?: string;
  }): Promise<Webhook & { secret: string }> => {
    const response = await apiClient.post('/api/v1/webhooks', data);
    return response.data;
  },

  update: async (
    id: string,
    data: Partial<{ url: string; events: EventType[]; is_active: boolean }>
  ): Promise<Webhook> => {
    const response = await apiClient.put<Webhook>(`/api/v1/webhooks/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/webhooks/${id}`);
  },

  test: async (id: string): Promise<TestResult> => {
    const response = await apiClient.post<TestResult>(`/api/v1/webhooks/${id}/test`);
    return response.data;
  },

  rotateSecret: async (id: string): Promise<{ secret: string }> => {
    const response = await apiClient.post(`/api/v1/webhooks/${id}/rotate-secret`);
    return response.data;
  },

  listDeliveries: async (
    id: string,
    limit = 20,
    offset = 0
  ): Promise<{ deliveries: WebhookDelivery[]; total: number }> => {
    const response = await apiClient.get(
      `/api/v1/webhooks/${id}/deliveries?limit=${limit}&offset=${offset}`
    );
    return response.data;
  },

  retryDelivery: async (webhookId: string, deliveryId: string): Promise<WebhookDelivery> => {
    const response = await apiClient.post(
      `/api/v1/webhooks/${webhookId}/deliveries/${deliveryId}/retry`
    );
    return response.data;
  },

  getEventTypes: async (): Promise<Record<string, EventType[]>> => {
    const response = await apiClient.get('/api/v1/webhooks/events');
    return response.data;
  },
};
```
