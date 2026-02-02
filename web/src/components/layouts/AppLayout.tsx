import { useState } from "react"
import { Outlet, Link, useLocation } from "react-router-dom"
import { MenuIcon } from "lucide-react"
import { useLogout } from "@/hooks/useAuth"
import { useWorkspaces } from "@/hooks/useWorkspace"
import { useAuthStore } from "@/stores/authStore"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { useLicenseStore } from "@/stores/licenseStore"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import WorkspaceSwitcher from "@/components/features/workspaces/WorkspaceSwitcher"

export default function AppLayout() {
  const { user } = useAuthStore()
  const { canManageMembers } = useWorkspaceStore()
  const { hasFeature } = useLicenseStore()
  const logout = useLogout()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  // Fetch workspaces on mount (syncs to store)
  useWorkspaces()

  const navItems = [
    { label: "Dashboard", href: "/" },
    { label: "Links", href: "/links" },
    { label: "Analytics", href: "/analytics" },
    { label: "Bio Pages", href: "/bio-pages" },
    ...(canManageMembers()
      ? [
          { label: "Domains", href: "/domains" },
          { label: "API Keys", href: "/api-keys" },
          { label: "Webhooks", href: "/webhooks" },
          ...(hasFeature("audit_logs")
            ? [{ label: "Audit Logs", href: "/audit-logs" }]
            : []),
          ...(hasFeature("white_label")
            ? [{ label: "Branding", href: "/branding" }]
            : []),
          ...(hasFeature("saml")
            ? [{ label: "SSO", href: "/sso-settings" }]
            : []),
          ...(hasFeature("scim")
            ? [{ label: "SCIM", href: "/scim-settings" }]
            : []),
          { label: "Team", href: "/team" },
          { label: "Settings", href: "/settings" },
        ]
      : []),
  ]

  function isActive(href: string) {
    if (href === "/") return location.pathname === "/"
    return location.pathname.startsWith(href)
  }

  return (
    <div className="flex min-h-screen flex-col">
      <header className="border-b bg-background">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-6">
            <h1 className="text-lg font-semibold">Linkrift</h1>
            <div className="hidden md:block">
              <WorkspaceSwitcher />
            </div>
            <nav className="hidden items-center gap-1 md:flex">
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  to={item.href}
                  className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                    isActive(item.href)
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
              <span className="hidden text-sm text-muted-foreground md:block">
                {user.email}
              </span>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
              className="hidden md:inline-flex"
            >
              {logout.isPending ? "Logging out..." : "Logout"}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden"
              onClick={() => setMobileMenuOpen(true)}
            >
              <MenuIcon className="h-5 w-5" />
              <span className="sr-only">Open menu</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Mobile navigation sheet */}
      <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
        <SheetContent side="left" className="w-72">
          <SheetHeader>
            <SheetTitle>Linkrift</SheetTitle>
            <SheetDescription className="sr-only">Navigation menu</SheetDescription>
          </SheetHeader>
          <div className="flex flex-col gap-4 px-4">
            <WorkspaceSwitcher />
            <nav className="flex flex-col gap-1">
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  to={item.href}
                  onClick={() => setMobileMenuOpen(false)}
                  className={`rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                    isActive(item.href)
                      ? "bg-muted text-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
            <div className="border-t pt-4">
              {user && (
                <p className="mb-3 truncate text-sm text-muted-foreground">
                  {user.email}
                </p>
              )}
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start"
                onClick={() => {
                  logout.mutate()
                  setMobileMenuOpen(false)
                }}
                disabled={logout.isPending}
              >
                {logout.isPending ? "Logging out..." : "Logout"}
              </Button>
            </div>
          </div>
        </SheetContent>
      </Sheet>

      <main className="flex-1">
        <div className="mx-auto max-w-7xl px-4 py-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
