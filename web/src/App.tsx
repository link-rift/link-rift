import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import AuthLayout from "@/components/layouts/AuthLayout"
import AppLayout from "@/components/layouts/AppLayout"
import ProtectedRoute from "@/components/features/auth/ProtectedRoute"
import LoginPage from "@/pages/auth/LoginPage"
import RegisterPage from "@/pages/auth/RegisterPage"
import ForgotPasswordPage from "@/pages/auth/ForgotPasswordPage"
import ResetPasswordPage from "@/pages/auth/ResetPasswordPage"
import DashboardPage from "@/pages/dashboard/DashboardPage"
import LinksPage from "@/pages/dashboard/LinksPage"
import TeamMembersPage from "@/pages/dashboard/TeamMembersPage"
import WorkspaceSettingsPage from "@/pages/dashboard/WorkspaceSettingsPage"
import AnalyticsPage from "@/pages/dashboard/AnalyticsPage"
import CustomDomainsPage from "@/pages/dashboard/CustomDomainsPage"
import BioPagesPage from "@/pages/dashboard/BioPagesPage"
import BioPageEditorPage from "@/pages/dashboard/BioPageEditorPage"
import APIKeysPage from "@/pages/dashboard/APIKeysPage"
import WebhooksPage from "@/pages/dashboard/WebhooksPage"
import PublicBioPage from "@/pages/public/PublicBioPage"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<AuthLayout />}>
            <Route path="/auth/login" element={<LoginPage />} />
            <Route path="/auth/register" element={<RegisterPage />} />
            <Route path="/auth/forgot-password" element={<ForgotPasswordPage />} />
            <Route path="/auth/reset-password" element={<ResetPasswordPage />} />
          </Route>
          <Route element={<ProtectedRoute />}>
            <Route element={<AppLayout />}>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/links" element={<LinksPage />} />
              <Route path="/analytics" element={<AnalyticsPage />} />
              <Route path="/analytics/:linkId" element={<AnalyticsPage />} />
              <Route path="/domains" element={<CustomDomainsPage />} />
              <Route path="/bio-pages" element={<BioPagesPage />} />
              <Route path="/bio-pages/:id" element={<BioPageEditorPage />} />
              <Route path="/api-keys" element={<APIKeysPage />} />
              <Route path="/webhooks" element={<WebhooksPage />} />
              <Route path="/team" element={<TeamMembersPage />} />
              <Route path="/settings" element={<WorkspaceSettingsPage />} />
            </Route>
          </Route>
          <Route path="/b/:slug" element={<PublicBioPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
