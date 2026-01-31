package license

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
)

//go:embed keys/public.pem
var embeddedPublicKey []byte

var (
	ErrInvalidSignature     = errors.New("invalid license signature")
	ErrLicenseExpired       = errors.New("license has expired")
	ErrLicenseNotYetValid   = errors.New("license is not yet valid")
	ErrInvalidLicenseFormat = errors.New("invalid license format")
)

// Verifier validates license signatures using an Ed25519 public key.
type Verifier struct {
	publicKey ed25519.PublicKey
}

// NewVerifier creates a verifier with the embedded public key.
func NewVerifier() (*Verifier, error) {
	return NewVerifierWithKey(embeddedPublicKey)
}

// NewVerifierWithKey creates a verifier with a PEM-encoded public key.
func NewVerifierWithKey(pubKeyPEM []byte) (*Verifier, error) {
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block: %w", ErrInvalidLicenseFormat)
	}

	if len(block.Bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %w", ErrInvalidLicenseFormat)
	}

	return &Verifier{
		publicKey: ed25519.PublicKey(block.Bytes),
	}, nil
}

// Verify validates a SignedLicense and returns the decoded License.
func (v *Verifier) Verify(signed *SignedLicense) (*License, error) {
	if signed == nil {
		return nil, ErrInvalidLicenseFormat
	}

	licenseBytes, err := base64.StdEncoding.DecodeString(signed.License)
	if err != nil {
		return nil, fmt.Errorf("decoding license data: %w", ErrInvalidLicenseFormat)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signed.Signature)
	if err != nil {
		return nil, fmt.Errorf("decoding signature: %w", ErrInvalidLicenseFormat)
	}

	if !ed25519.Verify(v.publicKey, licenseBytes, sigBytes) {
		return nil, ErrInvalidSignature
	}

	var lic License
	if err := json.Unmarshal(licenseBytes, &lic); err != nil {
		return nil, fmt.Errorf("unmarshalling license: %w", ErrInvalidLicenseFormat)
	}

	if !lic.IssuedAt.IsZero() && !lic.ExpiresAt.IsZero() && lic.IssuedAt.After(lic.ExpiresAt) {
		return nil, ErrInvalidLicenseFormat
	}

	if lic.IsExpired() {
		return &lic, ErrLicenseExpired
	}

	if !lic.IsValid() {
		return &lic, ErrLicenseNotYetValid
	}

	return &lic, nil
}

// VerifyString parses a base64-encoded JSON signed license key string.
func (v *Verifier) VerifyString(licenseKey string) (*License, error) {
	decoded, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, fmt.Errorf("decoding license key: %w", ErrInvalidLicenseFormat)
	}

	var signed SignedLicense
	if err := json.Unmarshal(decoded, &signed); err != nil {
		return nil, fmt.Errorf("parsing license key: %w", ErrInvalidLicenseFormat)
	}

	return v.Verify(&signed)
}
