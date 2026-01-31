package validator

import (
	"strings"
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example.com/path?q=1", true},
		{"https://sub.example.com:8080", true},
		{"", false},
		{"not-a-url", false},
		{"ftp://example.com", false},
		{"://missing-scheme", false},
		{"https://", false},
	}
	for _, tt := range tests {
		got := IsValidURL(tt.input)
		if got != tt.want {
			t.Errorf("IsValidURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsValidShortCode(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc123", true},
		{"a", true},
		{"ab", true},
		{"my-code", true},
		{"ABC-xyz-123", true},
		{"a-b", true},
		{"", false},
		{"-start", false},
		{"end-", false},
		{strings.Repeat("a", 51), false},
		{"has spaces", false},
		{"has_underscore", false},
	}
	for _, tt := range tests {
		got := IsValidShortCode(tt.input)
		if got != tt.want {
			t.Errorf("IsValidShortCode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"my-workspace", true},
		{"a", true},
		{"test123", true},
		{"my-cool-workspace", true},
		{"", false},
		{"My-Slug", false},
		{"-start", false},
		{"end-", false},
		{strings.Repeat("a", 101), false},
	}
	for _, tt := range tests {
		got := IsValidSlug(tt.input)
		if got != tt.want {
			t.Errorf("IsValidSlug(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"user@example.com", true},
		{"user.name+tag@domain.co", true},
		{"user@sub.domain.com", true},
		{"", false},
		{"not-an-email", false},
		{"@no-user.com", false},
		{"no-domain@", false},
		{"user@.com", false},
	}
	for _, tt := range tests {
		got := IsValidEmail(tt.input)
		if got != tt.want {
			t.Errorf("IsValidEmail(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"example.com", "https://example.com"},
		{"https://example.com/", "https://example.com"},
		{"http://example.com///", "http://example.com"},
		{"  https://example.com  ", "https://example.com"},
		{"", ""},
		{"https://example.com/path", "https://example.com/path"},
	}
	for _, tt := range tests {
		got := NormalizeURL(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
