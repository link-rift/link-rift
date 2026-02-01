import type { AuditLog } from "@/types/auditlog"
import { ACTION_LABELS, RESOURCE_TYPE_LABELS } from "@/types/auditlog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

interface AuditLogDetailProps {
  log: AuditLog
  onClose: () => void
}

export default function AuditLogDetail({ log, onClose }: AuditLogDetailProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-lg">Audit Log Detail</CardTitle>
        <Button variant="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Action</span>
            <div className="mt-1">
              <Badge variant="secondary">
                {ACTION_LABELS[log.action] ?? log.action}
              </Badge>
            </div>
          </div>
          <div>
            <span className="text-muted-foreground">Resource</span>
            <div className="mt-1">
              {RESOURCE_TYPE_LABELS[log.resource_type] ?? log.resource_type}
            </div>
          </div>
          <div>
            <span className="text-muted-foreground">Resource ID</span>
            <div className="mt-1 font-mono text-xs">
              {log.resource_id ?? "-"}
            </div>
          </div>
          <div>
            <span className="text-muted-foreground">User ID</span>
            <div className="mt-1 font-mono text-xs">{log.user_id ?? "-"}</div>
          </div>
          <div>
            <span className="text-muted-foreground">IP Address</span>
            <div className="mt-1">{log.ip_address || "-"}</div>
          </div>
          <div>
            <span className="text-muted-foreground">Time</span>
            <div className="mt-1">
              {new Date(log.created_at).toLocaleString()}
            </div>
          </div>
        </div>

        {log.old_values && Object.keys(log.old_values).length > 0 && (
          <div>
            <h4 className="mb-1 text-sm font-medium text-muted-foreground">
              Previous Values
            </h4>
            <pre className="rounded-md bg-muted p-3 text-xs overflow-auto max-h-40">
              {JSON.stringify(log.old_values, null, 2)}
            </pre>
          </div>
        )}

        {log.new_values && Object.keys(log.new_values).length > 0 && (
          <div>
            <h4 className="mb-1 text-sm font-medium text-muted-foreground">
              New Values
            </h4>
            <pre className="rounded-md bg-muted p-3 text-xs overflow-auto max-h-40">
              {JSON.stringify(log.new_values, null, 2)}
            </pre>
          </div>
        )}

        {log.metadata && Object.keys(log.metadata).length > 0 && (
          <div>
            <h4 className="mb-1 text-sm font-medium text-muted-foreground">
              Metadata
            </h4>
            <pre className="rounded-md bg-muted p-3 text-xs overflow-auto max-h-40">
              {JSON.stringify(log.metadata, null, 2)}
            </pre>
          </div>
        )}

        {log.user_agent && (
          <div>
            <h4 className="mb-1 text-sm font-medium text-muted-foreground">
              User Agent
            </h4>
            <p className="text-xs text-muted-foreground break-all">
              {log.user_agent}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
