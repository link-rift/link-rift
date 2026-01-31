import { apiRequest } from "./api"
import type {
  Workspace,
  WorkspaceMember,
  CreateWorkspaceRequest,
  UpdateWorkspaceRequest,
  InviteMemberRequest,
  UpdateMemberRoleRequest,
  TransferOwnershipRequest,
} from "@/types/workspace"

export async function createWorkspace(data: CreateWorkspaceRequest): Promise<Workspace> {
  const res = await apiRequest<Workspace>("/workspaces", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create workspace")
  }
  return res.data
}

export async function getWorkspaces(): Promise<Workspace[]> {
  const res = await apiRequest<Workspace[]>("/workspaces")
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch workspaces")
  }
  return res.data
}

export async function getWorkspace(id: string): Promise<Workspace> {
  const res = await apiRequest<Workspace>(`/workspaces/${id}`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch workspace")
  }
  return res.data
}

export async function updateWorkspace(id: string, data: UpdateWorkspaceRequest): Promise<Workspace> {
  const res = await apiRequest<Workspace>(`/workspaces/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update workspace")
  }
  return res.data
}

export async function deleteWorkspace(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`/workspaces/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete workspace")
  }
}

export async function getMembers(workspaceId: string): Promise<WorkspaceMember[]> {
  const res = await apiRequest<WorkspaceMember[]>(`/workspaces/${workspaceId}/members`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch members")
  }
  return res.data
}

export async function inviteMember(workspaceId: string, data: InviteMemberRequest): Promise<WorkspaceMember> {
  const res = await apiRequest<WorkspaceMember>(`/workspaces/${workspaceId}/members`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to invite member")
  }
  return res.data
}

export async function updateMemberRole(workspaceId: string, userId: string, data: UpdateMemberRoleRequest): Promise<WorkspaceMember> {
  const res = await apiRequest<WorkspaceMember>(`/workspaces/${workspaceId}/members/${userId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to update member role")
  }
  return res.data
}

export async function removeMember(workspaceId: string, userId: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`/workspaces/${workspaceId}/members/${userId}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to remove member")
  }
}

export async function transferOwnership(workspaceId: string, data: TransferOwnershipRequest): Promise<void> {
  const res = await apiRequest<{ message: string }>(`/workspaces/${workspaceId}/transfer`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to transfer ownership")
  }
}
