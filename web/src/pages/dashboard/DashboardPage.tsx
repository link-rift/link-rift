import { Link } from "react-router-dom"
import { useAuthStore } from "@/stores/authStore"
import { useWorkspaceAnalytics } from "@/hooks/useAnalytics"
import { useRealtimeAnalytics } from "@/hooks/useRealtimeAnalytics"
import StatsCards from "@/components/features/analytics/StatsCards"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export default function DashboardPage() {
  const { user } = useAuthStore()
  const workspaceAnalytics = useWorkspaceAnalytics("7d")
  const { recentClicks } = useRealtimeAnalytics()

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">
          Welcome{user?.name ? `, ${user.name}` : ""}
        </h2>
        <p className="text-muted-foreground mt-1">
          Here&apos;s an overview of your link performance.
        </p>
      </div>

      <StatsCards
        stats={workspaceAnalytics.data}
        isLoading={workspaceAnalytics.isLoading}
      />

      <div className="grid gap-6 md:grid-cols-2">
        {/* Top Links */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Top Links</CardTitle>
            <Link
              to="/analytics"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              View all
            </Link>
          </CardHeader>
          <CardContent>
            {workspaceAnalytics.isLoading ? (
              <div className="space-y-3">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center justify-between">
                    <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                    <div className="h-4 w-16 animate-pulse rounded bg-muted" />
                  </div>
                ))}
              </div>
            ) : workspaceAnalytics.data?.top_links &&
              workspaceAnalytics.data.top_links.length > 0 ? (
              <div className="space-y-3">
                {workspaceAnalytics.data.top_links.map((link) => (
                  <Link
                    key={link.link_id}
                    to={`/analytics/${link.link_id}`}
                    className="flex items-center justify-between rounded-md px-2 py-1.5 transition-colors hover:bg-muted/50"
                  >
                    <span className="text-sm font-medium text-primary">
                      /{link.short_code}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      {link.total_clicks} clicks
                    </span>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                No link data yet. Create your first link to get started.
              </p>
            )}
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            {recentClicks.length > 0 ? (
              <div className="space-y-3">
                {recentClicks.slice(0, 5).map((click, i) => (
                  <div
                    key={`${click.link_id}-${click.timestamp}-${i}`}
                    className="flex items-center justify-between text-sm"
                  >
                    <span className="font-medium text-primary">
                      /{click.short_code}
                    </span>
                    <span className="text-muted-foreground">
                      {click.country_code && `${click.country_code} · `}
                      {click.device_type ?? "unknown"}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                Real-time clicks will appear here as they happen.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
