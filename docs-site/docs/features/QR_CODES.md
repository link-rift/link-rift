# QR Codes

> Last Updated: 2025-01-24

Linkrift provides comprehensive QR code generation with customization options, batch processing, and dedicated analytics tracking for QR-originated traffic.

## Table of Contents

- [Overview](#overview)
- [Dynamic vs Static QR Codes](#dynamic-vs-static-qr-codes)
- [QR Code Generation](#qr-code-generation)
  - [go-qrcode Integration](#go-qrcode-integration)
  - [Customization Options](#customization-options)
- [Batch Generation](#batch-generation)
- [QR Code Analytics](#qr-code-analytics)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

QR code features in Linkrift include:

- **Dynamic QR codes** that redirect through Linkrift for tracking
- **Static QR codes** that encode URLs directly (no tracking)
- **Extensive customization** including colors, logos, shapes, and frames
- **Batch generation** for creating multiple QR codes at once
- **Dedicated analytics** for QR code scans vs regular clicks

## Dynamic vs Static QR Codes

| Feature | Dynamic QR Code | Static QR Code |
|---------|-----------------|----------------|
| Encoded URL | Short Linkrift URL | Original destination URL |
| Tracking | Full analytics | No tracking |
| Editable | Can change destination | Cannot change after creation |
| QR Size | Small (shorter URL) | Larger (longer URL) |
| Use Case | Marketing, tracking | Business cards, permanent links |

---

## QR Code Generation

### go-qrcode Integration

```go
// internal/qrcode/generator.go
package qrcode

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"github.com/skip2/go-qrcode"
	"github.com/nfnt/resize"
	"github.com/link-rift/link-rift/internal/storage"
)

// QRCodeType defines whether the QR code is dynamic or static
type QRCodeType string

const (
	QRTypeDynamic QRCodeType = "dynamic"
	QRTypeStatic  QRCodeType = "static"
)

// QRCode represents a generated QR code
type QRCode struct {
	ID            string           `json:"id" db:"id"`
	LinkID        string           `json:"link_id" db:"link_id"`
	WorkspaceID   string           `json:"workspace_id" db:"workspace_id"`
	Type          QRCodeType       `json:"type" db:"type"`
	EncodedURL    string           `json:"encoded_url" db:"encoded_url"`
	Options       QRCodeOptions    `json:"options" db:"options"`
	ImageURL      string           `json:"image_url" db:"image_url"`
	TotalScans    int64            `json:"total_scans" db:"total_scans"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
}

// QRCodeOptions defines customization options
type QRCodeOptions struct {
	Size            int            `json:"size"`             // Image size in pixels
	ErrorCorrection string         `json:"error_correction"` // L, M, Q, H
	ForegroundColor string         `json:"foreground_color"` // Hex color
	BackgroundColor string         `json:"background_color"` // Hex color
	LogoURL         string         `json:"logo_url,omitempty"`
	LogoSize        int            `json:"logo_size,omitempty"` // Percentage of QR size
	DotStyle        string         `json:"dot_style"`           // square, rounded, dots
	CornerStyle     string         `json:"corner_style"`        // square, rounded
	Frame           *FrameOptions  `json:"frame,omitempty"`
	Margin          int            `json:"margin"` // Quiet zone in modules
}

// FrameOptions defines QR code frame customization
type FrameOptions struct {
	Style     string `json:"style"`      // none, simple, rounded, banner
	Color     string `json:"color"`      // Hex color
	Text      string `json:"text"`       // Text to display
	TextColor string `json:"text_color"` // Hex color
	Position  string `json:"position"`   // top, bottom
}

// Generator handles QR code generation
type Generator struct {
	storage storage.ObjectStorage
}

// NewGenerator creates a new QR code generator
func NewGenerator(storage storage.ObjectStorage) *Generator {
	return &Generator{storage: storage}
}

// Generate creates a QR code image
func (g *Generator) Generate(url string, options QRCodeOptions) ([]byte, error) {
	// Set defaults
	if options.Size == 0 {
		options.Size = 512
	}
	if options.ErrorCorrection == "" {
		options.ErrorCorrection = "M"
	}
	if options.ForegroundColor == "" {
		options.ForegroundColor = "#000000"
	}
	if options.BackgroundColor == "" {
		options.BackgroundColor = "#FFFFFF"
	}
	if options.Margin == 0 {
		options.Margin = 4
	}

	// Map error correction level
	ecLevel := qrcode.Medium
	switch options.ErrorCorrection {
	case "L":
		ecLevel = qrcode.Low
	case "M":
		ecLevel = qrcode.Medium
	case "Q":
		ecLevel = qrcode.High
	case "H":
		ecLevel = qrcode.Highest
	}

	// Generate base QR code
	qr, err := qrcode.New(url, ecLevel)
	if err != nil {
		return nil, err
	}

	// Set colors
	fgColor, _ := parseHexColor(options.ForegroundColor)
	bgColor, _ := parseHexColor(options.BackgroundColor)
	qr.ForegroundColor = fgColor
	qr.BackgroundColor = bgColor

	// Generate image
	img := qr.Image(options.Size)

	// Apply customizations
	if options.DotStyle != "square" || options.CornerStyle != "square" {
		img = g.applyStyles(img, options)
	}

	// Add logo if specified
	if options.LogoURL != "" {
		img, err = g.addLogo(img, options.LogoURL, options.LogoSize)
		if err != nil {
			// Continue without logo if it fails
		}
	}

	// Add frame if specified
	if options.Frame != nil && options.Frame.Style != "none" {
		img = g.addFrame(img, options.Frame)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// applyStyles applies dot and corner styles
func (g *Generator) applyStyles(img image.Image, options QRCodeOptions) image.Image {
	bounds := img.Bounds()
	styled := image.NewRGBA(bounds)
	draw.Draw(styled, bounds, img, bounds.Min, draw.Src)

	// Apply dot style modifications
	// This is a simplified version - actual implementation would analyze
	// QR code matrix and redraw with styled dots

	return styled
}

// addLogo overlays a logo in the center of the QR code
func (g *Generator) addLogo(img image.Image, logoURL string, sizePercent int) (image.Image, error) {
	if sizePercent == 0 {
		sizePercent = 20 // Default 20% of QR size
	}

	// Fetch logo from URL
	logoData, err := g.storage.Get(context.Background(), logoURL)
	if err != nil {
		return img, err
	}

	// Decode logo
	logo, _, err := image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return img, err
	}

	// Calculate logo size
	qrSize := img.Bounds().Dx()
	logoSize := uint(qrSize * sizePercent / 100)

	// Resize logo
	resizedLogo := resize.Resize(logoSize, logoSize, logo, resize.Lanczos3)

	// Create result image
	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Src)

	// Calculate center position
	logoX := (qrSize - int(logoSize)) / 2
	logoY := (qrSize - int(logoSize)) / 2

	// Draw logo
	draw.Draw(result, image.Rect(logoX, logoY, logoX+int(logoSize), logoY+int(logoSize)),
		resizedLogo, image.Point{}, draw.Over)

	return result, nil
}

// addFrame adds a decorative frame around the QR code
func (g *Generator) addFrame(img image.Image, frame *FrameOptions) image.Image {
	bounds := img.Bounds()
	frameHeight := 60 // Height for text banner

	var newBounds image.Rectangle
	if frame.Position == "top" {
		newBounds = image.Rect(0, 0, bounds.Dx(), bounds.Dy()+frameHeight)
	} else {
		newBounds = image.Rect(0, 0, bounds.Dx(), bounds.Dy()+frameHeight)
	}

	result := image.NewRGBA(newBounds)

	// Fill background
	frameColor, _ := parseHexColor(frame.Color)
	draw.Draw(result, newBounds, &image.Uniform{frameColor}, image.Point{}, draw.Src)

	// Draw QR code
	qrY := 0
	if frame.Position == "top" {
		qrY = frameHeight
	}
	draw.Draw(result, image.Rect(0, qrY, bounds.Dx(), qrY+bounds.Dy()),
		img, image.Point{}, draw.Src)

	// Add text (simplified - actual implementation would use font rendering)
	// Would use golang.org/x/image/font for proper text rendering

	return result
}

// parseHexColor converts hex color string to color.Color
func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return color.Black, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}
```

### Customization Options

```go
// internal/qrcode/styles.go
package qrcode

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// DotStyle defines QR code module styles
type DotStyle string

const (
	DotStyleSquare  DotStyle = "square"
	DotStyleRounded DotStyle = "rounded"
	DotStyleDots    DotStyle = "dots"
	DotStyleDiamond DotStyle = "diamond"
)

// CornerStyle defines finder pattern styles
type CornerStyle string

const (
	CornerStyleSquare  CornerStyle = "square"
	CornerStyleRounded CornerStyle = "rounded"
	CornerStyleCircle  CornerStyle = "circle"
)

// StyleApplier applies visual styles to QR codes
type StyleApplier struct{}

// NewStyleApplier creates a new style applier
func NewStyleApplier() *StyleApplier {
	return &StyleApplier{}
}

// ApplyDotStyle redraws QR modules with the specified style
func (sa *StyleApplier) ApplyDotStyle(
	img image.Image,
	matrix [][]bool,
	dotStyle DotStyle,
	fgColor color.Color,
	moduleSize int,
) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Copy background
	draw.Draw(result, bounds, img, bounds.Min, draw.Src)

	for y, row := range matrix {
		for x, filled := range row {
			if !filled {
				continue
			}

			px := x * moduleSize
			py := y * moduleSize

			switch dotStyle {
			case DotStyleRounded:
				sa.drawRoundedRect(result, px, py, moduleSize, fgColor, moduleSize/4)
			case DotStyleDots:
				sa.drawCircle(result, px+moduleSize/2, py+moduleSize/2, moduleSize/2-1, fgColor)
			case DotStyleDiamond:
				sa.drawDiamond(result, px, py, moduleSize, fgColor)
			default:
				sa.drawSquare(result, px, py, moduleSize, fgColor)
			}
		}
	}

	return result
}

func (sa *StyleApplier) drawSquare(img *image.RGBA, x, y, size int, c color.Color) {
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			img.Set(x+dx, y+dy, c)
		}
	}
}

func (sa *StyleApplier) drawCircle(img *image.RGBA, cx, cy, radius int, c color.Color) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				img.Set(cx+x, cy+y, c)
			}
		}
	}
}

func (sa *StyleApplier) drawRoundedRect(img *image.RGBA, x, y, size int, c color.Color, radius int) {
	// Draw main rectangle
	sa.drawSquare(img, x+radius, y, size-2*radius, c)
	sa.drawSquare(img, x, y+radius, size, c)
	sa.drawSquare(img, x+radius, y+size-1, size-2*radius, c)

	// Draw rounded corners
	sa.drawCircle(img, x+radius, y+radius, radius, c)
	sa.drawCircle(img, x+size-radius-1, y+radius, radius, c)
	sa.drawCircle(img, x+radius, y+size-radius-1, radius, c)
	sa.drawCircle(img, x+size-radius-1, y+size-radius-1, radius, c)
}

func (sa *StyleApplier) drawDiamond(img *image.RGBA, x, y, size int, c color.Color) {
	center := size / 2
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			if abs(dx-center)+abs(dy-center) <= center {
				img.Set(x+dx, y+dy, c)
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Predefined style templates
var StyleTemplates = map[string]QRCodeOptions{
	"classic": {
		Size:            512,
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "square",
		CornerStyle:     "square",
		Margin:          4,
	},
	"modern": {
		Size:            512,
		ForegroundColor: "#1a1a2e",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "rounded",
		CornerStyle:     "rounded",
		Margin:          4,
	},
	"dots": {
		Size:            512,
		ForegroundColor: "#2d3436",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "dots",
		CornerStyle:     "rounded",
		Margin:          4,
	},
	"gradient_blue": {
		Size:            512,
		ForegroundColor: "#667eea",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "rounded",
		CornerStyle:     "rounded",
		Margin:          4,
	},
	"dark_mode": {
		Size:            512,
		ForegroundColor: "#FFFFFF",
		BackgroundColor: "#1a1a2e",
		DotStyle:        "rounded",
		CornerStyle:     "rounded",
		Margin:          4,
	},
}
```

---

## Batch Generation

```go
// internal/qrcode/batch.go
package qrcode

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/link-rift/link-rift/internal/storage"
)

// BatchRequest represents a batch QR code generation request
type BatchRequest struct {
	Links   []BatchLinkItem `json:"links"`
	Options QRCodeOptions   `json:"options"`
	Format  string          `json:"format"` // zip, individual
}

// BatchLinkItem represents a single link in a batch request
type BatchLinkItem struct {
	ID       string `json:"id"`
	ShortURL string `json:"short_url"`
	Name     string `json:"name"` // Used for filename
}

// BatchResult represents the result of batch generation
type BatchResult struct {
	TotalRequested int            `json:"total_requested"`
	TotalGenerated int            `json:"total_generated"`
	TotalFailed    int            `json:"total_failed"`
	ZipURL         string         `json:"zip_url,omitempty"`
	Results        []BatchItemResult `json:"results,omitempty"`
}

// BatchItemResult represents result for a single QR code
type BatchItemResult struct {
	LinkID   string `json:"link_id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BatchGenerator handles batch QR code generation
type BatchGenerator struct {
	generator *Generator
	storage   storage.ObjectStorage
	workers   int
}

// NewBatchGenerator creates a new batch generator
func NewBatchGenerator(gen *Generator, storage storage.ObjectStorage, workers int) *BatchGenerator {
	return &BatchGenerator{
		generator: gen,
		storage:   storage,
		workers:   workers,
	}
}

// Generate creates QR codes for multiple links
func (bg *BatchGenerator) Generate(ctx context.Context, workspaceID string, req BatchRequest) (*BatchResult, error) {
	result := &BatchResult{
		TotalRequested: len(req.Links),
		Results:        make([]BatchItemResult, 0, len(req.Links)),
	}

	// Create channels for work distribution
	jobs := make(chan BatchLinkItem, len(req.Links))
	results := make(chan BatchItemResult, len(req.Links))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bg.workers; i++ {
		wg.Add(1)
		go bg.worker(ctx, &wg, jobs, results, req.Options, workspaceID)
	}

	// Send jobs
	for _, link := range req.Links {
		jobs <- link
	}
	close(jobs)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var images []imageItem
	for res := range results {
		result.Results = append(result.Results, res)
		if res.Error == "" {
			result.TotalGenerated++
			images = append(images, imageItem{
				name:     res.Name,
				imageURL: res.ImageURL,
			})
		} else {
			result.TotalFailed++
		}
	}

	// Create ZIP if requested
	if req.Format == "zip" && len(images) > 0 {
		zipURL, err := bg.createZip(ctx, workspaceID, images)
		if err == nil {
			result.ZipURL = zipURL
		}
	}

	return result, nil
}

func (bg *BatchGenerator) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan BatchLinkItem,
	results chan<- BatchItemResult,
	options QRCodeOptions,
	workspaceID string,
) {
	defer wg.Done()

	for link := range jobs {
		result := BatchItemResult{
			LinkID: link.ID,
			Name:   link.Name,
		}

		// Generate QR code
		imgData, err := bg.generator.Generate(link.ShortURL, options)
		if err != nil {
			result.Error = err.Error()
			results <- result
			continue
		}

		// Upload to storage
		filename := fmt.Sprintf("qr/%s/%s.png", workspaceID, link.ID)
		url, err := bg.storage.Upload(ctx, filename, imgData, "image/png")
		if err != nil {
			result.Error = err.Error()
			results <- result
			continue
		}

		result.ImageURL = url
		results <- result
	}
}

type imageItem struct {
	name     string
	imageURL string
}

func (bg *BatchGenerator) createZip(ctx context.Context, workspaceID string, images []imageItem) (string, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, img := range images {
		// Fetch image
		data, err := bg.storage.Get(ctx, img.imageURL)
		if err != nil {
			continue
		}

		// Add to zip
		filename := fmt.Sprintf("%s.png", sanitizeFilename(img.name))
		fw, err := zw.Create(filename)
		if err != nil {
			continue
		}
		fw.Write(data)
	}

	if err := zw.Close(); err != nil {
		return "", err
	}

	// Upload zip
	zipFilename := fmt.Sprintf("qr/%s/batch-%d.zip", workspaceID, time.Now().Unix())
	url, err := bg.storage.Upload(ctx, zipFilename, buf.Bytes(), "application/zip")
	if err != nil {
		return "", err
	}

	return url, nil
}

func sanitizeFilename(name string) string {
	// Remove or replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
```

---

## QR Code Analytics

```go
// internal/qrcode/analytics.go
package qrcode

import (
	"context"
	"time"
)

// ScanEvent represents a QR code scan
type ScanEvent struct {
	QRCodeID    string    `json:"qr_code_id"`
	LinkID      string    `json:"link_id"`
	WorkspaceID string    `json:"workspace_id"`
	Timestamp   time.Time `json:"timestamp"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	DeviceType  string    `json:"device_type"`
	OS          string    `json:"os"`
	Browser     string    `json:"browser"`
}

// QRAnalytics contains analytics data for a QR code
type QRAnalytics struct {
	QRCodeID     string           `json:"qr_code_id"`
	TotalScans   int64            `json:"total_scans"`
	UniqueScans  int64            `json:"unique_scans"`
	ScansByDay   []DailyScans     `json:"scans_by_day"`
	TopCountries []CountryScans   `json:"top_countries"`
	DeviceBreakdown DeviceBreakdown `json:"device_breakdown"`
	ScanLocations []ScanLocation  `json:"scan_locations,omitempty"`
}

type DailyScans struct {
	Date  string `json:"date"`
	Scans int64  `json:"scans"`
}

type CountryScans struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Scans       int64  `json:"scans"`
}

type DeviceBreakdown struct {
	Mobile  int64 `json:"mobile"`
	Desktop int64 `json:"desktop"`
	Tablet  int64 `json:"tablet"`
}

type ScanLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Scans     int64   `json:"scans"`
}

// QRAnalyticsService provides QR code analytics
type QRAnalyticsService struct {
	clickhouse ClickHouseClient
}

// NewQRAnalyticsService creates a new analytics service
func NewQRAnalyticsService(ch ClickHouseClient) *QRAnalyticsService {
	return &QRAnalyticsService{clickhouse: ch}
}

// GetAnalytics retrieves analytics for a QR code
func (s *QRAnalyticsService) GetAnalytics(
	ctx context.Context,
	qrCodeID string,
	startDate, endDate time.Time,
) (*QRAnalytics, error) {
	analytics := &QRAnalytics{
		QRCodeID: qrCodeID,
	}

	// Get total and unique scans
	totals, err := s.getTotals(ctx, qrCodeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	analytics.TotalScans = totals.Total
	analytics.UniqueScans = totals.Unique

	// Get daily breakdown
	analytics.ScansByDay, err = s.getDailyScans(ctx, qrCodeID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get top countries
	analytics.TopCountries, err = s.getTopCountries(ctx, qrCodeID, startDate, endDate, 10)
	if err != nil {
		return nil, err
	}

	// Get device breakdown
	analytics.DeviceBreakdown, err = s.getDeviceBreakdown(ctx, qrCodeID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return analytics, nil
}

func (s *QRAnalyticsService) getTotals(
	ctx context.Context,
	qrCodeID string,
	startDate, endDate time.Time,
) (*struct{ Total, Unique int64 }, error) {
	query := `
		SELECT
			count() as total,
			uniqExact(ip_address) as unique
		FROM qr_scans
		WHERE qr_code_id = ?
		AND timestamp BETWEEN ? AND ?
	`

	var result struct{ Total, Unique int64 }
	err := s.clickhouse.QueryRow(ctx, query, qrCodeID, startDate, endDate).Scan(&result.Total, &result.Unique)
	return &result, err
}

func (s *QRAnalyticsService) getDailyScans(
	ctx context.Context,
	qrCodeID string,
	startDate, endDate time.Time,
) ([]DailyScans, error) {
	query := `
		SELECT
			toDate(timestamp) as date,
			count() as scans
		FROM qr_scans
		WHERE qr_code_id = ?
		AND timestamp BETWEEN ? AND ?
		GROUP BY date
		ORDER BY date
	`

	rows, err := s.clickhouse.Query(ctx, query, qrCodeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []DailyScans
	for rows.Next() {
		var ds DailyScans
		if err := rows.Scan(&ds.Date, &ds.Scans); err != nil {
			continue
		}
		results = append(results, ds)
	}

	return results, nil
}

func (s *QRAnalyticsService) getTopCountries(
	ctx context.Context,
	qrCodeID string,
	startDate, endDate time.Time,
	limit int,
) ([]CountryScans, error) {
	query := `
		SELECT
			country,
			country_code,
			count() as scans
		FROM qr_scans
		WHERE qr_code_id = ?
		AND timestamp BETWEEN ? AND ?
		GROUP BY country, country_code
		ORDER BY scans DESC
		LIMIT ?
	`

	rows, err := s.clickhouse.Query(ctx, query, qrCodeID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CountryScans
	for rows.Next() {
		var cs CountryScans
		if err := rows.Scan(&cs.Country, &cs.CountryCode, &cs.Scans); err != nil {
			continue
		}
		results = append(results, cs)
	}

	return results, nil
}

func (s *QRAnalyticsService) getDeviceBreakdown(
	ctx context.Context,
	qrCodeID string,
	startDate, endDate time.Time,
) (DeviceBreakdown, error) {
	query := `
		SELECT
			countIf(device_type = 'mobile') as mobile,
			countIf(device_type = 'desktop') as desktop,
			countIf(device_type = 'tablet') as tablet
		FROM qr_scans
		WHERE qr_code_id = ?
		AND timestamp BETWEEN ? AND ?
	`

	var breakdown DeviceBreakdown
	err := s.clickhouse.QueryRow(ctx, query, qrCodeID, startDate, endDate).
		Scan(&breakdown.Mobile, &breakdown.Desktop, &breakdown.Tablet)

	return breakdown, err
}
```

---

## API Endpoints

```go
// internal/api/handlers/qrcode.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/qrcode"
)

// QRCodeHandler handles QR code API requests
type QRCodeHandler struct {
	service *qrcode.Service
}

// NewQRCodeHandler creates a new QR code handler
func NewQRCodeHandler(service *qrcode.Service) *QRCodeHandler {
	return &QRCodeHandler{service: service}
}

// RegisterRoutes registers QR code routes
func (h *QRCodeHandler) RegisterRoutes(app *fiber.App) {
	qr := app.Group("/api/v1/qr")

	qr.Get("/", h.ListQRCodes)
	qr.Post("/", h.CreateQRCode)
	qr.Get("/:id", h.GetQRCode)
	qr.Put("/:id", h.UpdateQRCode)
	qr.Delete("/:id", h.DeleteQRCode)
	qr.Get("/:id/image", h.GetQRCodeImage)
	qr.Get("/:id/analytics", h.GetQRCodeAnalytics)
	qr.Post("/batch", h.BatchGenerate)
	qr.Get("/templates", h.GetStyleTemplates)
}

// CreateQRCodeRequest represents QR code creation request
type CreateQRCodeRequest struct {
	LinkID  string               `json:"link_id" validate:"required"`
	Type    qrcode.QRCodeType    `json:"type" validate:"required,oneof=dynamic static"`
	Options qrcode.QRCodeOptions `json:"options"`
}

// CreateQRCode creates a new QR code
// @Summary Create a QR code
// @Tags QR Codes
// @Accept json
// @Produce json
// @Param body body CreateQRCodeRequest true "QR code details"
// @Success 201 {object} QRCodeResponse
// @Router /api/v1/qr [post]
func (h *QRCodeHandler) CreateQRCode(c *fiber.Ctx) error {
	var req CreateQRCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	qr, err := h.service.Create(c.Context(), workspaceID, req.LinkID, req.Type, req.Options)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create QR code",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(qr)
}

// GetQRCodeImage returns the QR code image
// @Summary Get QR code image
// @Tags QR Codes
// @Produce image/png
// @Param id path string true "QR Code ID"
// @Param size query int false "Image size"
// @Param format query string false "Image format (png, svg)"
// @Success 200 {file} binary
// @Router /api/v1/qr/{id}/image [get]
func (h *QRCodeHandler) GetQRCodeImage(c *fiber.Ctx) error {
	qrID := c.Params("id")
	size := c.QueryInt("size", 512)
	format := c.Query("format", "png")

	imageData, contentType, err := h.service.GetImage(c.Context(), qrID, size, format)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "QR code not found",
		})
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=86400")
	return c.Send(imageData)
}

// BatchGenerateRequest represents batch generation request
type BatchGenerateRequest struct {
	LinkIDs []string             `json:"link_ids" validate:"required,min=1,max=100"`
	Options qrcode.QRCodeOptions `json:"options"`
	Format  string               `json:"format" validate:"oneof=zip individual"`
}

// BatchGenerate generates QR codes for multiple links
// @Summary Batch generate QR codes
// @Tags QR Codes
// @Accept json
// @Produce json
// @Param body body BatchGenerateRequest true "Batch request"
// @Success 200 {object} BatchResult
// @Router /api/v1/qr/batch [post]
func (h *QRCodeHandler) BatchGenerate(c *fiber.Ctx) error {
	var req BatchGenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	result, err := h.service.BatchGenerate(c.Context(), workspaceID, req.LinkIDs, req.Options, req.Format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Batch generation failed",
		})
	}

	return c.JSON(result)
}

// GetStyleTemplates returns available style templates
// @Summary Get QR code style templates
// @Tags QR Codes
// @Produce json
// @Success 200 {object} map[string]QRCodeOptions
// @Router /api/v1/qr/templates [get]
func (h *QRCodeHandler) GetStyleTemplates(c *fiber.Ctx) error {
	return c.JSON(qrcode.StyleTemplates)
}

// GetQRCodeAnalytics returns analytics for a QR code
// @Summary Get QR code analytics
// @Tags QR Codes
// @Produce json
// @Param id path string true "QR Code ID"
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} QRAnalytics
// @Router /api/v1/qr/{id}/analytics [get]
func (h *QRCodeHandler) GetQRCodeAnalytics(c *fiber.Ctx) error {
	qrID := c.Params("id")

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	analytics, err := h.service.GetAnalytics(c.Context(), qrID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch analytics",
		})
	}

	return c.JSON(analytics)
}
```

---

## React Components

### QR Code Generator Component

```typescript
// src/components/qrcode/QRCodeGenerator.tsx
import React, { useState, useEffect, useMemo } from 'react';
import { useMutation } from '@tanstack/react-query';
import { qrCodeApi, QRCodeOptions, StyleTemplate } from '@/api/qrcode';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Slider } from '@/components/ui/slider';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Download, Copy, RefreshCw } from 'lucide-react';

interface QRCodeGeneratorProps {
  linkId: string;
  shortUrl: string;
  onGenerated?: (qrCodeId: string) => void;
}

export const QRCodeGenerator: React.FC<QRCodeGeneratorProps> = ({
  linkId,
  shortUrl,
  onGenerated,
}) => {
  const [options, setOptions] = useState<QRCodeOptions>({
    size: 512,
    error_correction: 'M',
    foreground_color: '#000000',
    background_color: '#FFFFFF',
    dot_style: 'square',
    corner_style: 'square',
    margin: 4,
  });

  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [selectedTemplate, setSelectedTemplate] = useState<string>('classic');

  // Generate preview URL
  const previewParams = useMemo(() => {
    const params = new URLSearchParams({
      url: shortUrl,
      size: options.size.toString(),
      fg: options.foreground_color.replace('#', ''),
      bg: options.background_color.replace('#', ''),
      dot: options.dot_style,
      corner: options.corner_style,
    });
    return params.toString();
  }, [options, shortUrl]);

  useEffect(() => {
    setPreviewUrl(`/api/v1/qr/preview?${previewParams}`);
  }, [previewParams]);

  const createMutation = useMutation({
    mutationFn: () => qrCodeApi.create({
      link_id: linkId,
      type: 'dynamic',
      options,
    }),
    onSuccess: (data) => {
      onGenerated?.(data.id);
    },
  });

  const handleTemplateSelect = (templateName: string) => {
    setSelectedTemplate(templateName);
    const template = StyleTemplates[templateName];
    if (template) {
      setOptions(prev => ({ ...prev, ...template }));
    }
  };

  const handleDownload = async (format: 'png' | 'svg') => {
    const response = await qrCodeApi.downloadPreview(shortUrl, options, format);
    const blob = new Blob([response], { type: format === 'png' ? 'image/png' : 'image/svg+xml' });
    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = `qrcode.${format}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Preview */}
      <Card>
        <CardHeader>
          <CardTitle>Preview</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center space-y-4">
          <div className="border rounded-lg p-4 bg-white">
            {previewUrl && (
              <img
                src={previewUrl}
                alt="QR Code Preview"
                className="w-64 h-64"
              />
            )}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => handleDownload('png')}>
              <Download className="w-4 h-4 mr-2" />
              PNG
            </Button>
            <Button variant="outline" onClick={() => handleDownload('svg')}>
              <Download className="w-4 h-4 mr-2" />
              SVG
            </Button>
          </div>
          <Button
            onClick={() => createMutation.mutate()}
            disabled={createMutation.isPending}
            className="w-full"
          >
            {createMutation.isPending ? (
              <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
            ) : null}
            Save QR Code
          </Button>
        </CardContent>
      </Card>

      {/* Customization */}
      <Card>
        <CardHeader>
          <CardTitle>Customize</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="templates">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="templates">Templates</TabsTrigger>
              <TabsTrigger value="colors">Colors</TabsTrigger>
              <TabsTrigger value="style">Style</TabsTrigger>
            </TabsList>

            <TabsContent value="templates" className="space-y-4">
              <div className="grid grid-cols-3 gap-2">
                {Object.keys(StyleTemplates).map((name) => (
                  <button
                    key={name}
                    onClick={() => handleTemplateSelect(name)}
                    className={`p-2 border rounded-lg text-sm capitalize ${
                      selectedTemplate === name ? 'border-primary bg-primary/10' : ''
                    }`}
                  >
                    {name.replace('_', ' ')}
                  </button>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="colors" className="space-y-4">
              <div className="space-y-2">
                <Label>Foreground Color</Label>
                <div className="flex gap-2">
                  <input
                    type="color"
                    value={options.foreground_color}
                    onChange={(e) => setOptions(prev => ({
                      ...prev,
                      foreground_color: e.target.value,
                    }))}
                    className="w-12 h-10 rounded cursor-pointer"
                  />
                  <Input
                    value={options.foreground_color}
                    onChange={(e) => setOptions(prev => ({
                      ...prev,
                      foreground_color: e.target.value,
                    }))}
                    className="flex-1"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label>Background Color</Label>
                <div className="flex gap-2">
                  <input
                    type="color"
                    value={options.background_color}
                    onChange={(e) => setOptions(prev => ({
                      ...prev,
                      background_color: e.target.value,
                    }))}
                    className="w-12 h-10 rounded cursor-pointer"
                  />
                  <Input
                    value={options.background_color}
                    onChange={(e) => setOptions(prev => ({
                      ...prev,
                      background_color: e.target.value,
                    }))}
                    className="flex-1"
                  />
                </div>
              </div>
            </TabsContent>

            <TabsContent value="style" className="space-y-4">
              <div className="space-y-2">
                <Label>Size: {options.size}px</Label>
                <Slider
                  value={[options.size]}
                  onValueChange={([value]) => setOptions(prev => ({
                    ...prev,
                    size: value,
                  }))}
                  min={128}
                  max={1024}
                  step={64}
                />
              </div>

              <div className="space-y-2">
                <Label>Dot Style</Label>
                <RadioGroup
                  value={options.dot_style}
                  onValueChange={(value) => setOptions(prev => ({
                    ...prev,
                    dot_style: value,
                  }))}
                >
                  <div className="flex gap-4">
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="square" id="dot-square" />
                      <Label htmlFor="dot-square">Square</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="rounded" id="dot-rounded" />
                      <Label htmlFor="dot-rounded">Rounded</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="dots" id="dot-dots" />
                      <Label htmlFor="dot-dots">Dots</Label>
                    </div>
                  </div>
                </RadioGroup>
              </div>

              <div className="space-y-2">
                <Label>Corner Style</Label>
                <RadioGroup
                  value={options.corner_style}
                  onValueChange={(value) => setOptions(prev => ({
                    ...prev,
                    corner_style: value,
                  }))}
                >
                  <div className="flex gap-4">
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="square" id="corner-square" />
                      <Label htmlFor="corner-square">Square</Label>
                    </div>
                    <div className="flex items-center space-x-2">
                      <RadioGroupItem value="rounded" id="corner-rounded" />
                      <Label htmlFor="corner-rounded">Rounded</Label>
                    </div>
                  </div>
                </RadioGroup>
              </div>

              <div className="space-y-2">
                <Label>Error Correction</Label>
                <RadioGroup
                  value={options.error_correction}
                  onValueChange={(value) => setOptions(prev => ({
                    ...prev,
                    error_correction: value,
                  }))}
                >
                  <div className="flex gap-4">
                    {['L', 'M', 'Q', 'H'].map((level) => (
                      <div key={level} className="flex items-center space-x-2">
                        <RadioGroupItem value={level} id={`ec-${level}`} />
                        <Label htmlFor={`ec-${level}`}>{level}</Label>
                      </div>
                    ))}
                  </div>
                </RadioGroup>
              </div>

              <div className="space-y-2">
                <Label>Margin: {options.margin} modules</Label>
                <Slider
                  value={[options.margin]}
                  onValueChange={([value]) => setOptions(prev => ({
                    ...prev,
                    margin: value,
                  }))}
                  min={0}
                  max={10}
                  step={1}
                />
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
};

const StyleTemplates: Record<string, Partial<QRCodeOptions>> = {
  classic: {
    foreground_color: '#000000',
    background_color: '#FFFFFF',
    dot_style: 'square',
    corner_style: 'square',
  },
  modern: {
    foreground_color: '#1a1a2e',
    background_color: '#FFFFFF',
    dot_style: 'rounded',
    corner_style: 'rounded',
  },
  dots: {
    foreground_color: '#2d3436',
    background_color: '#FFFFFF',
    dot_style: 'dots',
    corner_style: 'rounded',
  },
  gradient_blue: {
    foreground_color: '#667eea',
    background_color: '#FFFFFF',
    dot_style: 'rounded',
    corner_style: 'rounded',
  },
  dark_mode: {
    foreground_color: '#FFFFFF',
    background_color: '#1a1a2e',
    dot_style: 'rounded',
    corner_style: 'rounded',
  },
};
```

### QR Code API Client

```typescript
// src/api/qrcode.ts
import { apiClient } from './client';

export interface QRCodeOptions {
  size: number;
  error_correction: string;
  foreground_color: string;
  background_color: string;
  dot_style: string;
  corner_style: string;
  margin: number;
  logo_url?: string;
  logo_size?: number;
  frame?: {
    style: string;
    color: string;
    text: string;
    text_color: string;
    position: string;
  };
}

export interface QRCode {
  id: string;
  link_id: string;
  type: 'dynamic' | 'static';
  encoded_url: string;
  options: QRCodeOptions;
  image_url: string;
  total_scans: number;
  created_at: string;
}

export interface QRAnalytics {
  qr_code_id: string;
  total_scans: number;
  unique_scans: number;
  scans_by_day: Array<{ date: string; scans: number }>;
  top_countries: Array<{ country: string; country_code: string; scans: number }>;
  device_breakdown: {
    mobile: number;
    desktop: number;
    tablet: number;
  };
}

export interface BatchResult {
  total_requested: number;
  total_generated: number;
  total_failed: number;
  zip_url?: string;
  results: Array<{
    link_id: string;
    name: string;
    image_url?: string;
    error?: string;
  }>;
}

export const qrCodeApi = {
  list: async (): Promise<QRCode[]> => {
    const response = await apiClient.get<QRCode[]>('/api/v1/qr');
    return response.data;
  },

  create: async (data: {
    link_id: string;
    type: 'dynamic' | 'static';
    options: QRCodeOptions;
  }): Promise<QRCode> => {
    const response = await apiClient.post<QRCode>('/api/v1/qr', data);
    return response.data;
  },

  get: async (id: string): Promise<QRCode> => {
    const response = await apiClient.get<QRCode>(`/api/v1/qr/${id}`);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/qr/${id}`);
  },

  getImage: async (id: string, size?: number, format?: string): Promise<Blob> => {
    const params = new URLSearchParams();
    if (size) params.set('size', size.toString());
    if (format) params.set('format', format);

    const response = await apiClient.get(`/api/v1/qr/${id}/image?${params}`, {
      responseType: 'blob',
    });
    return response.data;
  },

  getAnalytics: async (
    id: string,
    startDate?: string,
    endDate?: string
  ): Promise<QRAnalytics> => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);

    const response = await apiClient.get<QRAnalytics>(
      `/api/v1/qr/${id}/analytics?${params}`
    );
    return response.data;
  },

  batchGenerate: async (data: {
    link_ids: string[];
    options: QRCodeOptions;
    format: 'zip' | 'individual';
  }): Promise<BatchResult> => {
    const response = await apiClient.post<BatchResult>('/api/v1/qr/batch', data);
    return response.data;
  },

  downloadPreview: async (
    url: string,
    options: QRCodeOptions,
    format: 'png' | 'svg'
  ): Promise<ArrayBuffer> => {
    const params = new URLSearchParams({
      url,
      size: options.size.toString(),
      fg: options.foreground_color.replace('#', ''),
      bg: options.background_color.replace('#', ''),
      dot: options.dot_style,
      corner: options.corner_style,
      format,
    });

    const response = await apiClient.get(`/api/v1/qr/preview?${params}`, {
      responseType: 'arraybuffer',
    });
    return response.data;
  },

  getTemplates: async (): Promise<Record<string, QRCodeOptions>> => {
    const response = await apiClient.get('/api/v1/qr/templates');
    return response.data;
  },
};
```
