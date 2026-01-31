import { create } from "zustand"
import type { Workspace, WorkspaceRole } from "@/types/workspace"
import { ROLE_LEVELS } from "@/types/workspace"

interface WorkspaceState {
  workspaces: Workspace[]
  currentWorkspace: Workspace | null
  isLoading: boolean
  setWorkspaces: (workspaces: Workspace[]) => void
  setCurrentWorkspace: (workspace: Workspace) => void
  clearWorkspaces: () => void
  hasRole: (minRole: WorkspaceRole) => boolean
  canEdit: () => boolean
  canManageMembers: () => boolean
  isOwner: () => boolean
}

function getSavedWorkspaceId(): string | null {
  return localStorage.getItem("current_workspace_id")
}

function saveWorkspaceId(id: string) {
  localStorage.setItem("current_workspace_id", id)
}

export const useWorkspaceStore = create<WorkspaceState>((set, get) => ({
  workspaces: [],
  currentWorkspace: null,
  isLoading: true,

  setWorkspaces: (workspaces) => {
    const savedId = getSavedWorkspaceId()
    const current = get().currentWorkspace
    // If we already have a current workspace that's still in the list, keep it
    if (current && workspaces.some((w) => w.id === current.id)) {
      set({ workspaces, isLoading: false })
      return
    }
    // Try to restore from localStorage, otherwise pick the first workspace
    const restored = savedId ? workspaces.find((w) => w.id === savedId) : null
    const selected = restored ?? workspaces[0] ?? null
    if (selected) saveWorkspaceId(selected.id)
    set({ workspaces, currentWorkspace: selected, isLoading: false })
  },

  setCurrentWorkspace: (workspace) => {
    saveWorkspaceId(workspace.id)
    set({ currentWorkspace: workspace })
  },

  clearWorkspaces: () => {
    localStorage.removeItem("current_workspace_id")
    set({ workspaces: [], currentWorkspace: null, isLoading: false })
  },

  hasRole: (minRole) => {
    const { currentWorkspace } = get()
    if (!currentWorkspace?.current_user_role) return false
    return ROLE_LEVELS[currentWorkspace.current_user_role] >= ROLE_LEVELS[minRole]
  },

  canEdit: () => get().hasRole("editor"),

  canManageMembers: () => get().hasRole("admin"),

  isOwner: () => get().hasRole("owner"),
}))
