import type { AuditLog } from "@/types/auditlog"
import { ACTION_LABELS, RESOURCE_TYPE_LABELS } from "@/types/auditlog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"

interface AuditLogTableProps {
  logs: AuditLog[]
  onViewDetail: (log: AuditLog) => void
}

const actionColors: Record<string, string> = {
  create: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  update: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
  delete: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
  revoke: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
}

export default function AuditLogTable({ logs, onViewDetail }: AuditLogTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Time</TableHead>
          <TableHead>Action</TableHead>
          <TableHead>Resource</TableHead>
          <TableHead>User</TableHead>
          <TableHead>IP Address</TableHead>
          <TableHead className="w-20"></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {logs.length === 0 && (
          <TableRow>
            <TableCell colSpan={6} className="text-center text-muted-foreground">
              No audit logs found
            </TableCell>
          </TableRow>
        )}
        {logs.map((log) => (
          <TableRow key={log.id}>
            <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
              {new Date(log.created_at).toLocaleString()}
            </TableCell>
            <TableCell>
              <Badge
                variant="secondary"
                className={actionColors[log.action] ?? ""}
              >
                {ACTION_LABELS[log.action] ?? log.action}
              </Badge>
            </TableCell>
            <TableCell>
              <span className="text-sm">
                {RESOURCE_TYPE_LABELS[log.resource_type] ?? log.resource_type}
              </span>
              {log.resource_id && (
                <span className="ml-1 text-xs text-muted-foreground">
                  {log.resource_id.substring(0, 8)}...
                </span>
              )}
            </TableCell>
            <TableCell className="text-xs text-muted-foreground">
              {log.user_id ? `${log.user_id.substring(0, 8)}...` : "-"}
            </TableCell>
            <TableCell className="text-xs text-muted-foreground">
              {log.ip_address || "-"}
            </TableCell>
            <TableCell>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onViewDetail(log)}
              >
                View
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
