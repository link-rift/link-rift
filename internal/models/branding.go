package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type WorkspaceBranding struct {
	ID               uuid.UUID `json:"id"`
	WorkspaceID      uuid.UUID `json:"workspace_id"`
	LogoURL          string    `json:"logo_url,omitempty"`
	LogoDarkURL      string    `json:"logo_dark_url,omitempty"`
	FaviconURL       string    `json:"favicon_url,omitempty"`
	PrimaryColor     string    `json:"primary_color,omitempty"`
	SecondaryColor   string    `json:"secondary_color,omitempty"`
	AccentColor      string    `json:"accent_color,omitempty"`
	CustomCSS        string    `json:"custom_css,omitempty"`
	HidePoweredBy    bool      `json:"hide_powered_by"`
	CustomFooterText string    `json:"custom_footer_text,omitempty"`
	CustomFooterURL  string    `json:"custom_footer_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UpdateBrandingInput struct {
	LogoURL          *string `json:"logo_url"`
	LogoDarkURL      *string `json:"logo_dark_url"`
	FaviconURL       *string `json:"favicon_url"`
	PrimaryColor     *string `json:"primary_color"`
	SecondaryColor   *string `json:"secondary_color"`
	AccentColor      *string `json:"accent_color"`
	CustomCSS        *string `json:"custom_css"`
	HidePoweredBy    *bool   `json:"hide_powered_by"`
	CustomFooterText *string `json:"custom_footer_text"`
	CustomFooterURL  *string `json:"custom_footer_url"`
}

func BrandingFromSqlc(b sqlc.WorkspaceBranding) *WorkspaceBranding {
	return &WorkspaceBranding{
		ID:               b.ID,
		WorkspaceID:      b.WorkspaceID,
		LogoURL:          b.LogoUrl.String,
		LogoDarkURL:      b.LogoDarkUrl.String,
		FaviconURL:       b.FaviconUrl.String,
		PrimaryColor:     b.PrimaryColor.String,
		SecondaryColor:   b.SecondaryColor.String,
		AccentColor:      b.AccentColor.String,
		CustomCSS:        b.CustomCss.String,
		HidePoweredBy:    b.HidePoweredBy,
		CustomFooterText: b.CustomFooterText.String,
		CustomFooterURL:  b.CustomFooterUrl.String,
		CreatedAt:        b.CreatedAt.Time,
		UpdatedAt:        b.UpdatedAt.Time,
	}
}
