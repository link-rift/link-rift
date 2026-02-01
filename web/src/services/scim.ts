import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  SCIMToken,
  CreateSCIMTokenInput,
  CreateSCIMTokenResponse,
  SCIMSyncLog,
} from "@/types/scim"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/scim`
}

export async function createSCIMToken(
  input: CreateSCIMTokenInput
): Promise<CreateSCIMTokenResponse> {
  const res = await apiRequest<CreateSCIMTokenResponse>(`${wsBase()}/tokens`, {
    method: "POST",
    body: JSON.stringify(input),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create SCIM token")
  }
  return res.data
}

export async function listSCIMTokens(): Promise<SCIMToken[]> {
  const res = await apiRequest<SCIMToken[]>(`${wsBase()}/tokens`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to list SCIM tokens")
  }
  return res.data
}

export async function revokeSCIMToken(tokenId: string): Promise<void> {
  const res = await apiRequest(`${wsBase()}/tokens/${tokenId}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to revoke SCIM token")
  }
}

export async function listSCIMSyncLogs(
  limit = 50,
  offset = 0
): Promise<{ data: SCIMSyncLog[]; total: number }> {
  const res = await apiRequest<SCIMSyncLog[]>(
    `${wsBase()}/sync-logs?limit=${limit}&offset=${offset}`
  )
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to list SCIM sync logs")
  }
  return { data: res.data, total: res.meta?.total ?? res.data.length }
}
