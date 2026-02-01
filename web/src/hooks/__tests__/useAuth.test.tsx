import { describe, it, expect, beforeAll, afterAll, afterEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { useCurrentUser, useLogin, useRegister, useLogout } from "../useAuth"
import { useAuthStore } from "@/stores/authStore"
import type { User } from "@/types/auth"

const mockUser: User = {
  id: "user-1",
  email: "test@example.com",
  name: "Test User",
  two_factor_enabled: false,
  created_at: "2025-01-01T00:00:00Z",
  updated_at: "2025-01-01T00:00:00Z",
}

const handlers = [
  http.get("/api/v1/auth/me", () => {
    return HttpResponse.json({ success: true, data: mockUser })
  }),

  http.post("/api/v1/auth/login", async ({ request }) => {
    const body = (await request.json()) as Record<string, string>
    if (body.email === "bad@example.com") {
      return HttpResponse.json(
        { success: false, error: { code: "INVALID_CREDENTIALS", message: "Invalid credentials" } },
        { status: 401 },
      )
    }
    return HttpResponse.json({
      success: true,
      data: {
        access_token: "test-access-token",
        refresh_token: "test-refresh-token",
        user: mockUser,
      },
    })
  }),

  http.post("/api/v1/auth/register", () => {
    return HttpResponse.json({
      success: true,
      data: {
        access_token: "reg-access-token",
        refresh_token: "reg-refresh-token",
        user: { ...mockUser, id: "user-new" },
      },
    })
  }),

  http.post("/api/v1/auth/logout", () => {
    return HttpResponse.json({ success: true, data: { message: "Logged out" } })
  }),

  // Handle refresh so apiRequest doesn't attempt navigation on 401
  http.post("/api/v1/auth/refresh", () => {
    return HttpResponse.json({
      success: true,
      data: { access_token: "refreshed-token", refresh_token: "refreshed-rt" },
    })
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
  localStorage.clear()
  useAuthStore.setState({
    user: null,
    accessToken: null,
    refreshToken: null,
    isAuthenticated: false,
    isLoading: true,
  })
})
afterAll(() => server.close())

describe("useCurrentUser", () => {
  it("fetches user when authenticated", async () => {
    useAuthStore.setState({ isAuthenticated: true })

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.email).toBe(mockUser.email)
  })

  it("does not fetch when not authenticated", () => {
    useAuthStore.setState({ isAuthenticated: false })

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })

  it("clears auth on fetch error", async () => {
    useAuthStore.setState({ isAuthenticated: true })

    server.use(
      http.get("/api/v1/auth/me", () => {
        return HttpResponse.json(
          { success: false, error: { code: "UNAUTHORIZED", message: "Invalid token" } },
          { status: 401 },
        )
      }),
    )

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toBeNull()
  })
})

describe("useRegister", () => {
  it("sets auth store on successful registration", async () => {
    const { result } = renderHook(() => useRegister(), { wrapper: createWrapper() })

    await act(async () => {
      await result.current.mutateAsync({
        email: "new@example.com",
        password: "password123",
        name: "New User",
      })
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.accessToken).toBe("reg-access-token")
  })
})

describe("useLogout", () => {
  it("clears auth store on logout", async () => {
    useAuthStore.getState().setAuth(mockUser, "at", "rt")
    expect(useAuthStore.getState().isAuthenticated).toBe(true)

    const { result } = renderHook(() => useLogout(), { wrapper: createWrapper() })

    await act(async () => {
      await result.current.mutateAsync()
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
  })
})

describe("useLogin", () => {
  it("sets auth store on successful login", async () => {
    const { result } = renderHook(() => useLogin(), { wrapper: createWrapper() })

    await act(async () => {
      await result.current.mutateAsync({ email: "test@example.com", password: "password" })
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.email).toBe(mockUser.email)
    expect(state.accessToken).toBe("test-access-token")
    expect(state.refreshToken).toBe("test-refresh-token")
  })

  it("handles invalid credentials error", async () => {
    const { result } = renderHook(() => useLogin(), { wrapper: createWrapper() })

    // Use mutate + waitFor instead of mutateAsync to avoid unhandled rejection leak
    act(() => {
      result.current.mutate({ email: "bad@example.com", password: "wrong" })
    })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error).toBeTruthy()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })
})
