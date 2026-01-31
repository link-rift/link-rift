import type { ApiResponse } from "@/types/auth"

const API_BASE = "/api/v1"

let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null

function getAccessToken(): string | null {
  return localStorage.getItem("access_token")
}

function getRefreshToken(): string | null {
  return localStorage.getItem("refresh_token")
}

function setTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem("access_token", accessToken)
  localStorage.setItem("refresh_token", refreshToken)
}

function clearTokens() {
  localStorage.removeItem("access_token")
  localStorage.removeItem("refresh_token")
}

async function attemptRefresh(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return null

  try {
    const response = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      clearTokens()
      return null
    }

    const data = await response.json()
    if (data.success && data.data) {
      setTokens(data.data.access_token, data.data.refresh_token)
      return data.data.access_token
    }

    clearTokens()
    return null
  } catch {
    clearTokens()
    return null
  }
}

export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const url = `${API_BASE}${endpoint}`
  const accessToken = getAccessToken()

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  }

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`
  }

  let response = await fetch(url, { ...options, headers })

  if (response.status === 401 && accessToken) {
    if (!isRefreshing) {
      isRefreshing = true
      refreshPromise = attemptRefresh()
    }

    const newToken = await refreshPromise
    isRefreshing = false
    refreshPromise = null

    if (newToken) {
      headers["Authorization"] = `Bearer ${newToken}`
      response = await fetch(url, { ...options, headers })
    } else {
      clearTokens()
      window.location.href = "/auth/login"
      return { success: false, error: { code: "UNAUTHORIZED", message: "Session expired" } }
    }
  }

  const data: ApiResponse<T> = await response.json()
  return data
}

export { setTokens, clearTokens, getAccessToken, getRefreshToken }
