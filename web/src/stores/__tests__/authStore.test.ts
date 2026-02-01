import { describe, it, expect, beforeEach, vi } from "vitest"
import { useAuthStore } from "../authStore"

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

describe("authStore", () => {
  beforeEach(() => {
    localStorageMock.clear()
    useAuthStore.setState({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: true,
    })
  })

  it("starts unauthenticated", () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.isAuthenticated).toBe(false)
  })

  describe("setAuth", () => {
    it("sets user, tokens, and isAuthenticated", () => {
      const user = { id: "1", email: "test@example.com", name: "Test" }
      useAuthStore.getState().setAuth(user as any, "access-tok", "refresh-tok")

      const state = useAuthStore.getState()
      expect(state.user).toEqual(user)
      expect(state.accessToken).toBe("access-tok")
      expect(state.refreshToken).toBe("refresh-tok")
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it("saves tokens to localStorage", () => {
      const user = { id: "1", email: "test@example.com", name: "Test" }
      useAuthStore.getState().setAuth(user as any, "at", "rt")

      expect(localStorageMock.setItem).toHaveBeenCalledWith("access_token", "at")
      expect(localStorageMock.setItem).toHaveBeenCalledWith("refresh_token", "rt")
    })
  })

  describe("clearAuth", () => {
    it("clears user and tokens", () => {
      const user = { id: "1", email: "test@example.com", name: "Test" }
      useAuthStore.getState().setAuth(user as any, "at", "rt")
      useAuthStore.getState().clearAuth()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.accessToken).toBeNull()
      expect(state.refreshToken).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.isLoading).toBe(false)
    })

    it("removes tokens from localStorage", () => {
      useAuthStore.getState().clearAuth()
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("access_token")
      expect(localStorageMock.removeItem).toHaveBeenCalledWith("refresh_token")
    })
  })

  describe("setUser", () => {
    it("updates user without changing tokens", () => {
      const user = { id: "1", email: "test@example.com", name: "Test" }
      useAuthStore.getState().setAuth(user as any, "at", "rt")

      const updatedUser = { id: "1", email: "updated@example.com", name: "Updated" }
      useAuthStore.getState().setUser(updatedUser as any)

      const state = useAuthStore.getState()
      expect(state.user?.email).toBe("updated@example.com")
      expect(state.accessToken).toBe("at")
      expect(state.isLoading).toBe(false)
    })
  })

  describe("setLoading", () => {
    it("toggles loading state", () => {
      useAuthStore.getState().setLoading(false)
      expect(useAuthStore.getState().isLoading).toBe(false)

      useAuthStore.getState().setLoading(true)
      expect(useAuthStore.getState().isLoading).toBe(true)
    })
  })
})
