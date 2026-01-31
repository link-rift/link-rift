package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type QRCode struct {
	ID              uuid.UUID `json:"id"`
	LinkID          uuid.UUID `json:"link_id"`
	QRType          string    `json:"qr_type"`
	ErrorCorrection string    `json:"error_correction"`
	ForegroundColor string    `json:"foreground_color"`
	BackgroundColor string    `json:"background_color"`
	LogoURL         *string   `json:"logo_url,omitempty"`
	PngURL          *string   `json:"png_url,omitempty"`
	SvgURL          *string   `json:"svg_url,omitempty"`
	DotStyle        string    `json:"dot_style"`
	CornerStyle     string    `json:"corner_style"`
	Size            int32     `json:"size"`
	Margin          int32     `json:"margin"`
	ScanCount       int64     `json:"scan_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type QRCodeResponse struct {
	ID              uuid.UUID `json:"id"`
	LinkID          uuid.UUID `json:"link_id"`
	QRType          string    `json:"qr_type"`
	ErrorCorrection string    `json:"error_correction"`
	ForegroundColor string    `json:"foreground_color"`
	BackgroundColor string    `json:"background_color"`
	LogoURL         *string   `json:"logo_url,omitempty"`
	PngURL          *string   `json:"png_url,omitempty"`
	SvgURL          *string   `json:"svg_url,omitempty"`
	DotStyle        string    `json:"dot_style"`
	CornerStyle     string    `json:"corner_style"`
	Size            int32     `json:"size"`
	Margin          int32     `json:"margin"`
	ScanCount       int64     `json:"scan_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateQRCodeInput struct {
	QRType          string  `json:"qr_type"`
	ErrorCorrection string  `json:"error_correction"`
	ForegroundColor string  `json:"foreground_color"`
	BackgroundColor string  `json:"background_color"`
	LogoURL         *string `json:"logo_url,omitempty"`
	DotStyle        string  `json:"dot_style"`
	CornerStyle     string  `json:"corner_style"`
	Size            *int32  `json:"size,omitempty"`
	Margin          *int32  `json:"margin,omitempty"`
}

type BulkQRCodeInput struct {
	LinkIDs []uuid.UUID       `json:"link_ids" binding:"required,min=1,max=50"`
	Options CreateQRCodeInput `json:"options"`
}

func QRCodeFromSqlc(q sqlc.QrCode) *QRCode {
	qr := &QRCode{
		ID:              q.ID,
		LinkID:          q.LinkID,
		QRType:          q.QrType,
		ErrorCorrection: q.ErrorCorrection,
		ForegroundColor: q.ForegroundColor,
		BackgroundColor: q.BackgroundColor,
		DotStyle:        q.DotStyle,
		CornerStyle:     q.CornerStyle,
		Size:            q.Size,
		Margin:          q.Margin,
		ScanCount:       q.ScanCount,
	}

	if q.LogoUrl.Valid {
		qr.LogoURL = &q.LogoUrl.String
	}
	if q.PngUrl.Valid {
		qr.PngURL = &q.PngUrl.String
	}
	if q.SvgUrl.Valid {
		qr.SvgURL = &q.SvgUrl.String
	}
	if q.CreatedAt.Valid {
		qr.CreatedAt = q.CreatedAt.Time
	}
	if q.UpdatedAt.Valid {
		qr.UpdatedAt = q.UpdatedAt.Time
	}

	return qr
}

func (q *QRCode) ToResponse() *QRCodeResponse {
	return &QRCodeResponse{
		ID:              q.ID,
		LinkID:          q.LinkID,
		QRType:          q.QRType,
		ErrorCorrection: q.ErrorCorrection,
		ForegroundColor: q.ForegroundColor,
		BackgroundColor: q.BackgroundColor,
		LogoURL:         q.LogoURL,
		PngURL:          q.PngURL,
		SvgURL:          q.SvgURL,
		DotStyle:        q.DotStyle,
		CornerStyle:     q.CornerStyle,
		Size:            q.Size,
		Margin:          q.Margin,
		ScanCount:       q.ScanCount,
		CreatedAt:       q.CreatedAt,
		UpdatedAt:       q.UpdatedAt,
	}
}
