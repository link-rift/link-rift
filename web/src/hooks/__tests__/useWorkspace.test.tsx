import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import {
  useWorkspaces,
  useWorkspace,
  useCreateWorkspace,
  useWorkspaceMembers,
} from "../useWorkspace"
import { useAuthStore } from "@/stores/authStore"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { Workspace, WorkspaceMember } from "@/types/workspace"

const mockWorkspace: Workspace = {
  id: "ws-1",
  name: "Test Workspace",
  slug: "test",
  owner_id: "user-1",
  plan: "free",
  settings: null,
  current_user_role: "owner",
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
}

const mockWorkspace2: Workspace = {
  ...mockWorkspace,
  id: "ws-2",
  name: "Second Workspace",
  slug: "second",
}

const mockMember: WorkspaceMember = {
  id: "m-1",
  workspace_id: "ws-1",
  user_id: "user-1",
  role: "owner",
  email: "owner@example.com",
  name: "Owner User",
  joined_at: "2025-01-01T00:00:00Z",
}

const handlers = [
  http.get("/api/v1/workspaces", () => {
    return HttpResponse.json({
      success: true,
      data: [mockWorkspace, mockWorkspace2],
    })
  }),

  http.get("/api/v1/workspaces/:id", ({ params }) => {
    if (params.id === mockWorkspace.id) {
      return HttpResponse.json({ success: true, data: mockWorkspace })
    }
    return HttpResponse.json(
      { success: false, error: { code: "NOT_FOUND", message: "Workspace not found" } },
      { status: 404 },
    )
  }),

  http.post("/api/v1/workspaces", async ({ request }) => {
    const body = (await request.json()) as Record<string, string>
    return HttpResponse.json(
      {
        success: true,
        data: { ...mockWorkspace, id: "ws-new", name: body.name, slug: body.slug || "new" },
      },
      { status: 201 },
    )
  }),

  http.get("/api/v1/workspaces/:id/members", ({ params }) => {
    if (params.id === mockWorkspace.id) {
      return HttpResponse.json({ success: true, data: [mockMember] })
    }
    return HttpResponse.json(
      { success: false, error: { code: "NOT_FOUND", message: "Workspace not found" } },
      { status: 404 },
    )
  }),
]

const server = setupServer(...handlers)

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => {
  server.resetHandlers()
  useAuthStore.setState({
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
    isLoading: false,
  })
  useWorkspaceStore.setState({
    workspaces: [],
    currentWorkspace: null,
    isLoading: true,
  })
})
afterAll(() => server.close())

describe("useWorkspaces", () => {
  it("fetches workspaces when authenticated", async () => {
    useAuthStore.setState({ isAuthenticated: true })

    const { result } = renderHook(() => useWorkspaces(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(2)
    expect(result.current.data?.[0].id).toBe(mockWorkspace.id)
  })

  it("updates workspace store on success", async () => {
    useAuthStore.setState({ isAuthenticated: true })

    const { result } = renderHook(() => useWorkspaces(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    const storeState = useWorkspaceStore.getState()
    expect(storeState.workspaces).toHaveLength(2)
    expect(storeState.currentWorkspace?.id).toBe(mockWorkspace.id)
  })

  it("does not fetch when not authenticated", () => {
    useAuthStore.setState({ isAuthenticated: false })

    const { result } = renderHook(() => useWorkspaces(), { wrapper: createWrapper() })

    expect(result.current.fetchStatus).toBe("idle")
  })

  it("clears workspaces on fetch error", async () => {
    useAuthStore.setState({ isAuthenticated: true })
    useWorkspaceStore.getState().setWorkspaces([mockWorkspace])

    server.use(
      http.get("/api/v1/workspaces", () => {
        return HttpResponse.json(
          { success: false, error: { code: "ERROR", message: "Server error" } },
          { status: 500 },
        )
      }),
    )

    const { result } = renderHook(() => useWorkspaces(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    // The hook catches the error and returns empty array
    expect(result.current.data).toEqual([])
    expect(useWorkspaceStore.getState().workspaces).toEqual([])
  })
})

describe("useWorkspace", () => {
  it("fetches a single workspace by id", async () => {
    const { result } = renderHook(() => useWorkspace(mockWorkspace.id), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.id).toBe(mockWorkspace.id)
    expect(result.current.data?.name).toBe(mockWorkspace.name)
  })

  it("does not fetch when id is empty", () => {
    const { result } = renderHook(() => useWorkspace(""), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useCreateWorkspace", () => {
  it("creates a workspace", async () => {
    const { result } = renderHook(() => useCreateWorkspace(), { wrapper: createWrapper() })

    let data: Workspace | undefined
    await act(async () => {
      data = await result.current.mutateAsync({ name: "New Workspace" })
    })

    expect(data?.id).toBe("ws-new")
    expect(data?.name).toBe("New Workspace")
  })
})

describe("useWorkspaceMembers", () => {
  it("fetches members for a workspace", async () => {
    const { result } = renderHook(() => useWorkspaceMembers(mockWorkspace.id), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toHaveLength(1)
    expect(result.current.data?.[0].role).toBe("owner")
    expect(result.current.data?.[0].email).toBe(mockMember.email)
  })

  it("does not fetch when workspaceId is empty", () => {
    const { result } = renderHook(() => useWorkspaceMembers(""), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})
