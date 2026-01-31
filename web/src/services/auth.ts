import { apiRequest, setTokens, clearTokens } from "./api"
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  VerifyEmailRequest,
  User,
} from "@/types/auth"

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Login failed")
  }
  setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await apiRequest<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Registration failed")
  }
  setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

export async function logout(): Promise<void> {
  await apiRequest("/auth/logout", { method: "POST" })
  clearTokens()
}

export async function getMe(): Promise<User> {
  const res = await apiRequest<User>("/auth/me")
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to get user")
  }
  return res.data
}

export async function forgotPassword(data: ForgotPasswordRequest): Promise<void> {
  const res = await apiRequest<{ message: string }>("/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to send reset email")
  }
}

export async function resetPassword(data: ResetPasswordRequest): Promise<void> {
  const res = await apiRequest<{ message: string }>("/auth/reset-password", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to reset password")
  }
}

export async function verifyEmail(data: VerifyEmailRequest): Promise<void> {
  const res = await apiRequest<{ message: string }>("/auth/verify-email", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to verify email")
  }
}
