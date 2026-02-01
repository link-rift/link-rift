export interface WorkspaceBranding {
  id: string
  workspace_id: string
  logo_url?: string
  logo_dark_url?: string
  favicon_url?: string
  primary_color?: string
  secondary_color?: string
  accent_color?: string
  custom_css?: string
  hide_powered_by: boolean
  custom_footer_text?: string
  custom_footer_url?: string
  created_at: string
  updated_at: string
}

export interface UpdateBrandingInput {
  logo_url?: string | null
  logo_dark_url?: string | null
  favicon_url?: string | null
  primary_color?: string | null
  secondary_color?: string | null
  accent_color?: string | null
  custom_css?: string | null
  hide_powered_by?: boolean
  custom_footer_text?: string | null
  custom_footer_url?: string | null
}

export interface PublicBranding {
  logo_url?: string
  logo_dark_url?: string
  favicon_url?: string
  primary_color?: string
  secondary_color?: string
  accent_color?: string
  custom_css?: string
  hide_powered_by: boolean
  custom_footer_text?: string
  custom_footer_url?: string
}
