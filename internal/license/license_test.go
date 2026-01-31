package license

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// --- Tier Tests ---

func TestTierLevel(t *testing.T) {
	tests := []struct {
		tier  Tier
		level int
	}{
		{TierFree, 1},
		{TierPro, 2},
		{TierBusiness, 3},
		{TierEnterprise, 4},
		{Tier("unknown"), 0},
	}

	for _, tt := range tests {
		if got := tt.tier.Level(); got != tt.level {
			t.Errorf("Tier(%q).Level() = %d, want %d", tt.tier, got, tt.level)
		}
	}
}

func TestTierIncludesTier(t *testing.T) {
	tests := []struct {
		tier     Tier
		other    Tier
		includes bool
	}{
		{TierEnterprise, TierFree, true},
		{TierEnterprise, TierEnterprise, true},
		{TierFree, TierPro, false},
		{TierPro, TierBusiness, false},
		{TierBusiness, TierBusiness, true},
	}

	for _, tt := range tests {
		if got := tt.tier.IncludesTier(tt.other); got != tt.includes {
			t.Errorf("Tier(%q).IncludesTier(%q) = %v, want %v", tt.tier, tt.other, got, tt.includes)
		}
	}
}

func TestTierIsValid(t *testing.T) {
	if !TierFree.IsValid() {
		t.Error("TierFree should be valid")
	}
	if Tier("invalid").IsValid() {
		t.Error("invalid tier should not be valid")
	}
}

// --- Feature Tests ---

func TestGetFeatureDefinition(t *testing.T) {
	def, ok := GetFeatureDefinition(FeatureSAML)
	if !ok {
		t.Fatal("FeatureSAML should have a definition")
	}
	if def.MinTier != TierEnterprise {
		t.Errorf("FeatureSAML MinTier = %q, want %q", def.MinTier, TierEnterprise)
	}
}

func TestGetFeatureDefinitionUnknown(t *testing.T) {
	_, ok := GetFeatureDefinition(Feature("nonexistent"))
	if ok {
		t.Error("unknown feature should return false")
	}
}

func TestFeaturesForTier(t *testing.T) {
	freeFeatures := FeaturesForTier(TierFree)
	entFeatures := FeaturesForTier(TierEnterprise)

	if len(entFeatures) <= len(freeFeatures) {
		t.Errorf("enterprise should have more features than free: ent=%d, free=%d", len(entFeatures), len(freeFeatures))
	}
}

func TestAllFeatures(t *testing.T) {
	all := AllFeatures()
	if len(all) == 0 {
		t.Error("AllFeatures should return at least one feature")
	}
}

// --- Limits Tests ---

func TestDefaultLimits(t *testing.T) {
	free := DefaultLimits(TierFree)
	if free.MaxUsers != 1 {
		t.Errorf("Free MaxUsers = %d, want 1", free.MaxUsers)
	}

	ent := DefaultLimits(TierEnterprise)
	if ent.MaxUsers != -1 {
		t.Errorf("Enterprise MaxUsers = %d, want -1 (unlimited)", ent.MaxUsers)
	}
}

func TestDefaultLimitsUnknownTier(t *testing.T) {
	limits := DefaultLimits(Tier("unknown"))
	if limits.MaxUsers != 1 {
		t.Errorf("unknown tier should fall back to free limits, got MaxUsers=%d", limits.MaxUsers)
	}
}

func TestLimitsCheckLimit(t *testing.T) {
	limits := DefaultLimits(TierFree)

	// Free tier: MaxUsers = 1
	if !limits.CheckLimit(LimitMaxUsers, 0) {
		t.Error("0 users should be within limit of 1")
	}
	if limits.CheckLimit(LimitMaxUsers, 1) {
		t.Error("1 user should NOT be within limit of 1 (current < limit)")
	}
}

func TestLimitsCheckLimitUnlimited(t *testing.T) {
	limits := DefaultLimits(TierEnterprise)
	if !limits.CheckLimit(LimitMaxUsers, 999999) {
		t.Error("unlimited (-1) should always pass")
	}
}

func TestLimitsGetLimit(t *testing.T) {
	limits := DefaultLimits(TierPro)
	if got := limits.GetLimit(LimitMaxDomains); got != 3 {
		t.Errorf("Pro MaxDomains = %d, want 3", got)
	}
	if got := limits.GetLimit(LimitType("unknown")); got != 0 {
		t.Errorf("unknown limit type should return 0, got %d", got)
	}
}

// --- License Tests ---

func TestLicenseIsExpired(t *testing.T) {
	lic := &License{ExpiresAt: time.Now().Add(-1 * time.Hour)}
	if !lic.IsExpired() {
		t.Error("license with past expiry should be expired")
	}

	lic2 := &License{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if lic2.IsExpired() {
		t.Error("license with future expiry should not be expired")
	}

	lic3 := &License{} // zero ExpiresAt = perpetual
	if lic3.IsExpired() {
		t.Error("perpetual license should not be expired")
	}
}

func TestLicenseIsValid(t *testing.T) {
	lic := &License{
		IssuedAt:  time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if !lic.IsValid() {
		t.Error("valid license should be valid")
	}

	future := &License{
		IssuedAt:  time.Now().Add(1 * time.Hour),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	if future.IsValid() {
		t.Error("not-yet-valid license should be invalid")
	}
}

func TestLicenseHasFeature(t *testing.T) {
	lic := &License{
		Tier:     TierPro,
		Features: []Feature{FeatureSAML}, // explicitly granted despite tier
	}

	if !lic.HasFeature(FeatureSAML) {
		t.Error("explicitly granted feature should be available")
	}

	if !lic.HasFeature(FeatureCustomDomains) {
		t.Error("Pro tier should include CustomDomains (MinTier=Pro)")
	}

	if lic.HasFeature(FeatureAuditLogs) {
		t.Error("Pro tier should NOT include AuditLogs (MinTier=Enterprise)")
	}
}

func TestLicenseToResponse(t *testing.T) {
	lic := newTestLicense()
	resp := lic.ToResponse(false)

	if resp.IsCommunity {
		t.Error("should not be community")
	}
	if resp.ID != lic.ID {
		t.Error("ID should be set for non-community")
	}
	if resp.Tier != TierPro {
		t.Errorf("tier = %q, want %q", resp.Tier, TierPro)
	}

	ceResp := lic.ToResponse(true)
	if !ceResp.IsCommunity {
		t.Error("should be community")
	}
	if ceResp.ID != "" {
		t.Error("ID should be empty for community")
	}
}

// --- Verify Tests ---

func TestVerifyValidLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	lic := newTestLicense()
	signed := signer.Sign(t, lic)

	result, err := verifier.Verify(signed)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if result.ID != lic.ID {
		t.Errorf("ID = %q, want %q", result.ID, lic.ID)
	}
}

func TestVerifyExpiredLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	lic := newTestLicense()
	lic.ExpiresAt = time.Now().Add(-1 * time.Hour)
	signed := signer.Sign(t, lic)

	_, err = verifier.Verify(signed)
	if err != ErrLicenseExpired {
		t.Errorf("expected ErrLicenseExpired, got %v", err)
	}
}

func TestVerifyTamperedLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	lic := newTestLicense()
	signed := signer.Sign(t, lic)

	// Tamper with the license by using a different signer
	other := GenerateKeyPair(t)
	tampered := other.Sign(t, lic)
	signed.Signature = tampered.Signature

	_, err = verifier.Verify(signed)
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifyMalformedLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	_, err = verifier.Verify(nil)
	if err != ErrInvalidLicenseFormat {
		t.Errorf("expected ErrInvalidLicenseFormat for nil, got %v", err)
	}

	_, err = verifier.Verify(&SignedLicense{
		License:   "not-valid-base64!!!",
		Signature: "also-bad",
	})
	if err == nil {
		t.Error("expected error for malformed base64")
	}
}

func TestVerifyString(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	lic := newTestLicense()
	key := signer.SignToString(t, lic)

	result, err := verifier.VerifyString(key)
	if err != nil {
		t.Fatalf("VerifyString: %v", err)
	}

	if result.ID != lic.ID {
		t.Errorf("ID = %q, want %q", result.ID, lic.ID)
	}
}

func TestVerifyStringMalformed(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	_, err = verifier.VerifyString("not-a-license-key")
	if err == nil {
		t.Error("expected error for malformed key")
	}
}

// --- Manager Tests ---

func TestManagerCommunityEdition(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	if !mgr.IsCommunity() {
		t.Error("new manager should start as community edition")
	}
	if mgr.GetTier() != TierFree {
		t.Errorf("community tier = %q, want %q", mgr.GetTier(), TierFree)
	}
}

func TestManagerLoadLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	lic := newTestLicense()
	key := signer.SignToString(t, lic)

	if err := mgr.LoadLicense(key); err != nil {
		t.Fatalf("LoadLicense: %v", err)
	}

	if mgr.IsCommunity() {
		t.Error("should not be community after loading license")
	}
	if mgr.GetTier() != TierPro {
		t.Errorf("tier = %q, want %q", mgr.GetTier(), TierPro)
	}
}

func TestManagerHasFeature(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	lic := newTestLicense()
	lic.Tier = TierBusiness
	key := signer.SignToString(t, lic)
	if err := mgr.LoadLicense(key); err != nil {
		t.Fatalf("LoadLicense: %v", err)
	}

	if !mgr.HasFeature(FeatureTeamMembers) {
		t.Error("Business tier should have TeamMembers")
	}
	if mgr.HasFeature(FeatureSAML) {
		t.Error("Business tier should NOT have SAML (enterprise only)")
	}
}

func TestManagerCheckLimit(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	// CE = Free tier, MaxUsers = 1
	if !mgr.CheckLimit(LimitMaxUsers, 0) {
		t.Error("0 users should be within free limit")
	}
	if mgr.CheckLimit(LimitMaxUsers, 1) {
		t.Error("1 user should NOT be within free limit of 1")
	}
}

func TestManagerRemoveLicense(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	lic := newTestLicense()
	key := signer.SignToString(t, lic)
	if err := mgr.LoadLicense(key); err != nil {
		t.Fatalf("LoadLicense: %v", err)
	}

	mgr.RemoveLicense()

	if !mgr.IsCommunity() {
		t.Error("should be community after removing license")
	}
	if mgr.GetTier() != TierFree {
		t.Errorf("tier = %q, want %q after remove", mgr.GetTier(), TierFree)
	}
}

func TestManagerPeriodicCheck(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	// Load a license that expires very soon
	lic := newTestLicense()
	lic.ExpiresAt = time.Now().Add(50 * time.Millisecond)
	key := signer.SignToString(t, lic)
	if err := mgr.LoadLicense(key); err != nil {
		t.Fatalf("LoadLicense: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.StartPeriodicCheck(ctx, 100*time.Millisecond)

	// Wait for license to expire and periodic check to fire
	time.Sleep(300 * time.Millisecond)

	if !mgr.IsCommunity() {
		t.Error("should revert to community after license expires during periodic check")
	}
}

func TestManagerGetLicenseResponse(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	resp := mgr.GetLicenseResponse()
	if !resp.IsCommunity {
		t.Error("default response should be community")
	}
	if resp.Tier != TierFree {
		t.Errorf("tier = %q, want %q", resp.Tier, TierFree)
	}
}

func TestManagerLoadLicenseDefaultLimits(t *testing.T) {
	signer := GenerateKeyPair(t)
	verifier, err := NewVerifierWithKey(signer.PublicKeyPEM())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	logger := zap.NewNop()
	mgr := NewManager(verifier, logger)

	// License with zero limits — should get tier defaults
	lic := &License{
		ID:        "test-no-limits",
		Type:      LicenseTypeSubscription,
		Tier:      TierBusiness,
		IssuedAt:  time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}
	key := signer.SignToString(t, lic)
	if err := mgr.LoadLicense(key); err != nil {
		t.Fatalf("LoadLicense: %v", err)
	}

	expected := DefaultLimits(TierBusiness)
	if mgr.GetLimits().MaxUsers != expected.MaxUsers {
		t.Errorf("MaxUsers = %d, want %d (tier default)", mgr.GetLimits().MaxUsers, expected.MaxUsers)
	}
}
