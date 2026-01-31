export interface QRCode {
  id: string
  link_id: string
  qr_type: string
  error_correction: string
  foreground_color: string
  background_color: string
  logo_url?: string | null
  png_url?: string | null
  svg_url?: string | null
  dot_style: string
  corner_style: string
  size: number
  margin: number
  scan_count: number
  created_at: string
  updated_at: string
}

export interface CreateQRCodeRequest {
  qr_type?: string
  error_correction?: string
  foreground_color?: string
  background_color?: string
  logo_url?: string
  dot_style?: string
  corner_style?: string
  size?: number
  margin?: number
}

export interface BulkQRCodeRequest {
  link_ids: string[]
  options: CreateQRCodeRequest
}

export interface QRStyleTemplate {
  name: string
  foreground_color: string
  background_color: string
  dot_style: string
  corner_style: string
  error_correction: string
}
