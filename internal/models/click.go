package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type Click struct {
	ID             uuid.UUID  `json:"id"`
	LinkID         uuid.UUID  `json:"link_id"`
	ClickedAt      time.Time  `json:"clicked_at"`
	VisitorID      *string    `json:"visitor_id,omitempty"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      *string    `json:"user_agent,omitempty"`
	Referer        *string    `json:"referer,omitempty"`
	CountryCode    *string    `json:"country_code,omitempty"`
	Region         *string    `json:"region,omitempty"`
	City           *string    `json:"city,omitempty"`
	DeviceType     *string    `json:"device_type,omitempty"`
	Browser        *string    `json:"browser,omitempty"`
	BrowserVersion *string    `json:"browser_version,omitempty"`
	OS             *string    `json:"os,omitempty"`
	OSVersion      *string    `json:"os_version,omitempty"`
	IsBot          bool       `json:"is_bot"`
	UTMSource      *string    `json:"utm_source,omitempty"`
	UTMMedium      *string    `json:"utm_medium,omitempty"`
	UTMCampaign    *string    `json:"utm_campaign,omitempty"`
}

// ClickEvent is a lightweight struct for the async tracking pipeline.
type ClickEvent struct {
	LinkID      uuid.UUID `json:"link_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	ShortCode   string    `json:"short_code"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	Timestamp   time.Time `json:"timestamp"`
}

// ClickNotification is published to Redis Pub/Sub for real-time WebSocket updates.
type ClickNotification struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	LinkID      uuid.UUID `json:"link_id"`
	ShortCode   string    `json:"short_code"`
	Timestamp   time.Time `json:"timestamp"`
	CountryCode string    `json:"country_code,omitempty"`
	DeviceType  string    `json:"device_type,omitempty"`
	Browser     string    `json:"browser,omitempty"`
	Referer     string    `json:"referer,omitempty"`
}

func ClickFromSqlc(c sqlc.Click) *Click {
	click := &Click{
		ID:        c.ID,
		LinkID:    c.LinkID,
		IPAddress: c.IpAddress,
		IsBot:     c.IsBot,
	}

	if c.ClickedAt.Valid {
		click.ClickedAt = c.ClickedAt.Time
	}
	if c.VisitorID.Valid {
		click.VisitorID = &c.VisitorID.String
	}
	if c.UserAgent.Valid {
		click.UserAgent = &c.UserAgent.String
	}
	if c.Referer.Valid {
		click.Referer = &c.Referer.String
	}
	if c.CountryCode.Valid {
		click.CountryCode = &c.CountryCode.String
	}
	if c.Region.Valid {
		click.Region = &c.Region.String
	}
	if c.City.Valid {
		click.City = &c.City.String
	}
	if c.DeviceType.Valid {
		click.DeviceType = &c.DeviceType.String
	}
	if c.Browser.Valid {
		click.Browser = &c.Browser.String
	}
	if c.BrowserVersion.Valid {
		click.BrowserVersion = &c.BrowserVersion.String
	}
	if c.Os.Valid {
		click.OS = &c.Os.String
	}
	if c.OsVersion.Valid {
		click.OSVersion = &c.OsVersion.String
	}
	if c.UtmSource.Valid {
		click.UTMSource = &c.UtmSource.String
	}
	if c.UtmMedium.Valid {
		click.UTMMedium = &c.UtmMedium.String
	}
	if c.UtmCampaign.Valid {
		click.UTMCampaign = &c.UtmCampaign.String
	}

	return click
}
