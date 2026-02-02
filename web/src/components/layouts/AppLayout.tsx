import { useState } from "react"
import { Outlet, Link, useLocation, useNavigate } from "react-router-dom"
import { MenuIcon, Link2, LogOut, Settings, ChevronDown } from "lucide-react"
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import WorkspaceSwitcher from "@/components/features/workspaces/WorkspaceSwitcher"
import ThemeToggle from "@/components/ui/ThemeToggle"

export default function AppLayout() {
  const { user } = useAuthStore()
  const { canManageMembers } = useWorkspaceStore()
  const { hasFeature } = useLicenseStore()
  const logout = useLogout()
  const location = useLocation()
  const navigate = useNavigate()
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

  const initials = user?.name
    ? user.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : user?.email?.[0]?.toUpperCase() ?? "?"

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-xl">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                <Link2 className="h-4 w-4 text-primary-foreground" />
              </div>
              <span className="text-lg font-bold tracking-tight">Linkrift</span>
            </Link>
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
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:text-foreground hover:bg-accent"
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="hidden items-center gap-2 md:flex">
                  <div className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-br from-primary to-primary/70 text-xs font-semibold text-primary-foreground">
                    {initials}
                  </div>
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col gap-1">
                    {user?.name && (
                      <p className="text-sm font-medium">{user.name}</p>
                    )}
                    <p className="text-xs text-muted-foreground truncate">
                      {user?.email}
                    </p>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={() => navigate("/settings")}>
                  <Settings className="mr-2 h-4 w-4" />
                  Settings
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => logout.mutate()}
                  disabled={logout.isPending}
                >
                  <LogOut className="mr-2 h-4 w-4" />
                  {logout.isPending ? "Logging out..." : "Log out"}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
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
        <div className="h-px bg-gradient-to-r from-transparent via-primary/20 to-transparent" />
      </header>

      {/* Mobile navigation sheet */}
      <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
        <SheetContent side="left" className="w-72">
          <SheetHeader>
            <SheetTitle className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary">
                <Link2 className="h-3.5 w-3.5 text-primary-foreground" />
              </div>
              Linkrift
            </SheetTitle>
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
                      ? "bg-primary/10 text-primary"
                      : "text-muted-foreground hover:text-foreground hover:bg-accent"
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
                <LogOut className="mr-2 h-4 w-4" />
                {logout.isPending ? "Logging out..." : "Log out"}
              </Button>
            </div>
          </div>
        </SheetContent>
      </Sheet>

      <main className="flex-1 bg-muted/30">
        <div className="mx-auto max-w-7xl px-4 py-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
