package service

import (
	"context"
	"time"
)

// SSLProvider abstracts SSL certificate provisioning.
type SSLProvider interface {
	ProvisionSSL(ctx context.Context, domain string) (string, error)
	CheckSSLStatus(ctx context.Context, domain string) (string, *time.Time, error)
	RemoveSSL(ctx context.Context, domain string) error
}

// MockSSLProvider is a no-op implementation that returns "active" immediately.
// Replace with a Cloudflare or Let's Encrypt implementation in production.
type MockSSLProvider struct{}

func NewMockSSLProvider() SSLProvider {
	return &MockSSLProvider{}
}

func (m *MockSSLProvider) ProvisionSSL(_ context.Context, _ string) (string, error) {
	return "active", nil
}

func (m *MockSSLProvider) CheckSSLStatus(_ context.Context, _ string) (string, *time.Time, error) {
	expires := time.Now().Add(365 * 24 * time.Hour)
	return "active", &expires, nil
}

func (m *MockSSLProvider) RemoveSSL(_ context.Context, _ string) error {
	return nil
}
