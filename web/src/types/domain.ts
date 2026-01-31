export interface Domain {
  id: string
  workspace_id: string
  domain: string
  is_verified: boolean
  verified_at?: string | null
  ssl_status: DomainSSLStatus
  ssl_expires_at?: string | null
  dns_records?: Record<string, unknown> | null
  last_dns_check_at?: string | null
  default_redirect_url?: string | null
  custom_404_url?: string | null
  created_at: string
  updated_at: string
}

export type DomainSSLStatus = "pending" | "active" | "failed"

export interface CreateDomainRequest {
  domain: string
}

export interface DNSRecordInstruction {
  type: string
  host: string
  value: string
}

export interface VerificationInstructions {
  records: DNSRecordInstruction[]
}
