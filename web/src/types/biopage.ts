export interface BioPage {
  id: string
  workspace_id: string
  slug: string
  title: string
  bio?: string | null
  avatar_url?: string | null
  theme_id?: string | null
  custom_css?: string | null
  meta_title?: string | null
  meta_description?: string | null
  og_image_url?: string | null
  is_published: boolean
  created_at: string
  updated_at: string
  links?: BioPageLink[]
  link_count?: number
}

export interface BioPageLink {
  id: string
  bio_page_id: string
  title: string
  url: string
  icon?: string | null
  position: number
  is_visible: boolean
  visible_from?: string | null
  visible_until?: string | null
  click_count: number
  created_at: string
  updated_at: string
}

export interface BioPageTheme {
  id: string
  name: string
  description: string
  is_premium: boolean
  styles: ThemeStyles
}

export interface ThemeStyles {
  background_color: string
  text_color: string
  button_color: string
  button_text_color: string
  button_style: string
  font_family: string
  gradient?: GradientConfig | null
}

export interface GradientConfig {
  from: string
  to: string
  direction: string
}

export interface CreateBioPageRequest {
  title: string
  slug: string
  bio?: string
  avatar_url?: string
  theme_id?: string
  meta_title?: string
  meta_description?: string
  og_image_url?: string
}

export interface UpdateBioPageRequest {
  title?: string
  slug?: string
  bio?: string
  avatar_url?: string
  theme_id?: string
  custom_css?: string
  meta_title?: string
  meta_description?: string
  og_image_url?: string
}

export interface CreateBioPageLinkRequest {
  title: string
  url: string
  icon?: string
  is_visible?: boolean
  visible_from?: string
  visible_until?: string
}

export interface UpdateBioPageLinkRequest {
  title?: string
  url?: string
  icon?: string
  is_visible?: boolean
  visible_from?: string
  visible_until?: string
}

export interface ReorderBioLinksRequest {
  link_ids: string[]
}

export interface PublicBioPage {
  title: string
  bio?: string | null
  avatar_url?: string | null
  slug: string
  theme?: BioPageTheme | null
  custom_css?: string | null
  meta_title?: string | null
  meta_description?: string | null
  og_image_url?: string | null
  links: PublicBioLink[]
}

export interface PublicBioLink {
  id: string
  title: string
  url: string
  icon?: string | null
}
