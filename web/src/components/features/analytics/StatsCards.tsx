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

export default function StatsCards({ stats, isLoading }: StatsCardsProps) {
  const cards = [
    {
      title: "Total Clicks",
      value: stats?.total_clicks ?? 0,
    },
    {
      title: "Unique Clicks",
      value: stats?.unique_clicks ?? 0,
    },
    {
      title: "Last 24h",
      value: stats?.clicks_24h ?? 0,
    },
    {
      title: "Last 7d",
      value: stats?.clicks_7d ?? 0,
    },
  ]

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {card.title}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="h-8 w-20 animate-pulse rounded bg-muted" />
            ) : (
              <p className="text-2xl font-bold">{formatNumber(card.value)}</p>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
