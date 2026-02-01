import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { AuditLog, AuditLogFilter } from "@/types/auditlog"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/audit-logs`
}

export async function getAuditLogs(params?: {
  action?: string
  resource_type?: string
  user_id?: string
  start_date?: string
  end_date?: string
  limit?: number
  offset?: number
}): Promise<{ audit_logs: AuditLog[]; total: number }> {
  const searchParams = new URLSearchParams()
  if (params?.action) searchParams.set("action", params.action)
  if (params?.resource_type) searchParams.set("resource_type", params.resource_type)
  if (params?.user_id) searchParams.set("user_id", params.user_id)
  if (params?.start_date) searchParams.set("start_date", params.start_date)
  if (params?.end_date) searchParams.set("end_date", params.end_date)
  if (params?.limit) searchParams.set("limit", String(params.limit))
  if (params?.offset) searchParams.set("offset", String(params.offset))

  const qs = searchParams.toString()
  const url = qs ? `${wsBase()}?${qs}` : wsBase()
  const res = await apiRequest<AuditLog[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch audit logs")
  }
  return {
    audit_logs: res.data,
    total: res.meta?.total ?? res.data.length,
  }
}

export async function getAuditLog(id: string): Promise<AuditLog> {
  const res = await apiRequest<AuditLog>(`${wsBase()}/${id}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch audit log")
  }
  return res.data
}

export async function exportAuditLogs(
  filter: AuditLogFilter,
  format: "json" | "csv" = "json"
): Promise<Blob> {
  const searchParams = new URLSearchParams()
  if (filter.action) searchParams.set("action", filter.action)
  if (filter.resource_type) searchParams.set("resource_type", filter.resource_type)
  if (filter.user_id) searchParams.set("user_id", filter.user_id)
  if (filter.start_date) searchParams.set("start_date", filter.start_date)
  if (filter.end_date) searchParams.set("end_date", filter.end_date)
  searchParams.set("format", format)

  const qs = searchParams.toString()
  const url = `/api/v1${wsBase()}/export?${qs}`
  const token = localStorage.getItem("access_token")

  const response = await fetch(url, {
    headers: {
      Authorization: token ? `Bearer ${token}` : "",
    },
  })

  if (!response.ok) {
    throw new Error("Failed to export audit logs")
  }

  return response.blob()
}
