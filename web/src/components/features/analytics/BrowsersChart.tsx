import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { FeatureGate } from "@/components/ui/FeatureGate"
import type { BrowserStats } from "@/types/analytics"

interface BrowsersChartProps {
  data: BrowserStats[] | undefined
  isLoading: boolean
}

export default function BrowsersChart({ data, isLoading }: BrowsersChartProps) {
  return (
    <FeatureGate feature="advanced_analytics" upgradeVariant="card">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Browsers</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-6 animate-pulse rounded bg-muted" />
              ))}
            </div>
          ) : !data?.length ? (
            <p className="text-sm text-muted-foreground">No browser data yet</p>
          ) : (
            <div className="space-y-3">
              {data.map((item) => (
                <div key={item.browser} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{item.browser}</span>
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
