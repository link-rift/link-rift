export interface APIKey {
  id: string
  user_id: string
  workspace_id: string
  name: string
  key_prefix: string
  scopes: string[]
  last_used_at?: string | null
  request_count: number
  rate_limit?: number | null
  expires_at?: string | null
  created_at: string
}

export interface CreateAPIKeyRequest {
  name: string
  scopes: string[]
  expires_at?: string
}

export interface CreateAPIKeyResponse {
  api_key: APIKey
  key: string
}

export const API_KEY_SCOPES = [
  { value: "links:read", label: "Links (Read)", description: "View links" },
  { value: "links:write", label: "Links (Write)", description: "Create, update, delete links" },
  { value: "domains:read", label: "Domains (Read)", description: "View custom domains" },
  { value: "domains:write", label: "Domains (Write)", description: "Manage custom domains" },
  { value: "analytics:read", label: "Analytics (Read)", description: "View analytics data" },
  { value: "bio_pages:read", label: "Bio Pages (Read)", description: "View bio pages" },
  { value: "bio_pages:write", label: "Bio Pages (Write)", description: "Manage bio pages" },
  { value: "qr:read", label: "QR Codes (Read)", description: "View QR codes" },
  { value: "qr:write", label: "QR Codes (Write)", description: "Generate QR codes" },
  { value: "webhooks:read", label: "Webhooks (Read)", description: "View webhooks" },
  { value: "webhooks:write", label: "Webhooks (Write)", description: "Manage webhooks" },
] as const
