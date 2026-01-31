import { useMemo, useState } from "react"
import { format } from "date-fns"
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Brush,
} from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import type { TimeSeriesPoint, TimeSeriesInterval } from "@/types/analytics"

interface ClicksChartProps {
  data: TimeSeriesPoint[] | undefined
  isLoading: boolean
  interval: TimeSeriesInterval
  onIntervalChange: (interval: TimeSeriesInterval) => void
}

const intervals: { label: string; value: TimeSeriesInterval }[] = [
  { label: "Hourly", value: "hour" },
  { label: "Daily", value: "day" },
  { label: "Weekly", value: "week" },
  { label: "Monthly", value: "month" },
]

export default function ClicksChart({
  data,
  isLoading,
  interval,
  onIntervalChange,
}: ClicksChartProps) {
  const [showUnique, setShowUnique] = useState(false)

  const chartData = useMemo(() => {
    if (!data) return []
    return data.map((point) => ({
      ...point,
      date: format(new Date(point.timestamp), interval === "hour" ? "HH:mm" : "MMM d"),
    }))
  }, [data, interval])

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-base">Click Trends</CardTitle>
        <div className="flex items-center gap-2">
          <Button
            variant={showUnique ? "outline" : "default"}
            size="sm"
            onClick={() => setShowUnique(false)}
          >
            Total
          </Button>
          <Button
            variant={showUnique ? "default" : "outline"}
            size="sm"
            onClick={() => setShowUnique(true)}
          >
            Unique
          </Button>
          <span className="mx-1 h-4 w-px bg-border" />
          {intervals.map((i) => (
            <Button
              key={i.value}
              variant={interval === i.value ? "default" : "ghost"}
              size="sm"
              onClick={() => onIntervalChange(i.value)}
            >
              {i.label}
            </Button>
          ))}
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="h-[300px] animate-pulse rounded bg-muted" />
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="clickGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis dataKey="date" className="text-xs" />
              <YAxis className="text-xs" />
              <Tooltip
                contentStyle={{
                  backgroundColor: "hsl(var(--card))",
                  border: "1px solid hsl(var(--border))",
                  borderRadius: "6px",
                }}
              />
              <Area
                type="monotone"
                dataKey={showUnique ? "unique" : "clicks"}
                stroke="hsl(var(--primary))"
                fill="url(#clickGradient)"
                strokeWidth={2}
              />
              {chartData.length > 30 && (
                <Brush dataKey="date" height={30} stroke="hsl(var(--muted-foreground))" />
              )}
            </AreaChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
