export type Tier = "free" | "pro" | "business" | "enterprise"
export type LicenseType = "trial" | "subscription" | "perpetual" | "enterprise"

export type Feature =
  | "custom_domains"
  | "link_expiration"
  | "link_passwords"
  | "bulk_links"
  | "advanced_analytics"
  | "export_data"
  | "team_members"
  | "multi_workspace"
  | "api_access"
  | "webhooks"
  | "qr_customization"
  | "bio_pages"
  | "conditional_routing"
  | "saml"
  | "scim"
  | "audit_logs"
  | "white_label"
  | "custom_css"
  | "priority_support"
  | "sla"

export interface LicenseLimits {
  max_users: number
  max_domains: number
  max_links_per_month: number
  max_clicks_per_month: number
  max_workspaces: number
  max_api_requests_per_min: number
}

export interface Plan {
  tier: Tier
  name: string
  description: string
  price: string
}

export interface LicenseInfo {
  id?: string
  customer_name?: string
  type: LicenseType
  tier: Tier
  plan: Plan
  expires_at?: string | null
  features: Feature[]
  limits: LicenseLimits
  is_community: boolean
}

export interface ActivateLicenseRequest {
  license_key: string
}

export const TIER_LEVELS: Record<Tier, number> = {
  free: 1,
  pro: 2,
  business: 3,
  enterprise: 4,
}

export const FEATURE_DISPLAY_NAMES: Record<Feature, string> = {
  custom_domains: "Custom Domains",
  link_expiration: "Link Expiration",
  link_passwords: "Password-Protected Links",
  bulk_links: "Bulk Link Creation",
  advanced_analytics: "Advanced Analytics",
  export_data: "Data Export",
  team_members: "Team Members",
  multi_workspace: "Multiple Workspaces",
  api_access: "API Access",
  webhooks: "Webhooks",
  qr_customization: "QR Code Customization",
  bio_pages: "Bio Pages",
  conditional_routing: "Conditional Routing",
  saml: "SAML SSO",
  scim: "SCIM Provisioning",
  audit_logs: "Audit Logs",
  white_label: "White Label",
  custom_css: "Custom CSS",
  priority_support: "Priority Support",
  sla: "SLA",
}
