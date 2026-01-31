package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"
)

// Signer creates signed licenses for testing. Never used in production.
type Signer struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// GenerateKeyPair creates a test keypair.
func GenerateKeyPair(t *testing.T) *Signer {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	return &Signer{PrivateKey: priv, PublicKey: pub}
}

// PublicKeyPEM returns the PEM-encoded public key.
func (s *Signer) PublicKeyPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: s.PublicKey,
	})
}

// Sign creates a signed license from a License struct.
func (s *Signer) Sign(t *testing.T, lic *License) *SignedLicense {
	t.Helper()
	licBytes, err := json.Marshal(lic)
	if err != nil {
		t.Fatalf("marshal license: %v", err)
	}

	sig := ed25519.Sign(s.PrivateKey, licBytes)

	return &SignedLicense{
		License:   base64.StdEncoding.EncodeToString(licBytes),
		Signature: base64.StdEncoding.EncodeToString(sig),
		Version:   1,
	}
}

// SignToString creates a base64-encoded license key string.
func (s *Signer) SignToString(t *testing.T, lic *License) string {
	t.Helper()
	signed := s.Sign(t, lic)
	b, err := json.Marshal(signed)
	if err != nil {
		t.Fatalf("marshal signed license: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// newTestLicense creates a valid license for testing.
func newTestLicense() *License {
	return &License{
		ID:           "test-license-001",
		CustomerID:   "cust-001",
		CustomerName: "Test Corp",
		Email:        "admin@test.com",
		Type:         LicenseTypeSubscription,
		Tier:         TierPro,
		IssuedAt:     time.Now().Add(-1 * time.Hour),
		ExpiresAt:    time.Now().Add(365 * 24 * time.Hour),
		Features:     []Feature{FeatureCustomDomains, FeatureAdvancedAnalytics},
		Limits: Limits{
			MaxUsers:             10,
			MaxDomains:           5,
			MaxLinksPerMonth:     10000,
			MaxClicksPerMonth:    1000000,
			MaxWorkspaces:        5,
			MaxAPIRequestsPerMin: 120,
		},
	}
}
