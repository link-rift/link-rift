export interface Link {
  id: string
  user_id: string
  workspace_id: string
  domain_id?: string | null
  url: string
  short_code: string
  short_url: string
  title?: string | null
  description?: string | null
  favicon_url?: string | null
  og_image_url?: string | null
  is_active: boolean
  has_password: boolean
  expires_at?: string | null
  max_clicks?: number | null
  utm_source?: string | null
  utm_medium?: string | null
  utm_campaign?: string | null
  utm_term?: string | null
  utm_content?: string | null
  total_clicks: number
  unique_clicks: number
  created_at: string
  updated_at: string
}

export interface CreateLinkRequest {
  url: string
  short_code?: string
  title?: string
  description?: string
  password?: string
  expires_at?: string
  max_clicks?: number
  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
  utm_term?: string
  utm_content?: string
}

export interface UpdateLinkRequest {
  url?: string
  title?: string
  description?: string
  is_active?: boolean
  password?: string
  expires_at?: string
  max_clicks?: number
}

export interface BulkCreateRequest {
  links: CreateLinkRequest[]
}

export interface LinkFilter {
  search?: string
  is_active?: boolean
}

export interface LinkQuickStats {
  total_clicks: number
  unique_clicks: number
  clicks_24h: number
  clicks_7d: number
  created_at: string
}
