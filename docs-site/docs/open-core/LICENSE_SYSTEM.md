# Linkrift License Verification System

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [License Architecture](#license-architecture)
  - [License Structure](#license-structure)
  - [License Types](#license-types)
- [Ed25519 Cryptographic Signing](#ed25519-cryptographic-signing)
  - [Key Generation](#key-generation)
  - [License Signing](#license-signing)
  - [Signature Verification](#signature-verification)
- [License Verification Flow](#license-verification-flow)
  - [Startup Verification](#startup-verification)
  - [Runtime Verification](#runtime-verification)
- [Feature Flag Checking](#feature-flag-checking)
  - [Feature Registry](#feature-registry)
  - [Feature Checking Code](#feature-checking-code)
- [Middleware for Feature Gating](#middleware-for-feature-gating)
  - [HTTP Middleware](#http-middleware)
  - [gRPC Interceptors](#grpc-interceptors)
- [Frontend License Checking](#frontend-license-checking)
  - [License Context](#license-context)
  - [Feature Gate Component](#feature-gate-component)
  - [API Integration](#api-integration)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)

---

## Overview

Linkrift uses a cryptographic license verification system based on **Ed25519 digital signatures**. This system ensures:

- **Tamper-proof licenses** - Licenses cannot be modified without detection
- **Offline verification** - No network call needed to verify licenses
- **Feature gating** - Specific features can be enabled/disabled per license
- **Expiration enforcement** - Time-limited licenses are enforced

---

## License Architecture

### License Structure

Licenses are encoded as base64 JSON payloads with an Ed25519 signature:

```go
// Package license provides license verification for Linkrift
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// License represents a Linkrift license
type License struct {
	// Unique license identifier
	ID string `json:"id"`

	// License holder information
	CustomerID   string `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Email        string `json:"email"`

	// License type and tier
	Type LicenseType `json:"type"`
	Tier Tier        `json:"tier"`

	// Validity period
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`

	// Feature entitlements
	Features []Feature `json:"features"`

	// Usage limits
	Limits Limits `json:"limits"`

	// License metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Limits defines usage limits for the license
type Limits struct {
	MaxUsers         int   `json:"max_users"`
	MaxDomains       int   `json:"max_domains"`
	MaxLinksPerMonth int64 `json:"max_links_per_month"`
	MaxClicksPerMonth int64 `json:"max_clicks_per_month"`
	MaxWorkspaces    int   `json:"max_workspaces"`
	MaxAPIRequests   int64 `json:"max_api_requests_per_minute"`
}

// SignedLicense contains the license and its cryptographic signature
type SignedLicense struct {
	License   string `json:"license"`   // Base64-encoded license JSON
	Signature string `json:"signature"` // Base64-encoded Ed25519 signature
	Version   int    `json:"version"`   // License format version
}
```

### License Types

```go
package license

// LicenseType represents the type of license
type LicenseType string

const (
	// LicenseTypeTrial is a time-limited trial license
	LicenseTypeTrial LicenseType = "trial"

	// LicenseTypeSubscription is a recurring subscription license
	LicenseTypeSubscription LicenseType = "subscription"

	// LicenseTypePerpetual is a one-time purchase license
	LicenseTypePerpetual LicenseType = "perpetual"

	// LicenseTypeEnterprise is a custom enterprise agreement
	LicenseTypeEnterprise LicenseType = "enterprise"
)

// Tier represents the license tier
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierBusiness   Tier = "business"
	TierEnterprise Tier = "enterprise"
)

// TierLevel returns the numeric level of the tier for comparison
func (t Tier) Level() int {
	switch t {
	case TierEnterprise:
		return 4
	case TierBusiness:
		return 3
	case TierPro:
		return 2
	case TierFree:
		return 1
	default:
		return 0
	}
}

// IncludesTier checks if this tier includes features of another tier
func (t Tier) IncludesTier(other Tier) bool {
	return t.Level() >= other.Level()
}
```

---

## Ed25519 Cryptographic Signing

### Key Generation

Linkrift uses Ed25519 for license signing. Keys are generated during initial setup:

```go
package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
)

// GenerateKeyPair creates a new Ed25519 key pair for license signing
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}

// SavePrivateKey saves the private key to a PEM file (keep secure!)
func SavePrivateKey(privateKey ed25519.PrivateKey, path string) error {
	block := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: privateKey,
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}

// SavePublicKey saves the public key to a PEM file (distributed with binary)
func SavePublicKey(publicKey ed25519.PublicKey, path string) error {
	block := &pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: publicKey,
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, block)
}
```

### License Signing

The license server signs licenses with the private key:

```go
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
)

// Signer creates signed licenses
type Signer struct {
	privateKey ed25519.PrivateKey
}

// NewSigner creates a new license signer
func NewSigner(privateKey ed25519.PrivateKey) *Signer {
	return &Signer{privateKey: privateKey}
}

// Sign creates a signed license from a License struct
func (s *Signer) Sign(license *License) (*SignedLicense, error) {
	// Serialize license to JSON
	licenseJSON, err := json.Marshal(license)
	if err != nil {
		return nil, err
	}

	// Base64 encode the license
	licenseB64 := base64.StdEncoding.EncodeToString(licenseJSON)

	// Sign the base64-encoded license
	signature := ed25519.Sign(s.privateKey, []byte(licenseB64))

	return &SignedLicense{
		License:   licenseB64,
		Signature: base64.StdEncoding.EncodeToString(signature),
		Version:   1,
	}, nil
}

// SignToString creates a signed license and returns it as a single base64 string
func (s *Signer) SignToString(license *License) (string, error) {
	signed, err := s.Sign(license)
	if err != nil {
		return "", err
	}

	signedJSON, err := json.Marshal(signed)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signedJSON), nil
}
```

### Signature Verification

The public key is embedded in the Linkrift binary for verification:

```go
package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	_ "embed"
)

//go:embed keys/public.pem
var embeddedPublicKey []byte

var (
	ErrInvalidSignature = errors.New("invalid license signature")
	ErrLicenseExpired   = errors.New("license has expired")
	ErrLicenseNotYetValid = errors.New("license is not yet valid")
	ErrInvalidLicenseFormat = errors.New("invalid license format")
)

// Verifier verifies license signatures
type Verifier struct {
	publicKey ed25519.PublicKey
}

// NewVerifier creates a new license verifier with the embedded public key
func NewVerifier() (*Verifier, error) {
	publicKey, err := parsePublicKey(embeddedPublicKey)
	if err != nil {
		return nil, err
	}
	return &Verifier{publicKey: publicKey}, nil
}

// Verify checks the signature and validity of a signed license
func (v *Verifier) Verify(signedLicense *SignedLicense) (*License, error) {
	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(signedLicense.Signature)
	if err != nil {
		return nil, ErrInvalidLicenseFormat
	}

	// Verify signature
	if !ed25519.Verify(v.publicKey, []byte(signedLicense.License), signature) {
		return nil, ErrInvalidSignature
	}

	// Decode license
	licenseJSON, err := base64.StdEncoding.DecodeString(signedLicense.License)
	if err != nil {
		return nil, ErrInvalidLicenseFormat
	}

	var license License
	if err := json.Unmarshal(licenseJSON, &license); err != nil {
		return nil, ErrInvalidLicenseFormat
	}

	// Check validity period
	now := time.Now()
	if now.Before(license.IssuedAt) {
		return nil, ErrLicenseNotYetValid
	}
	if now.After(license.ExpiresAt) {
		return nil, ErrLicenseExpired
	}

	return &license, nil
}

// VerifyString verifies a license from a single base64-encoded string
func (v *Verifier) VerifyString(licenseKey string) (*License, error) {
	signedJSON, err := base64.StdEncoding.DecodeString(licenseKey)
	if err != nil {
		return nil, ErrInvalidLicenseFormat
	}

	var signed SignedLicense
	if err := json.Unmarshal(signedJSON, &signed); err != nil {
		return nil, ErrInvalidLicenseFormat
	}

	return v.Verify(&signed)
}
```

---

## License Verification Flow

### Startup Verification

License verification happens at application startup:

```go
package main

import (
	"log"
	"os"

	"github.com/link-rift/link-rift/internal/license"
)

func main() {
	// Initialize license verifier
	verifier, err := license.NewVerifier()
	if err != nil {
		log.Fatal("Failed to initialize license verifier:", err)
	}

	// Get license key from environment
	licenseKey := os.Getenv("LINKRIFT_LICENSE_KEY")

	// Initialize license manager
	manager := license.NewManager(verifier)

	if licenseKey != "" {
		// Verify and load commercial license
		lic, err := manager.LoadLicense(licenseKey)
		if err != nil {
			log.Printf("Warning: Invalid license key: %v", err)
			log.Println("Running in Community Edition mode")
			manager.SetCommunityEdition()
		} else {
			log.Printf("License verified: %s (%s tier)", lic.CustomerName, lic.Tier)
		}
	} else {
		log.Println("No license key provided. Running in Community Edition mode")
		manager.SetCommunityEdition()
	}

	// Start application with license manager
	app := NewApp(manager)
	app.Run()
}
```

### Runtime Verification

Periodic license checks ensure validity during runtime:

```go
package license

import (
	"context"
	"sync"
	"time"
)

// Manager manages license state and provides feature checks
type Manager struct {
	verifier *Verifier
	license  *License
	mu       sync.RWMutex

	// Callback for license state changes
	onLicenseChange func(*License)
}

// NewManager creates a new license manager
func NewManager(verifier *Verifier) *Manager {
	return &Manager{
		verifier: verifier,
	}
}

// LoadLicense loads and verifies a license key
func (m *Manager) LoadLicense(licenseKey string) (*License, error) {
	lic, err := m.verifier.VerifyString(licenseKey)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.license = lic
	m.mu.Unlock()

	if m.onLicenseChange != nil {
		m.onLicenseChange(lic)
	}

	return lic, nil
}

// SetCommunityEdition sets the manager to Community Edition mode
func (m *Manager) SetCommunityEdition() {
	m.mu.Lock()
	m.license = &License{
		Tier:     TierFree,
		Features: communityFeatures(),
		Limits:   communityLimits(),
	}
	m.mu.Unlock()
}

// StartPeriodicCheck starts a goroutine that periodically verifies the license
func (m *Manager) StartPeriodicCheck(ctx context.Context, licenseKey string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.mu.RLock()
				currentLicense := m.license
				m.mu.RUnlock()

				// Re-verify the license
				lic, err := m.verifier.VerifyString(licenseKey)
				if err != nil {
					// License became invalid (e.g., expired)
					if currentLicense != nil && currentLicense.Tier != TierFree {
						m.SetCommunityEdition()
					}
					continue
				}

				m.mu.Lock()
				m.license = lic
				m.mu.Unlock()
			}
		}
	}()
}

// GetLicense returns the current license (thread-safe)
func (m *Manager) GetLicense() *License {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license
}
```

---

## Feature Flag Checking

### Feature Registry

Features are defined with their minimum required tier:

```go
package license

// Feature represents an enterprise feature
type Feature string

const (
	// Analytics features
	FeatureAdvancedAnalytics Feature = "advanced_analytics"
	FeatureCustomReports     Feature = "custom_reports"
	FeatureRealTimeDashboard Feature = "realtime_dashboard"

	// Authentication features
	FeatureSAML     Feature = "saml"
	FeatureSCIM     Feature = "scim"
	FeatureOIDC     Feature = "oidc"

	// Domain features
	FeatureCustomDomains    Feature = "custom_domains"
	FeatureUnlimitedDomains Feature = "unlimited_domains"
	FeatureWhiteLabeling    Feature = "white_labeling"

	// Team features
	FeatureRBAC            Feature = "rbac"
	FeatureAdvancedRBAC    Feature = "advanced_rbac"
	FeatureWorkspaces      Feature = "workspaces"
	FeatureAuditLogs       Feature = "audit_logs"

	// Infrastructure features
	FeatureMultiRegion     Feature = "multi_region"
	FeatureHAClustering    Feature = "ha_clustering"
	FeatureCustomRateLimit Feature = "custom_rate_limit"

	// Compliance features
	FeatureSOC2     Feature = "soc2"
	FeatureHIPAA    Feature = "hipaa"
	FeatureGDPRTools Feature = "gdpr_tools"
)

// FeatureDefinition defines a feature and its requirements
type FeatureDefinition struct {
	Name        Feature
	Description string
	MinTier     Tier
	Category    string
}

// featureRegistry contains all feature definitions
var featureRegistry = map[Feature]FeatureDefinition{
	FeatureAdvancedAnalytics: {
		Name:        FeatureAdvancedAnalytics,
		Description: "Advanced analytics with detailed breakdowns",
		MinTier:     TierPro,
		Category:    "analytics",
	},
	FeatureSAML: {
		Name:        FeatureSAML,
		Description: "SAML 2.0 single sign-on",
		MinTier:     TierBusiness,
		Category:    "authentication",
	},
	FeatureAuditLogs: {
		Name:        FeatureAuditLogs,
		Description: "Comprehensive audit logging",
		MinTier:     TierBusiness,
		Category:    "compliance",
	},
	FeatureHIPAA: {
		Name:        FeatureHIPAA,
		Description: "HIPAA compliance features and BAA",
		MinTier:     TierEnterprise,
		Category:    "compliance",
	},
	// ... additional features
}
```

### Feature Checking Code

```go
package license

import "errors"

var ErrFeatureNotAvailable = errors.New("feature not available in current license")

// HasFeature checks if the current license includes a specific feature
func (m *Manager) HasFeature(feature Feature) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.license == nil {
		return false
	}

	// Check if feature is explicitly included
	for _, f := range m.license.Features {
		if f == feature {
			return true
		}
	}

	// Check if tier includes the feature
	def, ok := featureRegistry[feature]
	if !ok {
		return false
	}

	return m.license.Tier.IncludesTier(def.MinTier)
}

// RequireFeature returns an error if the feature is not available
func (m *Manager) RequireFeature(feature Feature) error {
	if !m.HasFeature(feature) {
		return ErrFeatureNotAvailable
	}
	return nil
}

// GetTier returns the current license tier
func (m *Manager) GetTier() Tier {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.license == nil {
		return TierFree
	}
	return m.license.Tier
}

// CheckLimit verifies if usage is within license limits
func (m *Manager) CheckLimit(limitType string, current int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.license == nil {
		return false
	}

	switch limitType {
	case "users":
		return int64(m.license.Limits.MaxUsers) == 0 || current < int64(m.license.Limits.MaxUsers)
	case "domains":
		return int64(m.license.Limits.MaxDomains) == 0 || current < int64(m.license.Limits.MaxDomains)
	case "links_per_month":
		return m.license.Limits.MaxLinksPerMonth == 0 || current < m.license.Limits.MaxLinksPerMonth
	case "clicks_per_month":
		return m.license.Limits.MaxClicksPerMonth == 0 || current < m.license.Limits.MaxClicksPerMonth
	default:
		return false
	}
}
```

---

## Middleware for Feature Gating

### HTTP Middleware

```go
package middleware

import (
	"net/http"

	"github.com/link-rift/link-rift/internal/license"
)

// FeatureGate creates middleware that gates access based on license features
func FeatureGate(manager *license.Manager, feature license.Feature) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !manager.HasFeature(feature) {
				http.Error(w, "This feature requires a higher license tier", http.StatusPaymentRequired)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TierGate creates middleware that requires a minimum tier
func TierGate(manager *license.Manager, minTier license.Tier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentTier := manager.GetTier()
			if !currentTier.IncludesTier(minTier) {
				http.Error(w, "This feature requires a higher license tier", http.StatusPaymentRequired)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Example usage with chi router
func SetupRoutes(r *chi.Mux, manager *license.Manager) {
	// Public routes (all tiers)
	r.Get("/api/links", listLinks)
	r.Post("/api/links", createLink)

	// Pro tier routes
	r.Group(func(r chi.Router) {
		r.Use(TierGate(manager, license.TierPro))
		r.Get("/api/analytics/detailed", detailedAnalytics)
	})

	// Business tier routes
	r.Group(func(r chi.Router) {
		r.Use(TierGate(manager, license.TierBusiness))
		r.Get("/api/audit-logs", getAuditLogs)
		r.Post("/api/saml/configure", configureSAML)
	})

	// Feature-specific routes
	r.Group(func(r chi.Router) {
		r.Use(FeatureGate(manager, license.FeatureCustomReports))
		r.Get("/api/reports/custom", getCustomReports)
		r.Post("/api/reports/custom", createCustomReport)
	})
}
```

### gRPC Interceptors

```go
package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/link-rift/link-rift/internal/license"
)

// FeatureGateInterceptor creates a gRPC unary interceptor for feature gating
func FeatureGateInterceptor(manager *license.Manager, featureMap map[string]license.Feature) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if this method requires a specific feature
		if feature, ok := featureMap[info.FullMethod]; ok {
			if !manager.HasFeature(feature) {
				return nil, status.Errorf(
					codes.PermissionDenied,
					"Feature %s requires a higher license tier",
					feature,
				)
			}
		}

		return handler(ctx, req)
	}
}

// TierGateInterceptor creates a gRPC interceptor for tier-based access
func TierGateInterceptor(manager *license.Manager, tierMap map[string]license.Tier) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if minTier, ok := tierMap[info.FullMethod]; ok {
			currentTier := manager.GetTier()
			if !currentTier.IncludesTier(minTier) {
				return nil, status.Errorf(
					codes.PermissionDenied,
					"This operation requires %s tier or higher",
					minTier,
				)
			}
		}

		return handler(ctx, req)
	}
}
```

---

## Frontend License Checking

### License Context

```typescript
// src/contexts/LicenseContext.tsx
import React, { createContext, useContext, useEffect, useState } from 'react';

interface LicenseInfo {
  tier: 'free' | 'pro' | 'business' | 'enterprise';
  features: string[];
  limits: {
    maxUsers: number;
    maxDomains: number;
    maxLinksPerMonth: number;
  };
  expiresAt: string | null;
  isValid: boolean;
}

interface LicenseContextType {
  license: LicenseInfo | null;
  loading: boolean;
  hasFeature: (feature: string) => boolean;
  hasTier: (tier: string) => boolean;
  checkLimit: (limitType: string, current: number) => boolean;
}

const LicenseContext = createContext<LicenseContextType | undefined>(undefined);

const tierLevels: Record<string, number> = {
  free: 1,
  pro: 2,
  business: 3,
  enterprise: 4,
};

export function LicenseProvider({ children }: { children: React.ReactNode }) {
  const [license, setLicense] = useState<LicenseInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchLicense();
  }, []);

  async function fetchLicense() {
    try {
      const response = await fetch('/api/license');
      const data = await response.json();
      setLicense(data);
    } catch (error) {
      console.error('Failed to fetch license:', error);
      // Default to free tier on error
      setLicense({
        tier: 'free',
        features: [],
        limits: { maxUsers: 1, maxDomains: 1, maxLinksPerMonth: 1000 },
        expiresAt: null,
        isValid: true,
      });
    } finally {
      setLoading(false);
    }
  }

  function hasFeature(feature: string): boolean {
    if (!license) return false;
    return license.features.includes(feature);
  }

  function hasTier(tier: string): boolean {
    if (!license) return false;
    const currentLevel = tierLevels[license.tier] || 0;
    const requiredLevel = tierLevels[tier] || 0;
    return currentLevel >= requiredLevel;
  }

  function checkLimit(limitType: string, current: number): boolean {
    if (!license) return false;
    const limit = license.limits[limitType as keyof typeof license.limits];
    return limit === 0 || current < limit; // 0 means unlimited
  }

  return (
    <LicenseContext.Provider
      value={{ license, loading, hasFeature, hasTier, checkLimit }}
    >
      {children}
    </LicenseContext.Provider>
  );
}

export function useLicense() {
  const context = useContext(LicenseContext);
  if (context === undefined) {
    throw new Error('useLicense must be used within a LicenseProvider');
  }
  return context;
}
```

### Feature Gate Component

```typescript
// src/components/FeatureGate.tsx
import React from 'react';
import { useLicense } from '../contexts/LicenseContext';

interface FeatureGateProps {
  feature?: string;
  tier?: 'free' | 'pro' | 'business' | 'enterprise';
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function FeatureGate({
  feature,
  tier,
  children,
  fallback,
}: FeatureGateProps) {
  const { hasFeature, hasTier, loading } = useLicense();

  if (loading) {
    return null; // Or a loading spinner
  }

  const hasAccess = feature
    ? hasFeature(feature)
    : tier
    ? hasTier(tier)
    : true;

  if (!hasAccess) {
    return fallback ? <>{fallback}</> : null;
  }

  return <>{children}</>;
}

// Upgrade prompt component
interface UpgradePromptProps {
  feature: string;
  requiredTier: string;
}

export function UpgradePrompt({ feature, requiredTier }: UpgradePromptProps) {
  return (
    <div className="p-4 bg-gradient-to-r from-purple-500 to-indigo-600 rounded-lg text-white">
      <h3 className="font-semibold">Upgrade Required</h3>
      <p className="text-sm opacity-90">
        The {feature} feature requires {requiredTier} tier or higher.
      </p>
      <a
        href="/settings/billing"
        className="mt-2 inline-block px-4 py-2 bg-white text-indigo-600 rounded font-medium"
      >
        Upgrade Now
      </a>
    </div>
  );
}

// Usage example
function AnalyticsDashboard() {
  return (
    <div>
      <h1>Analytics</h1>

      {/* Basic analytics - available to all */}
      <BasicStats />

      {/* Advanced analytics - Pro and above */}
      <FeatureGate
        tier="pro"
        fallback={
          <UpgradePrompt
            feature="Advanced Analytics"
            requiredTier="Pro"
          />
        }
      >
        <AdvancedAnalytics />
      </FeatureGate>

      {/* Custom reports - requires specific feature */}
      <FeatureGate
        feature="custom_reports"
        fallback={
          <UpgradePrompt
            feature="Custom Reports"
            requiredTier="Business"
          />
        }
      >
        <CustomReports />
      </FeatureGate>
    </div>
  );
}
```

### API Integration

```typescript
// src/lib/api.ts
import { useLicense } from '../contexts/LicenseContext';

// Hook for license-aware API calls
export function useLicenseAwareApi() {
  const { license, hasFeature, hasTier } = useLicense();

  async function fetchWithLicense<T>(
    url: string,
    options?: RequestInit
  ): Promise<T> {
    const response = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        'Content-Type': 'application/json',
      },
    });

    if (response.status === 402) {
      // Payment Required - feature not available
      const error = await response.json();
      throw new UpgradeRequiredError(error.message, error.requiredTier);
    }

    if (!response.ok) {
      throw new Error('API request failed');
    }

    return response.json();
  }

  return {
    fetch: fetchWithLicense,
    hasFeature,
    hasTier,
    license,
  };
}

class UpgradeRequiredError extends Error {
  constructor(
    message: string,
    public requiredTier: string
  ) {
    super(message);
    this.name = 'UpgradeRequiredError';
  }
}
```

---

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `LINKRIFT_LICENSE_KEY` | Base64-encoded license key | No | None (CE mode) |
| `LINKRIFT_LICENSE_FILE` | Path to license key file | No | None |
| `LINKRIFT_LICENSE_CHECK_INTERVAL` | License revalidation interval | No | `1h` |
| `LINKRIFT_OFFLINE_MODE` | Skip online license validation | No | `false` |

```bash
# Example .env file
LINKRIFT_LICENSE_KEY=eyJsaWNlbnNlIjoiZXlKcFpDSTZJbXhwWXkweE1qTTBOVFkzT...

# Or use a file
LINKRIFT_LICENSE_FILE=/etc/linkrift/license.key

# Validation settings
LINKRIFT_LICENSE_CHECK_INTERVAL=30m
LINKRIFT_OFFLINE_MODE=false
```

---

## Troubleshooting

### Common Issues

**License Not Recognized**
```bash
# Check license format
echo $LINKRIFT_LICENSE_KEY | base64 -d | jq .

# Verify license with CLI
linkrift license verify --key $LINKRIFT_LICENSE_KEY
```

**License Expired**
```bash
# Check expiration date
linkrift license info

# Output:
# License ID: lic-123456
# Customer: Acme Corp
# Tier: Business
# Expires: 2025-12-31T23:59:59Z
# Status: VALID (expires in 342 days)
```

**Feature Not Available**
```go
// Debug feature availability
func debugFeature(manager *license.Manager, feature license.Feature) {
    log.Printf("Feature: %s", feature)
    log.Printf("Has Feature: %v", manager.HasFeature(feature))
    log.Printf("Current Tier: %s", manager.GetTier())

    if def, ok := featureRegistry[feature]; ok {
        log.Printf("Required Tier: %s", def.MinTier)
    }
}
```

---

## Related Documentation

- [Open Core Model](./OPEN_CORE_MODEL.md) - Business model and pricing
- [Repository Structure](./REPOSITORY_STRUCTURE.md) - Code organization
- [API Reference](../api/README.md) - REST API documentation
