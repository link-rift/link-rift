package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	RefreshTokenHash string    `json:"-"`
	IPAddress        string    `json:"ip_address,omitempty"`
	UserAgent        *string   `json:"user_agent,omitempty"`
	DeviceName       *string   `json:"device_name,omitempty"`
	IsRevoked        bool      `json:"is_revoked"`
	LastActiveAt     time.Time `json:"last_active_at"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func SessionFromSqlc(s sqlc.Session) *Session {
	session := &Session{
		ID:               s.ID,
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		IPAddress:        s.IpAddress,
		IsRevoked:        s.IsRevoked,
	}

	if s.UserAgent.Valid {
		session.UserAgent = &s.UserAgent.String
	}
	if s.DeviceName.Valid {
		session.DeviceName = &s.DeviceName.String
	}
	if s.LastActiveAt.Valid {
		session.LastActiveAt = s.LastActiveAt.Time
	}
	if s.CreatedAt.Valid {
		session.CreatedAt = s.CreatedAt.Time
	}
	if s.ExpiresAt.Valid {
		session.ExpiresAt = s.ExpiresAt.Time
	}

	return session
}
