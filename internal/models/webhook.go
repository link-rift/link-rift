package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

// ValidWebhookEvents defines all valid webhook event types.
var ValidWebhookEvents = []string{
	"link.created",
	"link.updated",
	"link.deleted",
	"link.clicked",
	"link.expired",
	"qr.created",
	"qr.scanned",
	"biopage.created",
	"biopage.updated",
	"domain.added",
	"domain.verified",
	"domain.removed",
	"team.member_invited",
	"team.member_joined",
	"team.member_removed",
}

type Webhook struct {
	ID              uuid.UUID  `json:"id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id"`
	URL             string     `json:"url"`
	Secret          string     `json:"-"`
	Events          []string   `json:"events"`
	IsActive        bool       `json:"is_active"`
	FailureCount    int32      `json:"failure_count"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	LastSuccessAt   *time.Time `json:"last_success_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type WebhookDelivery struct {
	ID             uuid.UUID       `json:"id"`
	WebhookID      uuid.UUID       `json:"webhook_id"`
	Event          string          `json:"event"`
	Payload        json.RawMessage `json:"payload"`
	ResponseStatus *int32          `json:"response_status,omitempty"`
	ResponseBody   *string         `json:"response_body,omitempty"`
	Attempts       int32           `json:"attempts"`
	MaxAttempts    int32           `json:"max_attempts"`
	LastAttemptAt  *time.Time      `json:"last_attempt_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type WebhookEvent struct {
	Event       string          `json:"event"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Data        json.RawMessage `json:"data"`
}

type CreateWebhookInput struct {
	URL    string   `json:"url" binding:"required,url"`
	Events []string `json:"events" binding:"required,min=1"`
}

type CreateWebhookResponse struct {
	Webhook *Webhook `json:"webhook"`
	Secret  string   `json:"secret"`
}

func WebhookFromSqlc(w sqlc.Webhook) *Webhook {
	wh := &Webhook{
		ID:           w.ID,
		WorkspaceID:  w.WorkspaceID,
		URL:          w.Url,
		Secret:       w.Secret,
		Events:       w.Events,
		IsActive:     w.IsActive,
		FailureCount: w.FailureCount,
	}
	if w.LastTriggeredAt.Valid {
		t := w.LastTriggeredAt.Time
		wh.LastTriggeredAt = &t
	}
	if w.LastSuccessAt.Valid {
		t := w.LastSuccessAt.Time
		wh.LastSuccessAt = &t
	}
	if w.CreatedAt.Valid {
		wh.CreatedAt = w.CreatedAt.Time
	}
	if w.UpdatedAt.Valid {
		wh.UpdatedAt = w.UpdatedAt.Time
	}
	return wh
}

func WebhookDeliveryFromSqlc(d sqlc.WebhookDelivery) *WebhookDelivery {
	wd := &WebhookDelivery{
		ID:          d.ID,
		WebhookID:   d.WebhookID,
		Event:       d.Event,
		Payload:     d.Payload,
		Attempts:    d.Attempts,
		MaxAttempts: d.MaxAttempts,
	}
	if d.ResponseStatus.Valid {
		v := d.ResponseStatus.Int32
		wd.ResponseStatus = &v
	}
	if d.ResponseBody.Valid {
		wd.ResponseBody = &d.ResponseBody.String
	}
	if d.LastAttemptAt.Valid {
		t := d.LastAttemptAt.Time
		wd.LastAttemptAt = &t
	}
	if d.CompletedAt.Valid {
		t := d.CompletedAt.Time
		wd.CompletedAt = &t
	}
	if d.CreatedAt.Valid {
		wd.CreatedAt = d.CreatedAt.Time
	}
	return wd
}

func IsValidWebhookEvent(event string) bool {
	for _, e := range ValidWebhookEvents {
		if e == event {
			return true
		}
	}
	return false
}
