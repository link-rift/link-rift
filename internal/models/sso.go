package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type SSOConfig struct {
	ID               uuid.UUID         `json:"id"`
	WorkspaceID      uuid.UUID         `json:"workspace_id"`
	Provider         string            `json:"provider"`
	EntityID         string            `json:"entity_id"`
	SSOURL           string            `json:"sso_url"`
	SLOURL           string            `json:"slo_url,omitempty"`
	Certificate      string            `json:"certificate"`
	IDPMetadataURL   string            `json:"idp_metadata_url,omitempty"`
	IDPMetadataXML   string            `json:"-"` // never expose in API responses
	AttributeMapping json.RawMessage   `json:"attribute_mapping"`
	IsEnabled        bool              `json:"is_enabled"`
	EnforceSSO       bool              `json:"enforce_sso"`
	AllowedDomains   []string          `json:"allowed_domains"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type SSOIdentity struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	WorkspaceID   uuid.UUID       `json:"workspace_id"`
	Provider      string          `json:"provider"`
	ExternalID    string          `json:"external_id"`
	Email         string          `json:"email"`
	Name          string          `json:"name,omitempty"`
	RawAttributes json.RawMessage `json:"raw_attributes,omitempty"`
	LastLoginAt   *time.Time      `json:"last_login_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateSSOConfigInput struct {
	Provider         string          `json:"provider" binding:"required"`
	EntityID         string          `json:"entity_id" binding:"required"`
	SSOURL           string          `json:"sso_url" binding:"required"`
	SLOURL           string          `json:"slo_url"`
	Certificate      string          `json:"certificate" binding:"required"`
	IDPMetadataURL   string          `json:"idp_metadata_url"`
	IDPMetadataXML   string          `json:"idp_metadata_xml"`
	AttributeMapping json.RawMessage `json:"attribute_mapping"`
	IsEnabled        bool            `json:"is_enabled"`
	EnforceSSO       bool            `json:"enforce_sso"`
	AllowedDomains   []string        `json:"allowed_domains"`
}

type UpdateSSOConfigInput struct {
	Provider         *string          `json:"provider"`
	EntityID         *string          `json:"entity_id"`
	SSOURL           *string          `json:"sso_url"`
	SLOURL           *string          `json:"slo_url"`
	Certificate      *string          `json:"certificate"`
	IDPMetadataURL   *string          `json:"idp_metadata_url"`
	IDPMetadataXML   *string          `json:"idp_metadata_xml"`
	AttributeMapping *json.RawMessage `json:"attribute_mapping"`
	IsEnabled        *bool            `json:"is_enabled"`
	EnforceSSO       *bool            `json:"enforce_sso"`
	AllowedDomains   *[]string        `json:"allowed_domains"`
}

func SSOConfigFromSqlc(s sqlc.SsoConfig) *SSOConfig {
	return &SSOConfig{
		ID:               s.ID,
		WorkspaceID:      s.WorkspaceID,
		Provider:         s.Provider,
		EntityID:         s.EntityID,
		SSOURL:           s.SsoUrl,
		SLOURL:           s.SloUrl.String,
		Certificate:      s.Certificate,
		IDPMetadataURL:   s.IdpMetadataUrl.String,
		IDPMetadataXML:   s.IdpMetadataXml.String,
		AttributeMapping: s.AttributeMapping,
		IsEnabled:        s.IsEnabled,
		EnforceSSO:       s.EnforceSso,
		AllowedDomains:   s.AllowedDomains,
		CreatedAt:        s.CreatedAt.Time,
		UpdatedAt:        s.UpdatedAt.Time,
	}
}

func SSOIdentityFromSqlc(s sqlc.SsoIdentity) *SSOIdentity {
	identity := &SSOIdentity{
		ID:            s.ID,
		UserID:        s.UserID,
		WorkspaceID:   s.WorkspaceID,
		Provider:      s.Provider,
		ExternalID:    s.ExternalID,
		Email:         s.Email,
		Name:          s.Name.String,
		RawAttributes: s.RawAttributes,
		CreatedAt:     s.CreatedAt.Time,
		UpdatedAt:     s.UpdatedAt.Time,
	}
	if s.LastLoginAt.Valid {
		t := s.LastLoginAt.Time
		identity.LastLoginAt = &t
	}
	return identity
}
