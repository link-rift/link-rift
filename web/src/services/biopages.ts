import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  BioPage,
  BioPageLink,
  BioPageTheme,
  CreateBioPageRequest,
  UpdateBioPageRequest,
  CreateBioPageLinkRequest,
  UpdateBioPageLinkRequest,
  ReorderBioLinksRequest,
  PublicBioPage,
} from "@/types/biopage"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/bio-pages`
}

// Bio Pages CRUD

export async function createBioPage(data: CreateBioPageRequest): Promise<BioPage> {
  const res = await apiRequest<BioPage>(wsBase(), {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create bio page")
  }
  return res.data
}

export async function getBioPages(): Promise<BioPage[]> {
  const res = await apiRequest<BioPage[]>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch bio pages")
  }
  return res.data
}

export async function getBioPage(id: string): Promise<BioPage> {
  const res = await apiRequest<BioPage>(`${wsBase()}/${id}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch bio page")
  }
  return res.data
}

export async function updateBioPage(id: string, data: UpdateBioPageRequest): Promise<BioPage> {
  const res = await apiRequest<BioPage>(`${wsBase()}/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update bio page")
  }
  return res.data
}

export async function deleteBioPage(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete bio page")
  }
}

// Publish

export async function publishBioPage(id: string): Promise<BioPage> {
  const res = await apiRequest<BioPage>(`${wsBase()}/${id}/publish`, {
    method: "POST",
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to publish bio page")
  }
  return res.data
}

export async function unpublishBioPage(id: string): Promise<BioPage> {
  const res = await apiRequest<BioPage>(`${wsBase()}/${id}/unpublish`, {
    method: "POST",
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to unpublish bio page")
  }
  return res.data
}

// Links

export async function getBioPageLinks(pageId: string): Promise<BioPageLink[]> {
  const res = await apiRequest<BioPageLink[]>(`${wsBase()}/${pageId}/links`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch links")
  }
  return res.data
}

export async function addBioPageLink(pageId: string, data: CreateBioPageLinkRequest): Promise<BioPageLink> {
  const res = await apiRequest<BioPageLink>(`${wsBase()}/${pageId}/links`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to add link")
  }
  return res.data
}

export async function updateBioPageLink(
  pageId: string,
  linkId: string,
  data: UpdateBioPageLinkRequest
): Promise<BioPageLink> {
  const res = await apiRequest<BioPageLink>(`${wsBase()}/${pageId}/links/${linkId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update link")
  }
  return res.data
}

export async function deleteBioPageLink(pageId: string, linkId: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${pageId}/links/${linkId}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete link")
  }
}

export async function reorderBioPageLinks(pageId: string, data: ReorderBioLinksRequest): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${pageId}/links/reorder`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to reorder links")
  }
}

// Themes

export async function getBioPageThemes(): Promise<BioPageTheme[]> {
  const res = await apiRequest<BioPageTheme[]>(`/workspaces/${getWorkspaceId()}/bio-themes`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch themes")
  }
  return res.data
}

export async function getBioPageTheme(themeId: string): Promise<BioPageTheme> {
  const res = await apiRequest<BioPageTheme>(`/workspaces/${getWorkspaceId()}/bio-themes/${themeId}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch theme")
  }
  return res.data
}

// Public (no auth)

export async function getPublicBioPage(slug: string): Promise<PublicBioPage> {
  const response = await fetch(`/b/${slug}`)
  const data = await response.json()
  if (!data.success || !data.data) {
    throw new Error(data.error?.message || "Bio page not found")
  }
  return data.data
}

export async function trackBioLinkClick(slug: string, linkId: string): Promise<void> {
  try {
    await fetch(`/b/${slug}/click/${linkId}`, { method: "POST" })
  } catch {
    // Best-effort tracking
  }
}
