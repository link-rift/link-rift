import { describe, it, expect, beforeAll, beforeEach, afterAll, afterEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useLinks, useLink, useLinkStats } from "../useLinks"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { server } from "@/test/mocks/server"
import { mockLink, mockLinks, mockStats, MOCK_WORKSPACE_ID } from "@/test/mocks/handlers"

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
beforeEach(() => {
  // Links hooks require a current workspace to be set (enabled: !!wsId)
  useWorkspaceStore.setState({
    currentWorkspace: {
      id: MOCK_WORKSPACE_ID,
      name: "Test Workspace",
      slug: "test",
      owner_id: "user-1",
      plan: "free",
      settings: null,
      current_user_role: "owner",
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    },
    workspaces: [],
    isLoading: false,
  })
})
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe("useLinks", () => {
  it("fetches links list", async () => {
    const { result } = renderHook(() => useLinks(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.links).toHaveLength(mockLinks.length)
    expect(result.current.data?.links[0].short_code).toBe(mockLinks[0].short_code)
    expect(result.current.data?.total).toBe(mockLinks.length)
  })

  it("passes query params", async () => {
    const { result } = renderHook(() => useLinks({ search: "example", limit: 10 }), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.links).toBeDefined()
  })
})

describe("useLink", () => {
  it("fetches a single link by id", async () => {
    const { result } = renderHook(() => useLink(mockLink.id), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.id).toBe(mockLink.id)
    expect(result.current.data?.short_code).toBe(mockLink.short_code)
  })

  it("does not fetch when id is empty", () => {
    const { result } = renderHook(() => useLink(""), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useLinkStats", () => {
  it("fetches link stats", async () => {
    const { result } = renderHook(() => useLinkStats(mockLink.id), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data?.total_clicks).toBe(mockStats.total_clicks)
    expect(result.current.data?.clicks_24h).toBe(mockStats.clicks_24h)
  })

  it("does not fetch when id is empty", () => {
    const { result } = renderHook(() => useLinkStats(""), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})
