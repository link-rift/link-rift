export interface AuditLog {
  id: string
  workspace_id: string
  user_id?: string
  action: string
  resource_type: string
  resource_id?: string
  old_values?: Record<string, unknown>
  new_values?: Record<string, unknown>
  metadata?: Record<string, unknown>
  ip_address?: string
  user_agent?: string
  created_at: string
}

export interface AuditLogFilter {
  action?: string
  resource_type?: string
  user_id?: string
  start_date?: string
  end_date?: string
}

export interface AuditLogListResult {
  audit_logs: AuditLog[]
  total: number
}

export const AUDIT_ACTIONS = [
  "create",
  "update",
  "delete",
  "revoke",
] as const

export const AUDIT_RESOURCE_TYPES = [
  "link",
  "workspace",
  "member",
  "domain",
  "api_key",
  "webhook",
  "bio_page",
  "branding",
  "sso_config",
  "scim_token",
] as const

export const ACTION_LABELS: Record<string, string> = {
  create: "Created",
  update: "Updated",
  delete: "Deleted",
  revoke: "Revoked",
}

export const RESOURCE_TYPE_LABELS: Record<string, string> = {
  link: "Link",
  workspace: "Workspace",
  member: "Member",
  domain: "Domain",
  api_key: "API Key",
  webhook: "Webhook",
  bio_page: "Bio Page",
  branding: "Branding",
  sso_config: "SSO Config",
  scim_token: "SCIM Token",
}
