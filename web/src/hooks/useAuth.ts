import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "@/stores/authStore"
import * as authService from "@/services/auth"
import type { LoginRequest, RegisterRequest } from "@/types/auth"

export function useCurrentUser() {
  const { isAuthenticated, setUser, setLoading, clearAuth } = useAuthStore()

  return useQuery({
    queryKey: ["currentUser"],
    queryFn: async () => {
      try {
        const user = await authService.getMe()
        setUser(user)
        return user
      } catch {
        clearAuth()
        return null
      }
    },
    enabled: isAuthenticated,
    retry: false,
    staleTime: 5 * 60 * 1000,
    meta: {
      onSettled: () => setLoading(false),
    },
  })
}

export function useLogin() {
  const { setAuth } = useAuthStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (response) => {
      setAuth(response.user, response.access_token, response.refresh_token)
      queryClient.setQueryData(["currentUser"], response.user)
    },
  })
}

export function useRegister() {
  const { setAuth } = useAuthStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: RegisterRequest) => authService.register(data),
    onSuccess: (response) => {
      setAuth(response.user, response.access_token, response.refresh_token)
      queryClient.setQueryData(["currentUser"], response.user)
    },
  })
}

export function useLogout() {
  const { clearAuth } = useAuthStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => authService.logout(),
    onSuccess: () => {
      clearAuth()
      queryClient.clear()
    },
    onError: () => {
      clearAuth()
      queryClient.clear()
    },
  })
}

export function useForgotPassword() {
  return useMutation({
    mutationFn: authService.forgotPassword,
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: authService.resetPassword,
  })
}
