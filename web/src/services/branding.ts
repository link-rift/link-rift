import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { WorkspaceBranding, UpdateBrandingInput } from "@/types/branding"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/branding`
}

export async function getBranding(): Promise<WorkspaceBranding> {
  const res = await apiRequest<WorkspaceBranding>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch branding")
  }
  return res.data
}

export async function updateBranding(
  input: UpdateBrandingInput
): Promise<WorkspaceBranding> {
  const res = await apiRequest<WorkspaceBranding>(wsBase(), {
    method: "PUT",
    body: JSON.stringify(input),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update branding")
  }
  return res.data
}

export async function deleteBranding(): Promise<void> {
  const res = await apiRequest(wsBase(), { method: "DELETE" })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete branding")
  }
}
