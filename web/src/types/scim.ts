export interface SCIMToken {
  id: string
  workspace_id: string
  token_prefix: string
  name: string
  is_active: boolean
  last_used_at?: string | null
  created_at: string
  expires_at?: string | null
}

export interface CreateSCIMTokenInput {
  name: string
  expires_at?: string
}

export interface CreateSCIMTokenResponse {
  token: string
  scim_token: SCIMToken
}

export interface SCIMSyncLog {
  id: string
  workspace_id: string
  operation: string
  resource_type: string
  external_id?: string
  status: string
  details: Record<string, unknown>
  created_at: string
}
