import { Outlet, Navigate } from "react-router-dom"
import { useAuthStore } from "@/stores/authStore"

export default function AuthLayout() {
  const { isAuthenticated, isLoading } = useAuthStore()

  if (isLoading) {
    return null
  }

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 px-4">
      <div className="w-full max-w-md">
        <Outlet />
      </div>
    </div>
  )
}
