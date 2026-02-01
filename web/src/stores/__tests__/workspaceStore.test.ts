import { describe, it, expect, beforeEach, vi } from "vitest"
import { useWorkspaceStore } from "../workspaceStore"
import type { Workspace } from "@/types/workspace"

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
  }
})()

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock })

function makeWorkspace(overrides: Partial<Workspace> = {}): Workspace {
  return {
    id: "ws-1",
    name: "Test Workspace",
    slug: "test",
    owner_id: "user-1",
    plan: "free",
    settings: null,
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
    current_user_role: "owner",
    ...overrides,
  }
}

describe("workspaceStore", () => {
  beforeEach(() => {
    localStorageMock.clear()
    useWorkspaceStore.setState({
      workspaces: [],
      currentWorkspace: null,
      isLoading: true,
    })
  })

  it("starts with no workspaces", () => {
    const state = useWorkspaceStore.getState()
    expect(state.workspaces).toEqual([])
    expect(state.currentWorkspace).toBeNull()
  })

  describe("setWorkspaces", () => {
    it("sets workspaces and selects first if none current", () => {
      const ws = makeWorkspace()
      useWorkspaceStore.getState().setWorkspaces([ws])

      const state = useWorkspaceStore.getState()
      expect(state.workspaces).toHaveLength(1)
      expect(state.currentWorkspace?.id).toBe("ws-1")
      expect(state.isLoading).toBe(false)
    })

    it("keeps current workspace if still in list", () => {
      const ws1 = makeWorkspace({ id: "ws-1" })
      const ws2 = makeWorkspace({ id: "ws-2", name: "Second" })

      useWorkspaceStore.getState().setWorkspaces([ws1, ws2])
      useWorkspaceStore.getState().setCurrentWorkspace(ws2)

      // Re-set workspaces — should keep ws2
      useWorkspaceStore.getState().setWorkspaces([ws1, ws2])
      expect(useWorkspaceStore.getState().currentWorkspace?.id).toBe("ws-2")
    })

    it("sets currentWorkspace to null when list is empty", () => {
      useWorkspaceStore.getState().setWorkspaces([])
      expect(useWorkspaceStore.getState().currentWorkspace).toBeNull()
    })
  })

  describe("setCurrentWorkspace", () => {
    it("updates current workspace", () => {
      const ws = makeWorkspace({ id: "ws-new" })
      useWorkspaceStore.getState().setCurrentWorkspace(ws)

      expect(useWorkspaceStore.getState().currentWorkspace?.id).toBe("ws-new")
    })

    it("saves workspace id to localStorage", () => {
      const ws = makeWorkspace({ id: "ws-saved" })
      useWorkspaceStore.getState().setCurrentWorkspace(ws)

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        "current_workspace_id",
        "ws-saved"
      )
    })
  })

  describe("clearWorkspaces", () => {
    it("resets all workspace state", () => {
      useWorkspaceStore.getState().setWorkspaces([makeWorkspace()])
      useWorkspaceStore.getState().clearWorkspaces()

      const state = useWorkspaceStore.getState()
      expect(state.workspaces).toEqual([])
      expect(state.currentWorkspace).toBeNull()
      expect(state.isLoading).toBe(false)
    })
  })

  describe("role helpers", () => {
    it("hasRole returns false when no workspace", () => {
      expect(useWorkspaceStore.getState().hasRole("viewer")).toBe(false)
    })

    it("hasRole returns true for owner checking editor", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "owner" }),
      ])
      expect(useWorkspaceStore.getState().hasRole("editor")).toBe(true)
    })

    it("hasRole returns false for viewer checking admin", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "viewer" }),
      ])
      expect(useWorkspaceStore.getState().hasRole("admin")).toBe(false)
    })

    it("canEdit returns true for editor", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "editor" }),
      ])
      expect(useWorkspaceStore.getState().canEdit()).toBe(true)
    })

    it("canEdit returns false for viewer", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "viewer" }),
      ])
      expect(useWorkspaceStore.getState().canEdit()).toBe(false)
    })

    it("canManageMembers returns true for admin", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "admin" }),
      ])
      expect(useWorkspaceStore.getState().canManageMembers()).toBe(true)
    })

    it("isOwner returns true only for owner", () => {
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ current_user_role: "owner" }),
      ])
      expect(useWorkspaceStore.getState().isOwner()).toBe(true)

      // Use a different ID so setWorkspaces doesn't short-circuit by keeping old currentWorkspace
      useWorkspaceStore.getState().setWorkspaces([
        makeWorkspace({ id: "ws-admin", current_user_role: "admin" }),
      ])
      expect(useWorkspaceStore.getState().isOwner()).toBe(false)
    })
  })
})
