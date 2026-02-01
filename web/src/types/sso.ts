export interface SSOConfig {
  id: string
  workspace_id: string
  provider: string
  entity_id: string
  sso_url: string
  slo_url?: string
  certificate: string
  idp_metadata_url?: string
  attribute_mapping: Record<string, string>
  is_enabled: boolean
  enforce_sso: boolean
  allowed_domains: string[]
  created_at: string
  updated_at: string
}

export interface CreateSSOConfigInput {
  provider: string
  entity_id: string
  sso_url: string
  slo_url?: string
  certificate: string
  idp_metadata_url?: string
  idp_metadata_xml?: string
  attribute_mapping?: Record<string, string>
  is_enabled?: boolean
  enforce_sso?: boolean
  allowed_domains?: string[]
}

export interface UpdateSSOConfigInput {
  provider?: string
  entity_id?: string
  sso_url?: string
  slo_url?: string
  certificate?: string
  idp_metadata_url?: string
  idp_metadata_xml?: string
  attribute_mapping?: Record<string, string>
  is_enabled?: boolean
  enforce_sso?: boolean
  allowed_domains?: string[]
}

export interface SSOIdentity {
  id: string
  user_id: string
  workspace_id: string
  provider: string
  external_id: string
  email: string
  name?: string
  last_login_at?: string
  created_at: string
  updated_at: string
}
