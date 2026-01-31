package license

// Tier represents a subscription tier level.
type Tier string

const (
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierBusiness   Tier = "business"
	TierEnterprise Tier = "enterprise"
)

var tierLevels = map[Tier]int{
	TierFree:       1,
	TierPro:        2,
	TierBusiness:   3,
	TierEnterprise: 4,
}

// Level returns the numeric level of the tier (Free=1, Pro=2, Business=3, Enterprise=4).
func (t Tier) Level() int {
	if lvl, ok := tierLevels[t]; ok {
		return lvl
	}
	return 0
}

// IncludesTier returns true if this tier includes (is >= ) the given tier.
func (t Tier) IncludesTier(other Tier) bool {
	return t.Level() >= other.Level()
}

// IsValid returns true if the tier is a known tier.
func (t Tier) IsValid() bool {
	_, ok := tierLevels[t]
	return ok
}

// Plan holds display metadata for a tier.
type Plan struct {
	Tier        Tier   `json:"tier"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
}

// Plans maps each tier to its display metadata.
var Plans = map[Tier]Plan{
	TierFree: {
		Tier:        TierFree,
		Name:        "Community",
		Description: "Free self-hosted edition with core features",
		Price:       "Free",
	},
	TierPro: {
		Tier:        TierPro,
		Name:        "Pro",
		Description: "For professionals and small teams",
		Price:       "$19/mo",
	},
	TierBusiness: {
		Tier:        TierBusiness,
		Name:        "Business",
		Description: "For growing teams with advanced needs",
		Price:       "$49/mo",
	},
	TierEnterprise: {
		Tier:        TierEnterprise,
		Name:        "Enterprise",
		Description: "Custom solutions for large organizations",
		Price:       "Custom",
	},
}
