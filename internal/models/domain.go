package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

const (
	SSLPending = "pending"
	SSLActive  = "active"
	SSLFailed  = "failed"
)

type Domain struct {
	ID                 uuid.UUID  `json:"id"`
	WorkspaceID        uuid.UUID  `json:"workspace_id"`
	Domain             string     `json:"domain"`
	IsVerified         bool       `json:"is_verified"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	SSLStatus          string     `json:"ssl_status"`
	SSLExpiresAt       *time.Time `json:"ssl_expires_at,omitempty"`
	DNSRecords         []byte     `json:"dns_records,omitempty"`
	LastDNSCheckAt     *time.Time `json:"last_dns_check_at,omitempty"`
	DefaultRedirectURL *string    `json:"default_redirect_url,omitempty"`
	Custom404URL       *string    `json:"custom_404_url,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateDomainInput struct {
	Domain string `json:"domain" binding:"required"`
}

type UpdateDomainInput struct {
	DefaultRedirectURL *string `json:"default_redirect_url,omitempty"`
	Custom404URL       *string `json:"custom_404_url,omitempty"`
}

type DNSRecordsData struct {
	VerificationToken string `json:"verification_token"`
}

type VerificationInstructions struct {
	Records []DNSRecordInstruction `json:"records"`
}

type DNSRecordInstruction struct {
	Type  string `json:"type"`
	Host  string `json:"host"`
	Value string `json:"value"`
}

func DomainFromSqlc(d sqlc.Domain) *Domain {
	domain := &Domain{
		ID:          d.ID,
		WorkspaceID: d.WorkspaceID,
		Domain:      d.Domain,
		IsVerified:  d.IsVerified,
		SSLStatus:   d.SslStatus,
		DNSRecords:  d.DnsRecords,
	}

	if d.VerifiedAt.Valid {
		t := d.VerifiedAt.Time
		domain.VerifiedAt = &t
	}
	if d.SslExpiresAt.Valid {
		t := d.SslExpiresAt.Time
		domain.SSLExpiresAt = &t
	}
	if d.LastDnsCheckAt.Valid {
		t := d.LastDnsCheckAt.Time
		domain.LastDNSCheckAt = &t
	}
	if d.DefaultRedirectUrl.Valid {
		domain.DefaultRedirectURL = &d.DefaultRedirectUrl.String
	}
	if d.Custom404Url.Valid {
		domain.Custom404URL = &d.Custom404Url.String
	}
	if d.CreatedAt.Valid {
		domain.CreatedAt = d.CreatedAt.Time
	}
	if d.UpdatedAt.Valid {
		domain.UpdatedAt = d.UpdatedAt.Time
	}

	return domain
}

func (d *Domain) GetVerificationToken() string {
	if len(d.DNSRecords) == 0 {
		return ""
	}
	var data DNSRecordsData
	if err := json.Unmarshal(d.DNSRecords, &data); err != nil {
		return ""
	}
	return data.VerificationToken
}
