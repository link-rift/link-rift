package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type Link struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id"`
	DomainID     *uuid.UUID `json:"domain_id,omitempty"`
	URL          string     `json:"url"`
	ShortCode    string     `json:"short_code"`
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	FaviconURL   *string    `json:"favicon_url,omitempty"`
	OgImageURL   *string    `json:"og_image_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	PasswordHash *string    `json:"-"`
	HasPassword  bool       `json:"has_password"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	MaxClicks    *int32     `json:"max_clicks,omitempty"`
	UTMSource    *string    `json:"utm_source,omitempty"`
	UTMMedium    *string    `json:"utm_medium,omitempty"`
	UTMCampaign  *string    `json:"utm_campaign,omitempty"`
	UTMTerm      *string    `json:"utm_term,omitempty"`
	UTMContent   *string    `json:"utm_content,omitempty"`
	TotalClicks  int64      `json:"total_clicks"`
	UniqueClicks int64      `json:"unique_clicks"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type LinkResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id"`
	DomainID     *uuid.UUID `json:"domain_id,omitempty"`
	URL          string     `json:"url"`
	ShortCode    string     `json:"short_code"`
	ShortURL     string     `json:"short_url"`
	Title        *string    `json:"title,omitempty"`
	Description  *string    `json:"description,omitempty"`
	FaviconURL   *string    `json:"favicon_url,omitempty"`
	OgImageURL   *string    `json:"og_image_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	HasPassword  bool       `json:"has_password"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	MaxClicks    *int32     `json:"max_clicks,omitempty"`
	UTMSource    *string    `json:"utm_source,omitempty"`
	UTMMedium    *string    `json:"utm_medium,omitempty"`
	UTMCampaign  *string    `json:"utm_campaign,omitempty"`
	UTMTerm      *string    `json:"utm_term,omitempty"`
	UTMContent   *string    `json:"utm_content,omitempty"`
	TotalClicks  int64      `json:"total_clicks"`
	UniqueClicks int64      `json:"unique_clicks"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateLinkInput struct {
	URL         string  `json:"url" binding:"required,url"`
	ShortCode   *string `json:"short_code,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Password    *string `json:"password,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	MaxClicks   *int32  `json:"max_clicks,omitempty"`
	UTMSource   *string `json:"utm_source,omitempty"`
	UTMMedium   *string `json:"utm_medium,omitempty"`
	UTMCampaign *string `json:"utm_campaign,omitempty"`
	UTMTerm     *string `json:"utm_term,omitempty"`
	UTMContent  *string `json:"utm_content,omitempty"`
}

type UpdateLinkInput struct {
	URL         *string `json:"url,omitempty" binding:"omitempty,url"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Password    *string `json:"password,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	MaxClicks   *int32  `json:"max_clicks,omitempty"`
}

type BulkCreateLinkInput struct {
	Links []CreateLinkInput `json:"links" binding:"required,min=1,max=100,dive"`
}

type LinkFilter struct {
	Search   *string `form:"search"`
	IsActive *bool   `form:"is_active"`
}

type Pagination struct {
	Limit  int `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}

type LinkListResult struct {
	Links []*LinkResponse `json:"links"`
	Total int64           `json:"total"`
}

type LinkQuickStats struct {
	TotalClicks  int64     `json:"total_clicks"`
	UniqueClicks int64     `json:"unique_clicks"`
	Clicks24h    int64     `json:"clicks_24h"`
	Clicks7d     int64     `json:"clicks_7d"`
	CreatedAt    time.Time `json:"created_at"`
}

func LinkFromSqlc(l sqlc.Link) *Link {
	link := &Link{
		ID:           l.ID,
		UserID:       l.UserID,
		WorkspaceID:  l.WorkspaceID,
		URL:          l.Url,
		ShortCode:    l.ShortCode,
		IsActive:     l.IsActive,
		TotalClicks:  l.TotalClicks,
		UniqueClicks: l.UniqueClicks,
	}

	if l.DomainID.Valid {
		id := uuid.UUID(l.DomainID.Bytes)
		link.DomainID = &id
	}
	if l.Title.Valid {
		link.Title = &l.Title.String
	}
	if l.Description.Valid {
		link.Description = &l.Description.String
	}
	if l.FaviconUrl.Valid {
		link.FaviconURL = &l.FaviconUrl.String
	}
	if l.OgImageUrl.Valid {
		link.OgImageURL = &l.OgImageUrl.String
	}
	if l.PasswordHash.Valid {
		link.PasswordHash = &l.PasswordHash.String
		link.HasPassword = true
	}
	if l.ExpiresAt.Valid {
		t := l.ExpiresAt.Time
		link.ExpiresAt = &t
	}
	if l.MaxClicks.Valid {
		v := l.MaxClicks.Int32
		link.MaxClicks = &v
	}
	if l.UtmSource.Valid {
		link.UTMSource = &l.UtmSource.String
	}
	if l.UtmMedium.Valid {
		link.UTMMedium = &l.UtmMedium.String
	}
	if l.UtmCampaign.Valid {
		link.UTMCampaign = &l.UtmCampaign.String
	}
	if l.UtmTerm.Valid {
		link.UTMTerm = &l.UtmTerm.String
	}
	if l.UtmContent.Valid {
		link.UTMContent = &l.UtmContent.String
	}
	if l.CreatedAt.Valid {
		link.CreatedAt = l.CreatedAt.Time
	}
	if l.UpdatedAt.Valid {
		link.UpdatedAt = l.UpdatedAt.Time
	}

	return link
}

func LinkFromSqlcRow(r sqlc.ListLinksForWorkspaceRow) *Link {
	l := &Link{
		ID:           r.ID,
		UserID:       r.UserID,
		WorkspaceID:  r.WorkspaceID,
		URL:          r.Url,
		ShortCode:    r.ShortCode,
		IsActive:     r.IsActive,
		TotalClicks:  r.TotalClicks,
		UniqueClicks: r.UniqueClicks,
	}

	if r.DomainID.Valid {
		id := uuid.UUID(r.DomainID.Bytes)
		l.DomainID = &id
	}
	if r.Title.Valid {
		l.Title = &r.Title.String
	}
	if r.Description.Valid {
		l.Description = &r.Description.String
	}
	if r.FaviconUrl.Valid {
		l.FaviconURL = &r.FaviconUrl.String
	}
	if r.OgImageUrl.Valid {
		l.OgImageURL = &r.OgImageUrl.String
	}
	if r.PasswordHash.Valid {
		l.PasswordHash = &r.PasswordHash.String
		l.HasPassword = true
	}
	if r.ExpiresAt.Valid {
		t := r.ExpiresAt.Time
		l.ExpiresAt = &t
	}
	if r.MaxClicks.Valid {
		v := r.MaxClicks.Int32
		l.MaxClicks = &v
	}
	if r.UtmSource.Valid {
		l.UTMSource = &r.UtmSource.String
	}
	if r.UtmMedium.Valid {
		l.UTMMedium = &r.UtmMedium.String
	}
	if r.UtmCampaign.Valid {
		l.UTMCampaign = &r.UtmCampaign.String
	}
	if r.UtmTerm.Valid {
		l.UTMTerm = &r.UtmTerm.String
	}
	if r.UtmContent.Valid {
		l.UTMContent = &r.UtmContent.String
	}
	if r.CreatedAt.Valid {
		l.CreatedAt = r.CreatedAt.Time
	}
	if r.UpdatedAt.Valid {
		l.UpdatedAt = r.UpdatedAt.Time
	}

	return l
}

func (l *Link) ToResponse(redirectBaseURL string) *LinkResponse {
	return &LinkResponse{
		ID:           l.ID,
		UserID:       l.UserID,
		WorkspaceID:  l.WorkspaceID,
		DomainID:     l.DomainID,
		URL:          l.URL,
		ShortCode:    l.ShortCode,
		ShortURL:     redirectBaseURL + "/" + l.ShortCode,
		Title:        l.Title,
		Description:  l.Description,
		FaviconURL:   l.FaviconURL,
		OgImageURL:   l.OgImageURL,
		IsActive:     l.IsActive,
		HasPassword:  l.HasPassword,
		ExpiresAt:    l.ExpiresAt,
		MaxClicks:    l.MaxClicks,
		UTMSource:    l.UTMSource,
		UTMMedium:    l.UTMMedium,
		UTMCampaign:  l.UTMCampaign,
		UTMTerm:      l.UTMTerm,
		UTMContent:   l.UTMContent,
		TotalClicks:  l.TotalClicks,
		UniqueClicks: l.UniqueClicks,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}

func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*l.ExpiresAt)
}

func (l *Link) IsClickLimitReached() bool {
	if l.MaxClicks == nil {
		return false
	}
	return l.TotalClicks >= int64(*l.MaxClicks)
}

func OptionalText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func OptionalBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func OptionalInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func OptionalTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
