import { useState } from "react"
import { useWebhookDeliveries } from "@/hooks/useWebhooks"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"

interface DeliveryLogProps {
  webhookId: string
}

export default function DeliveryLog({ webhookId }: DeliveryLogProps) {
  const [page, setPage] = useState(0)
  const limit = 10
  const { data, isLoading } = useWebhookDeliveries(webhookId, {
    limit,
    offset: page * limit,
  })

  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    )
  }

  if (!data || data.deliveries.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-muted-foreground">
        No delivery attempts yet.
      </p>
    )
  }

  const totalPages = Math.ceil(data.total / limit)

  return (
    <div className="space-y-3">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Event</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Attempts</TableHead>
            <TableHead>Response</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.deliveries.map((delivery) => (
            <TableRow key={delivery.id}>
              <TableCell>
                <code className="text-xs">{delivery.event}</code>
              </TableCell>
              <TableCell>
                <StatusBadge
                  status={delivery.response_status}
                  completed={!!delivery.completed_at}
                />
              </TableCell>
              <TableCell className="text-sm">
                {delivery.attempts}/{delivery.max_attempts}
              </TableCell>
              <TableCell className="max-w-xs truncate text-xs text-muted-foreground">
                {delivery.response_body
                  ? delivery.response_body.slice(0, 100)
                  : delivery.completed_at
                    ? "No response body"
                    : "Pending..."}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {new Date(delivery.created_at).toLocaleString()}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">
            Showing {page * limit + 1}-{Math.min((page + 1) * limit, data.total)} of{" "}
            {data.total}
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= totalPages - 1}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function StatusBadge({
  status,
  completed,
}: {
  status?: number | null
  completed: boolean
}) {
  if (!completed) {
    return (
      <Badge variant="outline" className="text-xs">
        Pending
      </Badge>
    )
  }

  if (!status) {
    return (
      <Badge variant="destructive" className="text-xs">
        Failed
      </Badge>
    )
  }

  if (status >= 200 && status < 300) {
    return (
      <Badge className="bg-green-600 text-xs hover:bg-green-700">
        {status}
      </Badge>
    )
  }

  return (
    <Badge variant="destructive" className="text-xs">
      {status}
    </Badge>
  )
}
