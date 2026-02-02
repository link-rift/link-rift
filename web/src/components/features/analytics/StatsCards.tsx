import { MousePointerClick, Users, TrendingUp, Calendar } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { LinkAnalytics, WorkspaceAnalytics } from "@/types/analytics"

interface StatsCardsProps {
  stats: LinkAnalytics | WorkspaceAnalytics | undefined
  isLoading: boolean
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return String(n)
}

const cardConfig = [
  {
    title: "Total Clicks",
    key: "total_clicks" as const,
    icon: MousePointerClick,
    gradient: "from-primary/10 to-transparent",
  },
  {
    title: "Unique Clicks",
    key: "unique_clicks" as const,
    icon: Users,
    gradient: "from-chart-2/10 to-transparent",
  },
  {
    title: "Last 24h",
    key: "clicks_24h" as const,
    icon: TrendingUp,
    gradient: "from-chart-3/10 to-transparent",
  },
  {
    title: "Last 7d",
    key: "clicks_7d" as const,
    icon: Calendar,
    gradient: "from-chart-4/10 to-transparent",
  },
]

export default function StatsCards({ stats, isLoading }: StatsCardsProps) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {cardConfig.map((card) => {
        const Icon = card.icon
        return (
          <Card
            key={card.title}
            className={`bg-gradient-to-br ${card.gradient}`}
          >
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {card.title}
              </CardTitle>
              <Icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="h-9 w-20 animate-pulse rounded bg-muted" />
              ) : (
                <p className="text-3xl font-bold tracking-tight">
                  {formatNumber(stats?.[card.key] ?? 0)}
                </p>
              )}
            </CardContent>
          </Card>
        )
      })}
    </div>
  )
}
