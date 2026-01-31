package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

// ValidAPIKeyScopes defines all valid API key scopes.
var ValidAPIKeyScopes = []string{
	"links:read",
	"links:write",
	"domains:read",
	"domains:write",
	"analytics:read",
	"bio_pages:read",
	"bio_pages:write",
	"qr:read",
	"qr:write",
	"webhooks:read",
	"webhooks:write",
}

type APIKey struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id"`
	Name         string     `json:"name"`
	KeyHash      string     `json:"-"`
	KeyPrefix    string     `json:"key_prefix"`
	Scopes       []string   `json:"scopes"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	RequestCount int64      `json:"request_count"`
	RateLimit    *int32     `json:"rate_limit,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateAPIKeyInput struct {
	Name      string   `json:"name" binding:"required,min=1,max=100"`
	Scopes    []string `json:"scopes" binding:"required,min=1"`
	ExpiresAt *string  `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey *APIKey `json:"api_key"`
	Key    string  `json:"key"`
}

func APIKeyFromSqlc(k sqlc.ApiKey) *APIKey {
	ak := &APIKey{
		ID:           k.ID,
		UserID:       k.UserID,
		Name:         k.Name,
		KeyHash:      k.KeyHash,
		KeyPrefix:    k.KeyPrefix,
		Scopes:       k.Scopes,
		RequestCount: k.RequestCount,
	}
	if k.WorkspaceID.Valid {
		ak.WorkspaceID = uuid.UUID(k.WorkspaceID.Bytes)
	}
	if k.LastUsedAt.Valid {
		t := k.LastUsedAt.Time
		ak.LastUsedAt = &t
	}
	if k.RateLimit.Valid {
		v := k.RateLimit.Int32
		ak.RateLimit = &v
	}
	if k.ExpiresAt.Valid {
		t := k.ExpiresAt.Time
		ak.ExpiresAt = &t
	}
	if k.CreatedAt.Valid {
		ak.CreatedAt = k.CreatedAt.Time
	}
	return ak
}

func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func IsValidScope(scope string) bool {
	for _, s := range ValidAPIKeyScopes {
		if s == scope {
			return true
		}
	}
	return false
}
