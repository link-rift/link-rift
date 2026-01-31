import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { FeatureGate } from "@/components/ui/FeatureGate"
import type { DeviceBreakdown } from "@/types/analytics"

interface DevicesPieChartProps {
  data: DeviceBreakdown | undefined
  isLoading: boolean
}

const COLORS = [
  "hsl(var(--primary))",
  "hsl(var(--chart-2, 220 70% 50%))",
  "hsl(var(--chart-3, 280 65% 60%))",
  "hsl(var(--chart-4, 30 80% 55%))",
]

export default function DevicesPieChart({ data, isLoading }: DevicesPieChartProps) {
  const chartData = data
    ? [
        { name: "Desktop", value: data.desktop },
        { name: "Mobile", value: data.mobile },
        { name: "Tablet", value: data.tablet },
        { name: "Other", value: data.other },
      ].filter((d) => d.value > 0)
    : []

  const total = chartData.reduce((sum, d) => sum + d.value, 0)

  return (
    <FeatureGate feature="advanced_analytics" upgradeVariant="card">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Devices</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex h-[200px] items-center justify-center">
              <div className="h-32 w-32 animate-pulse rounded-full bg-muted" />
            </div>
          ) : total === 0 ? (
            <p className="text-sm text-muted-foreground">No device data yet</p>
          ) : (
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={chartData}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={80}
                  paddingAngle={2}
                  dataKey="value"
                >
                  {chartData.map((_, index) => (
                    <Cell key={index} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value) => [
                    `${value} (${((Number(value) / total) * 100).toFixed(1)}%)`,
                    "Clicks",
                  ]}
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>
    </FeatureGate>
  )
}
