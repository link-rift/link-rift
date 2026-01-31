package license

// LimitType identifies a resource limit.
type LimitType string

const (
	LimitMaxUsers           LimitType = "max_users"
	LimitMaxDomains         LimitType = "max_domains"
	LimitMaxLinksPerMonth   LimitType = "max_links_per_month"
	LimitMaxClicksPerMonth  LimitType = "max_clicks_per_month"
	LimitMaxWorkspaces      LimitType = "max_workspaces"
	LimitMaxAPIRequestsPerMin LimitType = "max_api_requests_per_min"
)

// Limits holds usage limits for a license tier.
type Limits struct {
	MaxUsers             int64 `json:"max_users"`
	MaxDomains           int64 `json:"max_domains"`
	MaxLinksPerMonth     int64 `json:"max_links_per_month"`
	MaxClicksPerMonth    int64 `json:"max_clicks_per_month"`
	MaxWorkspaces        int64 `json:"max_workspaces"`
	MaxAPIRequestsPerMin int64 `json:"max_api_requests_per_min"`
}

var defaultLimits = map[Tier]Limits{
	TierFree: {
		MaxUsers:             1,
		MaxDomains:           0,
		MaxLinksPerMonth:     100,
		MaxClicksPerMonth:    10000,
		MaxWorkspaces:        1,
		MaxAPIRequestsPerMin: 10,
	},
	TierPro: {
		MaxUsers:             5,
		MaxDomains:           3,
		MaxLinksPerMonth:     5000,
		MaxClicksPerMonth:    500000,
		MaxWorkspaces:        3,
		MaxAPIRequestsPerMin: 60,
	},
	TierBusiness: {
		MaxUsers:             25,
		MaxDomains:           10,
		MaxLinksPerMonth:     50000,
		MaxClicksPerMonth:    5000000,
		MaxWorkspaces:        10,
		MaxAPIRequestsPerMin: 300,
	},
	TierEnterprise: {
		MaxUsers:             -1, // unlimited
		MaxDomains:           -1,
		MaxLinksPerMonth:     -1,
		MaxClicksPerMonth:    -1,
		MaxWorkspaces:        -1,
		MaxAPIRequestsPerMin: 1000,
	},
}

// DefaultLimits returns the default limits for a tier.
func DefaultLimits(tier Tier) Limits {
	if l, ok := defaultLimits[tier]; ok {
		return l
	}
	return defaultLimits[TierFree]
}

// GetLimit returns the limit value for a specific limit type.
func (l Limits) GetLimit(lt LimitType) int64 {
	switch lt {
	case LimitMaxUsers:
		return l.MaxUsers
	case LimitMaxDomains:
		return l.MaxDomains
	case LimitMaxLinksPerMonth:
		return l.MaxLinksPerMonth
	case LimitMaxClicksPerMonth:
		return l.MaxClicksPerMonth
	case LimitMaxWorkspaces:
		return l.MaxWorkspaces
	case LimitMaxAPIRequestsPerMin:
		return l.MaxAPIRequestsPerMin
	default:
		return 0
	}
}

// CheckLimit returns true if the current usage is within the limit.
// A limit of -1 means unlimited.
func (l Limits) CheckLimit(lt LimitType, current int64) bool {
	limit := l.GetLimit(lt)
	if limit < 0 {
		return true // unlimited
	}
	return current < limit
}
