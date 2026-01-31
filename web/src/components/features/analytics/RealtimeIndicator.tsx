import type { ClickNotification } from "@/types/analytics"
import { formatDistanceToNow } from "date-fns"

interface RealtimeIndicatorProps {
  isConnected: boolean
  recentClicks: ClickNotification[]
}

export default function RealtimeIndicator({
  isConnected,
  recentClicks,
}: RealtimeIndicatorProps) {
  return (
    <div className="flex items-center gap-3">
      <div className="flex items-center gap-1.5">
        <div
          className={`h-2 w-2 rounded-full ${
            isConnected ? "bg-green-500" : "bg-red-500"
          }`}
        />
        <span className="text-xs text-muted-foreground">
          {isConnected ? "Live" : "Disconnected"}
        </span>
      </div>

      {recentClicks.length > 0 && (
        <div className="flex items-center gap-1 text-xs text-muted-foreground">
          <span>Latest:</span>
          <span className="font-medium">
            /{recentClicks[0].short_code}
          </span>
          <span>
            {formatDistanceToNow(new Date(recentClicks[0].timestamp), {
              addSuffix: true,
            })}
          </span>
        </div>
      )}
    </div>
  )
}
