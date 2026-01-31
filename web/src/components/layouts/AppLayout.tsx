import { Outlet, Link, useLocation } from "react-router-dom"
import { useLogout } from "@/hooks/useAuth"
import { useWorkspaces } from "@/hooks/useWorkspace"
import { useAuthStore } from "@/stores/authStore"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { Button } from "@/components/ui/button"
import WorkspaceSwitcher from "@/components/features/workspaces/WorkspaceSwitcher"

export default function AppLayout() {
  const { user } = useAuthStore()
  const { canManageMembers } = useWorkspaceStore()
  const logout = useLogout()
  const location = useLocation()

  // Fetch workspaces on mount (syncs to store)
  useWorkspaces()

  const navItems = [
    { label: "Dashboard", href: "/" },
    { label: "Links", href: "/links" },
    ...(canManageMembers()
      ? [
          { label: "Team", href: "/team" },
          { label: "Settings", href: "/settings" },
        ]
      : []),
  ]

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b bg-background">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-6">
            <h1 className="text-lg font-semibold">Linkrift</h1>
            <WorkspaceSwitcher />
            <nav className="flex items-center gap-1">
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  to={item.href}
                  className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                    location.pathname === item.href
                      ? "bg-muted text-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-4">
            {user && (
              <span className="text-sm text-muted-foreground">{user.email}</span>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
            >
              {logout.isPending ? "Logging out..." : "Logout"}
            </Button>
          </div>
        </div>
      </header>
      <main className="flex-1">
        <div className="mx-auto max-w-7xl px-4 py-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
