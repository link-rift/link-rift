import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  Link,
  CreateLinkRequest,
  UpdateLinkRequest,
  BulkCreateRequest,
  LinkQuickStats,
} from "@/types/link"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/links`
}

export async function createLink(data: CreateLinkRequest): Promise<Link> {
  const res = await apiRequest<Link>(wsBase(), {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create link")
  }
  return res.data
}

export async function getLinks(params?: {
  search?: string
  is_active?: boolean
  limit?: number
  offset?: number
}): Promise<{ links: Link[]; total: number }> {
  const searchParams = new URLSearchParams()
  if (params?.search) searchParams.set("search", params.search)
  if (params?.is_active !== undefined) searchParams.set("is_active", String(params.is_active))
  if (params?.limit) searchParams.set("limit", String(params.limit))
  if (params?.offset) searchParams.set("offset", String(params.offset))

  const qs = searchParams.toString()
  const url = qs ? `${wsBase()}?${qs}` : wsBase()
  const res = await apiRequest<Link[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch links")
  }
  return {
    links: res.data,
    total: res.meta?.total ?? res.data.length,
  }
}

export async function getLink(id: string): Promise<Link> {
  const res = await apiRequest<Link>(`${wsBase()}/${id}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch link")
  }
  return res.data
}

export async function updateLink(id: string, data: UpdateLinkRequest): Promise<Link> {
  const res = await apiRequest<Link>(`${wsBase()}/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update link")
  }
  return res.data
}

export async function deleteLink(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete link")
  }
}

export async function bulkCreateLinks(data: BulkCreateRequest): Promise<Link[]> {
  const res = await apiRequest<Link[]>(`${wsBase()}/bulk`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create links")
  }
  return res.data
}

export async function getLinkStats(id: string): Promise<LinkQuickStats> {
  const res = await apiRequest<LinkQuickStats>(`${wsBase()}/${id}/stats`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch link stats")
  }
  return res.data
}
