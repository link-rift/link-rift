import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { FeatureGate } from "@/components/ui/FeatureGate"
import type { ReferrerStats } from "@/types/analytics"

interface ReferrersTableProps {
  data: ReferrerStats[] | undefined
  isLoading: boolean
}

export default function ReferrersTable({ data, isLoading }: ReferrersTableProps) {
  return (
    <FeatureGate feature="advanced_analytics" upgradeVariant="card">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Top Referrers</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-6 animate-pulse rounded bg-muted" />
              ))}
            </div>
          ) : !data?.length ? (
            <p className="text-sm text-muted-foreground">No referrer data yet</p>
          ) : (
            <div className="space-y-3">
              {data.map((item) => (
                <div key={item.referrer} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="truncate font-medium">{item.referrer}</span>
                    <span className="text-muted-foreground">
                      {item.clicks} ({item.percent.toFixed(1)}%)
                    </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-primary"
                      style={{ width: `${item.percent}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </FeatureGate>
  )
}
