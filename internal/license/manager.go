package license

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager manages the current license state.
type Manager struct {
	mu          sync.RWMutex
	verifier    *Verifier
	license     *License
	licenseKey  string
	isCommunity bool
	logger      *zap.Logger
}

// NewManager creates a new license manager.
func NewManager(verifier *Verifier, logger *zap.Logger) *Manager {
	m := &Manager{
		verifier: verifier,
		logger:   logger,
	}
	m.SetCommunityEdition()
	return m
}

// LoadLicense loads and verifies a license key.
func (m *Manager) LoadLicense(key string) error {
	lic, err := m.verifier.VerifyString(key)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.license = lic
	m.licenseKey = key
	m.isCommunity = false

	// Apply custom limits from the license if they're set; otherwise use tier defaults
	if m.license.Limits == (Limits{}) {
		m.license.Limits = DefaultLimits(m.license.Tier)
	}

	return nil
}

// SetCommunityEdition resets to the free community edition.
func (m *Manager) SetCommunityEdition() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.license = &License{
		Type:     LicenseTypeSubscription,
		Tier:     TierFree,
		Limits:   DefaultLimits(TierFree),
		Features: FeaturesForTier(TierFree),
	}
	m.licenseKey = ""
	m.isCommunity = true
}

// RemoveLicense removes the current license and reverts to CE.
func (m *Manager) RemoveLicense() {
	m.SetCommunityEdition()
}

// GetLicense returns the current license (read-only snapshot).
func (m *Manager) GetLicense() *License {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid races
	lic := *m.license
	return &lic
}

// GetTier returns the current license tier.
func (m *Manager) GetTier() Tier {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license.Tier
}

// IsCommunity returns true if running in community edition mode.
func (m *Manager) IsCommunity() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isCommunity
}

// GetLimits returns the current license limits.
func (m *Manager) GetLimits() Limits {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license.Limits
}

// HasFeature checks whether the current license grants a feature.
func (m *Manager) HasFeature(f Feature) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license.HasFeature(f)
}

// CheckLimit returns true if the current usage is within the license limit.
func (m *Manager) CheckLimit(lt LimitType, current int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license.Limits.CheckLimit(lt, current)
}

// StartPeriodicCheck re-verifies the license at the given interval.
// If the license becomes invalid or expired, it falls back to CE.
func (m *Manager) StartPeriodicCheck(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.mu.RLock()
				key := m.licenseKey
				community := m.isCommunity
				m.mu.RUnlock()

				if community || key == "" {
					continue
				}

				if _, err := m.verifier.VerifyString(key); err != nil {
					m.logger.Warn("license verification failed, reverting to community edition",
						zap.Error(err),
					)
					m.SetCommunityEdition()
				}
			}
		}
	}()
}

// GetLicenseResponse returns a safe API response for the current license.
func (m *Manager) GetLicenseResponse() *LicenseResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.license.ToResponse(m.isCommunity)
}
