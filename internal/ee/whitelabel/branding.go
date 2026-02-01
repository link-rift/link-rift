package whitelabel

import (
	"regexp"
	"strings"
)

// hexColorRegex matches 3 or 6 digit hex colors (e.g. #fff, #FF0000).
var hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// ValidateHexColor returns true if the string is a valid hex color.
func ValidateHexColor(color string) bool {
	if color == "" {
		return true // empty is allowed (means "no override")
	}
	return hexColorRegex.MatchString(color)
}

// dangerousPatterns are CSS patterns that could enable XSS or external resource loading.
var dangerousPatterns = []string{
	"javascript:",
	"expression(",
	"url(",
	"@import",
	"behavior:",
	"-moz-binding",
	"<script",
	"</script",
}

// SanitizeCSS performs basic sanitization of custom CSS.
// It rejects CSS containing potentially dangerous patterns.
func SanitizeCSS(css string) (string, bool) {
	if css == "" {
		return "", true
	}

	lower := strings.ToLower(css)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			return "", false
		}
	}

	// Limit CSS size to 50KB
	if len(css) > 50*1024 {
		return "", false
	}

	return css, true
}
