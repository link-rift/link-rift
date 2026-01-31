export interface Webhook {
  id: string
  workspace_id: string
  url: string
  events: string[]
  is_active: boolean
  failure_count: number
  last_triggered_at?: string | null
  last_success_at?: string | null
  created_at: string
  updated_at: string
}

export interface WebhookDelivery {
  id: string
  webhook_id: string
  event: string
  payload: Record<string, unknown>
  response_status?: number | null
  response_body?: string | null
  attempts: number
  max_attempts: number
  last_attempt_at?: string | null
  completed_at?: string | null
  created_at: string
}

export interface CreateWebhookRequest {
  url: string
  events: string[]
}

export interface CreateWebhookResponse {
  webhook: Webhook
  secret: string
}

export const WEBHOOK_EVENTS = [
  { value: "link.created", label: "Link Created", category: "Links" },
  { value: "link.updated", label: "Link Updated", category: "Links" },
  { value: "link.deleted", label: "Link Deleted", category: "Links" },
  { value: "link.clicked", label: "Link Clicked", category: "Links" },
  { value: "link.expired", label: "Link Expired", category: "Links" },
  { value: "qr.created", label: "QR Code Created", category: "QR Codes" },
  { value: "qr.scanned", label: "QR Code Scanned", category: "QR Codes" },
  { value: "biopage.created", label: "Bio Page Created", category: "Bio Pages" },
  { value: "biopage.updated", label: "Bio Page Updated", category: "Bio Pages" },
  { value: "domain.added", label: "Domain Added", category: "Domains" },
  { value: "domain.verified", label: "Domain Verified", category: "Domains" },
  { value: "domain.removed", label: "Domain Removed", category: "Domains" },
  { value: "team.member_invited", label: "Member Invited", category: "Team" },
  { value: "team.member_joined", label: "Member Joined", category: "Team" },
  { value: "team.member_removed", label: "Member Removed", category: "Team" },
] as const
