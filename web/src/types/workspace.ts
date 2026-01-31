export type WorkspaceRole = "owner" | "admin" | "editor" | "viewer"

export const ROLE_LEVELS: Record<WorkspaceRole, number> = {
  viewer: 1,
  editor: 2,
  admin: 3,
  owner: 4,
}

export const ROLE_LABELS: Record<WorkspaceRole, string> = {
  viewer: "Viewer",
  editor: "Editor",
  admin: "Admin",
  owner: "Owner",
}

export interface Workspace {
  id: string
  name: string
  slug: string
  owner_id: string
  plan: string
  settings: Record<string, unknown> | null
  member_count?: number
  current_user_role?: WorkspaceRole
  created_at: string
  updated_at: string
}

export interface WorkspaceMember {
  id: string
  workspace_id: string
  user_id: string
  role: WorkspaceRole
  email: string
  name: string
  avatar_url?: string | null
  joined_at: string
}

export interface CreateWorkspaceRequest {
  name: string
  slug?: string
}

export interface UpdateWorkspaceRequest {
  name?: string
  slug?: string
}

export interface InviteMemberRequest {
  email: string
  role: WorkspaceRole
}

export interface UpdateMemberRoleRequest {
  role: WorkspaceRole
}

export interface TransferOwnershipRequest {
  new_owner_id: string
}
