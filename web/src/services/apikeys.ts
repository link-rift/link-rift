import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { APIKey, CreateAPIKeyRequest, CreateAPIKeyResponse } from "@/types/apikey"

function wsBase(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return `/workspaces/${ws.id}/api-keys`
}

export async function createAPIKey(data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> {
  const res = await apiRequest<CreateAPIKeyResponse>(wsBase(), {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create API key")
  }
  return res.data
}

export async function getAPIKeys(): Promise<APIKey[]> {
  const res = await apiRequest<APIKey[]>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch API keys")
  }
  return res.data
}

export async function revokeAPIKey(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to revoke API key")
  }
}
