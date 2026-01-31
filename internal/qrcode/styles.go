package qrcode

// DotStyle constants
const (
	DotStyleSquare  = "square"
	DotStyleRounded = "rounded"
	DotStyleDots    = "dots"
)

// CornerStyle constants
const (
	CornerStyleSquare  = "square"
	CornerStyleRounded = "rounded"
)

// StyleTemplate is a predefined QR code style configuration.
type StyleTemplate struct {
	Name            string `json:"name"`
	ForegroundColor string `json:"foreground_color"`
	BackgroundColor string `json:"background_color"`
	DotStyle        string `json:"dot_style"`
	CornerStyle     string `json:"corner_style"`
	ErrorCorrection string `json:"error_correction"`
}

// StyleTemplates contains predefined QR code styles.
var StyleTemplates = map[string]StyleTemplate{
	"classic": {
		Name:            "Classic",
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		DotStyle:        DotStyleSquare,
		CornerStyle:     CornerStyleSquare,
		ErrorCorrection: "M",
	},
	"modern": {
		Name:            "Modern",
		ForegroundColor: "#1a1a2e",
		BackgroundColor: "#f5f5f5",
		DotStyle:        DotStyleRounded,
		CornerStyle:     CornerStyleRounded,
		ErrorCorrection: "Q",
	},
	"dots": {
		Name:            "Dots",
		ForegroundColor: "#2d3436",
		BackgroundColor: "#ffffff",
		DotStyle:        DotStyleDots,
		CornerStyle:     CornerStyleRounded,
		ErrorCorrection: "H",
	},
	"dark_mode": {
		Name:            "Dark Mode",
		ForegroundColor: "#e0e0e0",
		BackgroundColor: "#1a1a1a",
		DotStyle:        DotStyleSquare,
		CornerStyle:     CornerStyleSquare,
		ErrorCorrection: "M",
	},
}

// ValidDotStyles returns all valid dot style values.
func ValidDotStyles() []string {
	return []string{DotStyleSquare, DotStyleRounded, DotStyleDots}
}

// ValidCornerStyles returns all valid corner style values.
func ValidCornerStyles() []string {
	return []string{CornerStyleSquare, CornerStyleRounded}
}

// IsValidDotStyle checks if the given dot style is valid.
func IsValidDotStyle(s string) bool {
	for _, v := range ValidDotStyles() {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidCornerStyle checks if the given corner style is valid.
func IsValidCornerStyle(s string) bool {
	for _, v := range ValidCornerStyles() {
		if v == s {
			return true
		}
	}
	return false
}
