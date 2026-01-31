package models

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

// BioPage represents a link-in-bio page.
type BioPage struct {
	ID              uuid.UUID    `json:"id"`
	WorkspaceID     uuid.UUID    `json:"workspace_id"`
	Slug            string       `json:"slug"`
	Title           string       `json:"title"`
	Bio             *string      `json:"bio,omitempty"`
	AvatarURL       *string      `json:"avatar_url,omitempty"`
	ThemeID         *uuid.UUID   `json:"theme_id,omitempty"`
	CustomCSS       *string      `json:"custom_css,omitempty"`
	MetaTitle       *string      `json:"meta_title,omitempty"`
	MetaDescription *string      `json:"meta_description,omitempty"`
	OgImageURL      *string      `json:"og_image_url,omitempty"`
	IsPublished     bool         `json:"is_published"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Links           []*BioPageLink `json:"links,omitempty"`
	LinkCount       int          `json:"link_count,omitempty"`
}

// BioPageLink represents a link within a bio page.
type BioPageLink struct {
	ID           uuid.UUID  `json:"id"`
	BioPageID    uuid.UUID  `json:"bio_page_id"`
	Title        string     `json:"title"`
	URL          string     `json:"url"`
	Icon         *string    `json:"icon,omitempty"`
	Position     int32      `json:"position"`
	IsVisible    bool       `json:"is_visible"`
	VisibleFrom  *time.Time `json:"visible_from,omitempty"`
	VisibleUntil *time.Time `json:"visible_until,omitempty"`
	ClickCount   int64      `json:"click_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Input types

type CreateBioPageInput struct {
	Title           string  `json:"title" binding:"required"`
	Slug            string  `json:"slug" binding:"required"`
	Bio             *string `json:"bio,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	ThemeID         *string `json:"theme_id,omitempty"`
	MetaTitle       *string `json:"meta_title,omitempty"`
	MetaDescription *string `json:"meta_description,omitempty"`
	OgImageURL      *string `json:"og_image_url,omitempty"`
}

type UpdateBioPageInput struct {
	Title           *string `json:"title,omitempty"`
	Slug            *string `json:"slug,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	ThemeID         *string `json:"theme_id,omitempty"`
	CustomCSS       *string `json:"custom_css,omitempty"`
	MetaTitle       *string `json:"meta_title,omitempty"`
	MetaDescription *string `json:"meta_description,omitempty"`
	OgImageURL      *string `json:"og_image_url,omitempty"`
}

type CreateBioPageLinkInput struct {
	Title        string  `json:"title" binding:"required"`
	URL          string  `json:"url" binding:"required"`
	Icon         *string `json:"icon,omitempty"`
	IsVisible    *bool   `json:"is_visible,omitempty"`
	VisibleFrom  *string `json:"visible_from,omitempty"`
	VisibleUntil *string `json:"visible_until,omitempty"`
}

type UpdateBioPageLinkInput struct {
	Title        *string `json:"title,omitempty"`
	URL          *string `json:"url,omitempty"`
	Icon         *string `json:"icon,omitempty"`
	IsVisible    *bool   `json:"is_visible,omitempty"`
	VisibleFrom  *string `json:"visible_from,omitempty"`
	VisibleUntil *string `json:"visible_until,omitempty"`
}

type ReorderBioLinksInput struct {
	LinkIDs []string `json:"link_ids" binding:"required"`
}

// Theme types

type BioPageTheme struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	IsPremium   bool        `json:"is_premium"`
	Styles      ThemeStyles `json:"styles"`
}

type ThemeStyles struct {
	BackgroundColor string          `json:"background_color"`
	TextColor       string          `json:"text_color"`
	ButtonColor     string          `json:"button_color"`
	ButtonTextColor string          `json:"button_text_color"`
	ButtonStyle     string          `json:"button_style"`
	FontFamily      string          `json:"font_family"`
	Gradient        *GradientConfig `json:"gradient,omitempty"`
}

type GradientConfig struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Direction string `json:"direction"`
}

// ThemeIDToUUID generates a deterministic UUID from a theme string ID.
func ThemeIDToUUID(themeID string) uuid.UUID {
	hash := sha256.Sum256([]byte("linkrift-theme:" + themeID))
	u, _ := uuid.FromBytes(hash[:16])
	u[6] = (u[6] & 0x0f) | 0x50 // version 5
	u[8] = (u[8] & 0x3f) | 0x80 // variant 2
	return u
}

// ThemeUUIDToID looks up a theme string ID from a UUID.
func ThemeUUIDToID(themeUUID uuid.UUID) string {
	for id := range PredefinedThemes {
		if ThemeIDToUUID(id) == themeUUID {
			return id
		}
	}
	return ""
}

// PredefinedThemes contains all available built-in themes.
var PredefinedThemes = map[string]BioPageTheme{
	"minimal_light": {
		ID:          "minimal_light",
		Name:        "Minimal Light",
		Description: "Clean and simple light theme",
		IsPremium:   false,
		Styles: ThemeStyles{
			BackgroundColor: "#ffffff",
			TextColor:       "#1a1a1a",
			ButtonColor:     "#1a1a1a",
			ButtonTextColor: "#ffffff",
			ButtonStyle:     "rounded",
			FontFamily:      "Inter, sans-serif",
		},
	},
	"minimal_dark": {
		ID:          "minimal_dark",
		Name:        "Minimal Dark",
		Description: "Sleek dark theme",
		IsPremium:   false,
		Styles: ThemeStyles{
			BackgroundColor: "#0a0a0a",
			TextColor:       "#f5f5f5",
			ButtonColor:     "#f5f5f5",
			ButtonTextColor: "#0a0a0a",
			ButtonStyle:     "rounded",
			FontFamily:      "Inter, sans-serif",
		},
	},
	"gradient_sunset": {
		ID:          "gradient_sunset",
		Name:        "Gradient Sunset",
		Description: "Warm sunset gradient theme",
		IsPremium:   true,
		Styles: ThemeStyles{
			BackgroundColor: "#ff6b35",
			TextColor:       "#ffffff",
			ButtonColor:     "rgba(255,255,255,0.2)",
			ButtonTextColor: "#ffffff",
			ButtonStyle:     "pill",
			FontFamily:      "Inter, sans-serif",
			Gradient: &GradientConfig{
				From:      "#ff6b35",
				To:        "#f72585",
				Direction: "to bottom right",
			},
		},
	},
}

// Conversion from SQLC model

func BioPageFromSqlc(b sqlc.BioPage) *BioPage {
	page := &BioPage{
		ID:          b.ID,
		WorkspaceID: b.WorkspaceID,
		Slug:        b.Slug,
		Title:       b.Title,
		IsPublished: b.IsPublished,
	}

	if b.Bio.Valid {
		page.Bio = &b.Bio.String
	}
	if b.AvatarUrl.Valid {
		page.AvatarURL = &b.AvatarUrl.String
	}
	if b.ThemeID.Valid {
		tid := uuid.UUID(b.ThemeID.Bytes)
		page.ThemeID = &tid
	}
	if b.CustomCss.Valid {
		page.CustomCSS = &b.CustomCss.String
	}
	if b.MetaTitle.Valid {
		page.MetaTitle = &b.MetaTitle.String
	}
	if b.MetaDescription.Valid {
		page.MetaDescription = &b.MetaDescription.String
	}
	if b.OgImageUrl.Valid {
		page.OgImageURL = &b.OgImageUrl.String
	}
	if b.CreatedAt.Valid {
		page.CreatedAt = b.CreatedAt.Time
	}
	if b.UpdatedAt.Valid {
		page.UpdatedAt = b.UpdatedAt.Time
	}

	return page
}

func BioPageLinkFromSqlc(l sqlc.BioPageLink) *BioPageLink {
	link := &BioPageLink{
		ID:         l.ID,
		BioPageID:  l.BioPageID,
		Title:      l.Title,
		URL:        l.Url,
		Position:   l.Position,
		IsVisible:  l.IsVisible,
		ClickCount: l.ClickCount,
	}

	if l.Icon.Valid {
		link.Icon = &l.Icon.String
	}
	if l.VisibleFrom.Valid {
		t := l.VisibleFrom.Time
		link.VisibleFrom = &t
	}
	if l.VisibleUntil.Valid {
		t := l.VisibleUntil.Time
		link.VisibleUntil = &t
	}
	if l.CreatedAt.Valid {
		link.CreatedAt = l.CreatedAt.Time
	}
	if l.UpdatedAt.Valid {
		link.UpdatedAt = l.UpdatedAt.Time
	}

	return link
}

// PublicBioPageResponse is the response for the public /b/:slug endpoint.
type PublicBioPageResponse struct {
	Title           string           `json:"title"`
	Bio             *string          `json:"bio,omitempty"`
	AvatarURL       *string          `json:"avatar_url,omitempty"`
	Slug            string           `json:"slug"`
	Theme           *BioPageTheme    `json:"theme,omitempty"`
	CustomCSS       *string          `json:"custom_css,omitempty"`
	MetaTitle       *string          `json:"meta_title,omitempty"`
	MetaDescription *string          `json:"meta_description,omitempty"`
	OgImageURL      *string          `json:"og_image_url,omitempty"`
	Links           []PublicBioLink  `json:"links"`
}

type PublicBioLink struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	URL   string    `json:"url"`
	Icon  *string   `json:"icon,omitempty"`
}

