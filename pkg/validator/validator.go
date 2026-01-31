package validator

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	shortCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,48}[a-zA-Z0-9]$|^[a-zA-Z0-9]{1,2}$`)
	slugRegex      = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,98}[a-z0-9]$|^[a-z0-9]$`)
	emailRegex     = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func IsValidURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func IsValidShortCode(s string) bool {
	if len(s) < 1 || len(s) > 50 {
		return false
	}
	return shortCodeRegex.MatchString(s)
}

func IsValidSlug(s string) bool {
	if len(s) < 1 || len(s) > 100 {
		return false
	}
	return slugRegex.MatchString(s)
}

func IsValidEmail(s string) bool {
	if len(s) > 254 {
		return false
	}
	return emailRegex.MatchString(s)
}

func NormalizeURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	s = strings.TrimRight(s, "/")
	return s
}
