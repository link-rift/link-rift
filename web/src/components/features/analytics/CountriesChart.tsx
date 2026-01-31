import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { FeatureGate } from "@/components/ui/FeatureGate"
import type { CountryStats } from "@/types/analytics"

interface CountriesChartProps {
  data: CountryStats[] | undefined
  isLoading: boolean
}

export default function CountriesChart({ data, isLoading }: CountriesChartProps) {
  return (
    <FeatureGate feature="advanced_analytics" upgradeVariant="card">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Top Countries</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-6 animate-pulse rounded bg-muted" />
              ))}
            </div>
          ) : !data?.length ? (
            <p className="text-sm text-muted-foreground">No country data yet</p>
          ) : (
            <div className="space-y-3">
              {data.map((item) => (
                <div key={item.country_code} className="space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{item.country_code}</span>
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
