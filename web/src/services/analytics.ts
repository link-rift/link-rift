import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  LinkAnalytics,
  WorkspaceAnalytics,
  TimeSeriesPoint,
  ReferrerStats,
  CountryStats,
  DeviceBreakdown,
  BrowserStats,
  DateRangePreset,
  TimeSeriesInterval,
  DateRange,
  ExportFormat,
} from "@/types/analytics"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function analyticsBase(): string {
  return `/workspaces/${getWorkspaceId()}/analytics`
}

function buildDateParams(
  range_?: DateRangePreset,
  dateRange?: DateRange
): URLSearchParams {
  const params = new URLSearchParams()
  if (range_) {
    params.set("range", range_)
  } else if (dateRange) {
    params.set("start", dateRange.start)
    params.set("end", dateRange.end)
  }
  return params
}

export async function getLinkStats(
  linkId: string,
  range_?: DateRangePreset,
  dateRange?: DateRange
): Promise<LinkAnalytics> {
  const params = buildDateParams(range_, dateRange)
  const qs = params.toString()
  const url = `${analyticsBase()}/links/${linkId}${qs ? `?${qs}` : ""}`
  const res = await apiRequest<LinkAnalytics>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch link stats")
  }
  return res.data
}

export async function getTimeSeries(
  linkId: string,
  interval: TimeSeriesInterval = "day",
  range_?: DateRangePreset,
  dateRange?: DateRange
): Promise<TimeSeriesPoint[]> {
  const params = buildDateParams(range_, dateRange)
  params.set("interval", interval)
  const url = `${analyticsBase()}/links/${linkId}/timeseries?${params}`
  const res = await apiRequest<TimeSeriesPoint[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch time series")
  }
  return res.data
}

export async function getReferrers(
  linkId: string,
  range_?: DateRangePreset,
  dateRange?: DateRange,
  limit = 10
): Promise<ReferrerStats[]> {
  const params = buildDateParams(range_, dateRange)
  params.set("limit", String(limit))
  const url = `${analyticsBase()}/links/${linkId}/referrers?${params}`
  const res = await apiRequest<ReferrerStats[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch referrers")
  }
  return res.data
}

export async function getCountries(
  linkId: string,
  range_?: DateRangePreset,
  dateRange?: DateRange,
  limit = 10
): Promise<CountryStats[]> {
  const params = buildDateParams(range_, dateRange)
  params.set("limit", String(limit))
  const url = `${analyticsBase()}/links/${linkId}/countries?${params}`
  const res = await apiRequest<CountryStats[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch countries")
  }
  return res.data
}

export async function getDevices(
  linkId: string,
  range_?: DateRangePreset,
  dateRange?: DateRange
): Promise<DeviceBreakdown> {
  const params = buildDateParams(range_, dateRange)
  const qs = params.toString()
  const url = `${analyticsBase()}/links/${linkId}/devices${qs ? `?${qs}` : ""}`
  const res = await apiRequest<DeviceBreakdown>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch devices")
  }
  return res.data
}

export async function getBrowsers(
  linkId: string,
  range_?: DateRangePreset,
  dateRange?: DateRange,
  limit = 10
): Promise<BrowserStats[]> {
  const params = buildDateParams(range_, dateRange)
  params.set("limit", String(limit))
  const url = `${analyticsBase()}/links/${linkId}/browsers?${params}`
  const res = await apiRequest<BrowserStats[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch browsers")
  }
  return res.data
}

export async function getWorkspaceStats(
  range_?: DateRangePreset,
  dateRange?: DateRange
): Promise<WorkspaceAnalytics> {
  const params = buildDateParams(range_, dateRange)
  const qs = params.toString()
  const url = `${analyticsBase()}/workspace${qs ? `?${qs}` : ""}`
  const res = await apiRequest<WorkspaceAnalytics>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch workspace stats")
  }
  return res.data
}

export async function exportData(
  linkId: string,
  format: ExportFormat = "csv",
  range_?: DateRangePreset,
  dateRange?: DateRange
): Promise<Blob> {
  const params = buildDateParams(range_, dateRange)
  params.set("format", format)
  params.set("link_id", linkId)
  const url = `${analyticsBase()}/export?${params}`

  const token = localStorage.getItem("access_token")
  const headers: Record<string, string> = {}
  if (token) headers["Authorization"] = `Bearer ${token}`

  const response = await fetch(`/api/v1${url}`, { headers })
  if (!response.ok) {
    throw new Error("Failed to export data")
  }
  return response.blob()
}
