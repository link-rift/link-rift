import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  SSOConfig,
  CreateSSOConfigInput,
  UpdateSSOConfigInput,
  SSOIdentity,
} from "@/types/sso"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/sso`
}

export async function getSSOConfig(): Promise<SSOConfig> {
  const res = await apiRequest<SSOConfig>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch SSO config")
  }
  return res.data
}

export async function createSSOConfig(
  input: CreateSSOConfigInput
): Promise<SSOConfig> {
  const res = await apiRequest<SSOConfig>(wsBase(), {
    method: "POST",
    body: JSON.stringify(input),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create SSO config")
  }
  return res.data
}

export async function updateSSOConfig(
  input: UpdateSSOConfigInput
): Promise<SSOConfig> {
  const res = await apiRequest<SSOConfig>(wsBase(), {
    method: "PUT",
    body: JSON.stringify(input),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update SSO config")
  }
  return res.data
}

export async function deleteSSOConfig(): Promise<void> {
  const res = await apiRequest(wsBase(), { method: "DELETE" })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete SSO config")
  }
}

export async function listSSOIdentities(): Promise<SSOIdentity[]> {
  const res = await apiRequest<SSOIdentity[]>(`${wsBase()}/identities`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch SSO identities")
  }
  return res.data
}
