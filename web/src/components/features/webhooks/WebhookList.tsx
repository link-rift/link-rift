import { useState } from "react"
import { useDeleteWebhook } from "@/hooks/useWebhooks"
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
import DeliveryLog from "./DeliveryLog"
import type { Webhook } from "@/types/webhook"

interface WebhookListProps {
  webhooks: Webhook[]
  isLoading: boolean
}

export default function WebhookList({ webhooks, isLoading }: WebhookListProps) {
  const deleteWebhook = useDeleteWebhook()
  const [expandedId, setExpandedId] = useState<string | null>(null)

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    )
  }

  if (webhooks.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-sm text-muted-foreground">
          No webhooks configured. Create one to receive real-time event notifications.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>URL</TableHead>
            <TableHead>Events</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Failures</TableHead>
            <TableHead>Last Triggered</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {webhooks.map((webhook) => (
            <>
              <TableRow key={webhook.id}>
                <TableCell className="max-w-xs truncate font-mono text-sm">
                  {webhook.url}
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {webhook.events.slice(0, 2).map((event) => (
                      <Badge key={event} variant="secondary" className="text-xs">
                        {event}
                      </Badge>
                    ))}
                    {webhook.events.length > 2 && (
                      <Badge variant="outline" className="text-xs">
                        +{webhook.events.length - 2}
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge
                    variant={webhook.is_active ? "default" : "destructive"}
                    className="text-xs"
                  >
                    {webhook.is_active ? "Active" : "Disabled"}
                  </Badge>
                </TableCell>
                <TableCell className="text-sm">
                  {webhook.failure_count > 0 ? (
                    <span className="text-destructive">{webhook.failure_count}</span>
                  ) : (
                    "0"
                  )}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {webhook.last_triggered_at
                    ? new Date(webhook.last_triggered_at).toLocaleString()
                    : "Never"}
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setExpandedId(
                          expandedId === webhook.id ? null : webhook.id
                        )
                      }
                    >
                      {expandedId === webhook.id ? "Hide Logs" : "Delivery Logs"}
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => deleteWebhook.mutate(webhook.id)}
                      disabled={deleteWebhook.isPending}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
              {expandedId === webhook.id && (
                <TableRow key={`${webhook.id}-deliveries`}>
                  <TableCell colSpan={6} className="p-0">
                    <div className="border-t bg-muted/30 p-4">
                      <DeliveryLog webhookId={webhook.id} />
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
