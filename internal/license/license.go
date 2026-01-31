package license

import "time"

// LicenseType represents the type of license.
type LicenseType string

const (
	LicenseTypeTrial        LicenseType = "trial"
	LicenseTypeSubscription LicenseType = "subscription"
	LicenseTypePerpetual    LicenseType = "perpetual"
	LicenseTypeEnterprise   LicenseType = "enterprise"
)

// License holds the decoded license data.
type License struct {
	ID           string            `json:"id"`
	CustomerID   string            `json:"customer_id"`
	CustomerName string            `json:"customer_name"`
	Email        string            `json:"email"`
	Type         LicenseType       `json:"type"`
	Tier         Tier              `json:"tier"`
	IssuedAt     time.Time         `json:"issued_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Features     []Feature         `json:"features"`
	Limits       Limits            `json:"limits"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SignedLicense is the wire format for a license key.
type SignedLicense struct {
	License   string `json:"license"`   // base64-encoded JSON
	Signature string `json:"signature"` // base64-encoded Ed25519 signature
	Version   int    `json:"version"`
}

// IsExpired returns true if the license has expired.
func (l *License) IsExpired() bool {
	if l.ExpiresAt.IsZero() {
		return false // perpetual
	}
	return time.Now().After(l.ExpiresAt)
}

// IsValid returns true if the license is not expired and not issued in the future.
func (l *License) IsValid() bool {
	now := time.Now()
	if !l.IssuedAt.IsZero() && now.Before(l.IssuedAt) {
		return false
	}
	return !l.IsExpired()
}

// HasFeature checks if this license explicitly includes a feature,
// or if the feature is available at the license's tier level.
func (l *License) HasFeature(f Feature) bool {
	for _, feat := range l.Features {
		if feat == f {
			return true
		}
	}
	def, ok := GetFeatureDefinition(f)
	if !ok {
		return false
	}
	return l.Tier.IncludesTier(def.MinTier)
}

// LicenseResponse is the safe API output for license info.
type LicenseResponse struct {
	ID           string            `json:"id,omitempty"`
	CustomerName string            `json:"customer_name,omitempty"`
	Type         LicenseType       `json:"type"`
	Tier         Tier              `json:"tier"`
	Plan         Plan              `json:"plan"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Features     []Feature         `json:"features"`
	Limits       Limits            `json:"limits"`
	IsCommunity  bool              `json:"is_community"`
}

// ToResponse converts a License to a safe API response.
func (l *License) ToResponse(isCommunity bool) *LicenseResponse {
	resp := &LicenseResponse{
		Type:        l.Type,
		Tier:        l.Tier,
		Plan:        Plans[l.Tier],
		Features:    FeaturesForTier(l.Tier),
		Limits:      l.Limits,
		IsCommunity: isCommunity,
	}

	if !isCommunity {
		resp.ID = l.ID
		resp.CustomerName = l.CustomerName
		if !l.ExpiresAt.IsZero() {
			resp.ExpiresAt = &l.ExpiresAt
		}
		// Merge explicitly granted features with tier-based features
		featureSet := make(map[Feature]struct{})
		for _, f := range resp.Features {
			featureSet[f] = struct{}{}
		}
		for _, f := range l.Features {
			featureSet[f] = struct{}{}
		}
		features := make([]Feature, 0, len(featureSet))
		for f := range featureSet {
			features = append(features, f)
		}
		resp.Features = features
	}

	return resp
}
