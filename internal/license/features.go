package license

// Feature represents a gated feature identifier.
type Feature string

const (
	FeatureCustomDomains     Feature = "custom_domains"
	FeatureLinkExpiration    Feature = "link_expiration"
	FeatureLinkPasswords     Feature = "link_passwords"
	FeatureBulkLinks         Feature = "bulk_links"
	FeatureAdvancedAnalytics Feature = "advanced_analytics"
	FeatureExportData        Feature = "export_data"
	FeatureTeamMembers       Feature = "team_members"
	FeatureMultiWorkspace    Feature = "multi_workspace"
	FeatureAPIAccess         Feature = "api_access"
	FeatureWebhooks          Feature = "webhooks"
	FeatureQRCustomization   Feature = "qr_customization"
	FeatureBioPages          Feature = "bio_pages"
	FeatureConditionalRouting Feature = "conditional_routing"
	FeatureSAML              Feature = "saml"
	FeatureSCIM              Feature = "scim"
	FeatureAuditLogs         Feature = "audit_logs"
	FeatureWhiteLabel        Feature = "white_label"
	FeatureCustomCSS         Feature = "custom_css"
	FeaturePrioritySupport   Feature = "priority_support"
	FeatureSLA               Feature = "sla"
)

// FeatureDefinition describes a feature and its minimum tier requirement.
type FeatureDefinition struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MinTier     Tier    `json:"min_tier"`
	Category    string  `json:"category"`
}

var featureRegistry = map[Feature]FeatureDefinition{
	FeatureCustomDomains: {
		Name:        "Custom Domains",
		Description: "Connect your own domains for branded short links",
		MinTier:     TierPro,
		Category:    "links",
	},
	FeatureLinkExpiration: {
		Name:        "Link Expiration",
		Description: "Set expiration dates for short links",
		MinTier:     TierFree,
		Category:    "links",
	},
	FeatureLinkPasswords: {
		Name:        "Password-Protected Links",
		Description: "Require a password to access links",
		MinTier:     TierPro,
		Category:    "links",
	},
	FeatureBulkLinks: {
		Name:        "Bulk Link Creation",
		Description: "Create multiple links at once via CSV or API",
		MinTier:     TierPro,
		Category:    "links",
	},
	FeatureAdvancedAnalytics: {
		Name:        "Advanced Analytics",
		Description: "Detailed click analytics with geographic and device breakdowns",
		MinTier:     TierPro,
		Category:    "analytics",
	},
	FeatureExportData: {
		Name:        "Data Export",
		Description: "Export analytics data as CSV or JSON",
		MinTier:     TierPro,
		Category:    "analytics",
	},
	FeatureTeamMembers: {
		Name:        "Team Members",
		Description: "Invite team members with role-based access",
		MinTier:     TierBusiness,
		Category:    "team",
	},
	FeatureMultiWorkspace: {
		Name:        "Multiple Workspaces",
		Description: "Create and manage multiple workspaces",
		MinTier:     TierBusiness,
		Category:    "team",
	},
	FeatureAPIAccess: {
		Name:        "API Access",
		Description: "Programmatic access via REST API keys",
		MinTier:     TierPro,
		Category:    "developer",
	},
	FeatureWebhooks: {
		Name:        "Webhooks",
		Description: "Receive real-time event notifications",
		MinTier:     TierBusiness,
		Category:    "developer",
	},
	FeatureQRCustomization: {
		Name:        "QR Code Customization",
		Description: "Custom colors, logos, and styles for QR codes",
		MinTier:     TierPro,
		Category:    "links",
	},
	FeatureBioPages: {
		Name:        "Bio Pages",
		Description: "Create link-in-bio pages",
		MinTier:     TierPro,
		Category:    "pages",
	},
	FeatureConditionalRouting: {
		Name:        "Conditional Routing",
		Description: "Route clicks based on device, location, or time rules",
		MinTier:     TierBusiness,
		Category:    "links",
	},
	FeatureSAML: {
		Name:        "SAML SSO",
		Description: "Enterprise single sign-on via SAML 2.0",
		MinTier:     TierEnterprise,
		Category:    "security",
	},
	FeatureSCIM: {
		Name:        "SCIM Provisioning",
		Description: "Automated user provisioning via SCIM 2.0",
		MinTier:     TierEnterprise,
		Category:    "security",
	},
	FeatureAuditLogs: {
		Name:        "Audit Logs",
		Description: "Detailed audit trail of all actions",
		MinTier:     TierEnterprise,
		Category:    "security",
	},
	FeatureWhiteLabel: {
		Name:        "White Label",
		Description: "Remove Linkrift branding and add your own",
		MinTier:     TierEnterprise,
		Category:    "branding",
	},
	FeatureCustomCSS: {
		Name:        "Custom CSS",
		Description: "Inject custom CSS for bio pages and redirects",
		MinTier:     TierEnterprise,
		Category:    "branding",
	},
	FeaturePrioritySupport: {
		Name:        "Priority Support",
		Description: "Priority email and chat support",
		MinTier:     TierBusiness,
		Category:    "support",
	},
	FeatureSLA: {
		Name:        "SLA",
		Description: "Guaranteed uptime and response time SLA",
		MinTier:     TierEnterprise,
		Category:    "support",
	},
}

// GetFeatureDefinition returns the definition for a feature.
func GetFeatureDefinition(f Feature) (FeatureDefinition, bool) {
	def, ok := featureRegistry[f]
	return def, ok
}

// AllFeatures returns all registered feature definitions.
func AllFeatures() map[Feature]FeatureDefinition {
	result := make(map[Feature]FeatureDefinition, len(featureRegistry))
	for k, v := range featureRegistry {
		result[k] = v
	}
	return result
}

// FeaturesForTier returns all features available at the given tier.
func FeaturesForTier(tier Tier) []Feature {
	var features []Feature
	for f, def := range featureRegistry {
		if tier.IncludesTier(def.MinTier) {
			features = append(features, f)
		}
	}
	return features
}
