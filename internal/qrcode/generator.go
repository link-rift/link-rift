package qrcode

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"

	goqrcode "github.com/skip2/go-qrcode"

	"github.com/link-rift/link-rift/pkg/storage"
)

// Options configures QR code generation.
type Options struct {
	Size            int
	ErrorCorrection string // L, M, Q, H
	ForegroundColor string // hex like #000000
	BackgroundColor string // hex like #FFFFFF
	LogoURL         string
	DotStyle        string
	CornerStyle     string
	Margin          int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size:            512,
		ErrorCorrection: "M",
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "square",
		CornerStyle:     "square",
		Margin:          4,
	}
}

// Generator generates QR code images.
type Generator struct {
	storage storage.ObjectStorage
}

// NewGenerator creates a new QR code generator.
func NewGenerator(store storage.ObjectStorage) *Generator {
	return &Generator{storage: store}
}

// ecLevel converts a string error correction level to the library type.
func ecLevel(level string) goqrcode.RecoveryLevel {
	switch strings.ToUpper(level) {
	case "L":
		return goqrcode.Low
	case "M":
		return goqrcode.Medium
	case "Q":
		return goqrcode.High
	case "H":
		return goqrcode.Highest
	default:
		return goqrcode.Medium
	}
}

// encodeQRMatrix generates a valid QR code boolean matrix using the go-qrcode library.
func encodeQRMatrix(data string, ecLevelStr string) ([][]bool, error) {
	qr, err := goqrcode.New(data, ecLevel(ecLevelStr))
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR data: %w", err)
	}
	return qr.Bitmap(), nil
}

// Generate creates a PNG QR code image and returns the bytes.
func (g *Generator) Generate(url string, opts Options) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 512
	}
	if opts.Size > 2048 {
		opts.Size = 2048
	}

	fg := parseHexColorWithDefault(opts.ForegroundColor, color.Black)
	bg := parseHexColorWithDefault(opts.BackgroundColor, color.White)

	matrix, err := encodeQRMatrix(url, opts.ErrorCorrection)
	if err != nil {
		return nil, err
	}

	margin := opts.Margin
	if margin < 0 {
		margin = 4
	}

	moduleCount := len(matrix)
	totalModules := moduleCount + 2*margin

	// Calculate module pixel size
	moduleSize := opts.Size / totalModules
	if moduleSize < 1 {
		moduleSize = 1
	}
	imgSize := totalModules * moduleSize

	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	// Fill background
	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			img.Set(x, y, bg)
		}
	}

	// Draw modules
	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			if matrix[row][col] {
				px := (col + margin) * moduleSize
				py := (row + margin) * moduleSize
				for dy := 0; dy < moduleSize; dy++ {
					for dx := 0; dx < moduleSize; dx++ {
						img.Set(px+dx, py+dy, fg)
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateSVG creates an SVG QR code and returns the bytes.
func (g *Generator) GenerateSVG(url string, opts Options) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 512
	}

	matrix, err := encodeQRMatrix(url, opts.ErrorCorrection)
	if err != nil {
		return nil, err
	}

	margin := opts.Margin
	if margin < 0 {
		margin = 4
	}

	fgHex := opts.ForegroundColor
	if fgHex == "" {
		fgHex = "#000000"
	}
	bgHex := opts.BackgroundColor
	if bgHex == "" {
		bgHex = "#FFFFFF"
	}

	moduleCount := len(matrix)
	totalModules := moduleCount + 2*margin
	moduleSize := opts.Size / totalModules
	if moduleSize < 1 {
		moduleSize = 1
	}
	totalSize := totalModules * moduleSize

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, totalSize, totalSize, opts.Size, opts.Size)
	fmt.Fprintf(&buf, `<rect width="%d" height="%d" fill="%s"/>`, totalSize, totalSize, bgHex)

	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			if matrix[row][col] {
				px := (col + margin) * moduleSize
				py := (row + margin) * moduleSize
				fmt.Fprintf(&buf, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
					px, py, moduleSize, moduleSize, fgHex)
			}
		}
	}

	buf.WriteString(`</svg>`)
	return buf.Bytes(), nil
}

// GenerateAndUpload generates a QR code and uploads it to storage.
func (g *Generator) GenerateAndUpload(ctx context.Context, url, storageKey string, opts Options) (pngURL string, err error) {
	pngBytes, err := g.Generate(url, opts)
	if err != nil {
		return "", err
	}

	pngURL, err = g.storage.Upload(ctx, storageKey, pngBytes, "image/png")
	if err != nil {
		return "", fmt.Errorf("failed to upload QR code: %w", err)
	}

	return pngURL, nil
}

// GenerateDataURI generates a QR code and returns it as a data URI string.
func (g *Generator) GenerateDataURI(url string, opts Options) (string, error) {
	pngBytes, err := g.Generate(url, opts)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	return "data:image/png;base64," + encoded, nil
}

func parseHexColorWithDefault(hex string, defaultColor color.Color) color.Color {
	c, err := ParseHexColor(hex)
	if err != nil {
		return defaultColor
	}
	return c
}

// ParseHexColor parses a hex color string like "#FF0000" into a color.Color.
func ParseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	green, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.RGBA{R: uint8(r), G: uint8(green), B: uint8(b), A: 255}, nil
}
