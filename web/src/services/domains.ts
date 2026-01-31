import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  Domain,
  CreateDomainRequest,
  VerificationInstructions,
} from "@/types/domain"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}/domains`
}

export async function createDomain(data: CreateDomainRequest): Promise<Domain> {
  const res = await apiRequest<Domain>(wsBase(), {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create domain")
  }
  return res.data
}

export async function getDomains(): Promise<Domain[]> {
  const res = await apiRequest<Domain[]>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch domains")
  }
  return res.data
}

export async function getDomain(id: string): Promise<Domain> {
  const res = await apiRequest<Domain>(`${wsBase()}/${id}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch domain")
  }
  return res.data
}

export async function verifyDomain(id: string): Promise<Domain> {
  const res = await apiRequest<Domain>(`${wsBase()}/${id}/verify`, {
    method: "POST",
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to verify domain")
  }
  return res.data
}

export async function deleteDomain(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete domain")
  }
}

export async function getDNSRecords(id: string): Promise<VerificationInstructions> {
  const res = await apiRequest<VerificationInstructions>(`${wsBase()}/${id}/dns-records`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch DNS records")
  }
  return res.data
}
