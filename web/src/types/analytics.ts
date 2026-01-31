export type DateRangePreset = "24h" | "7d" | "30d" | "90d"
export type TimeSeriesInterval = "hour" | "day" | "week" | "month"
export type ExportFormat = "csv" | "json"

export interface DateRange {
  start: string
  end: string
}

export interface LinkAnalytics {
  total_clicks: number
  unique_clicks: number
  clicks_24h: number
  clicks_7d: number
  clicks_30d: number
}

export interface WorkspaceAnalytics {
  total_links: number
  total_clicks: number
  unique_clicks: number
  clicks_24h: number
  clicks_7d: number
  clicks_30d: number
  top_links: TopLink[]
}

export interface TopLink {
  link_id: string
  short_code: string
  total_clicks: number
}

export interface TimeSeriesPoint {
  timestamp: string
  clicks: number
  unique: number
}

export interface ReferrerStats {
  referrer: string
  clicks: number
  percent: number
}

export interface CountryStats {
  country_code: string
  country: string
  clicks: number
  percent: number
}

export interface DeviceBreakdown {
  desktop: number
  mobile: number
  tablet: number
  other: number
}

export interface BrowserStats {
  browser: string
  clicks: number
  percent: number
}

export interface ClickNotification {
  workspace_id: string
  link_id: string
  short_code: string
  timestamp: string
  country_code?: string
  device_type?: string
  browser?: string
  referer?: string
}
