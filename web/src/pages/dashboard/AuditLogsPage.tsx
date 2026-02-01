import { useState, useCallback } from "react"
import { useAuditLogs } from "@/hooks/useAuditLogs"
import { exportAuditLogs } from "@/services/auditlogs"
import { FeatureGate } from "@/components/ui/FeatureGate"
import AuditLogFilters from "@/components/features/auditlogs/AuditLogFilters"
import AuditLogTable from "@/components/features/auditlogs/AuditLogTable"
import AuditLogDetail from "@/components/features/auditlogs/AuditLogDetail"
import { Button } from "@/components/ui/button"
import type { AuditLog, AuditLogFilter } from "@/types/auditlog"

const PAGE_SIZE = 20

export default function AuditLogsPage() {
  return (
    <FeatureGate feature="audit_logs" tier="enterprise">
      <AuditLogsContent />
    </FeatureGate>
  )
}

function AuditLogsContent() {
  const [filter, setFilter] = useState<AuditLogFilter>({})
  const [offset, setOffset] = useState(0)
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)

  const { data, isLoading } = useAuditLogs({
    ...filter,
    limit: PAGE_SIZE,
    offset,
  })

  const handleFilterChange = useCallback((newFilter: AuditLogFilter) => {
    setFilter(newFilter)
    setOffset(0)
  }, [])

  const handleExport = useCallback(
    async (format: "json" | "csv") => {
      try {
        const blob = await exportAuditLogs(filter, format)
        const url = URL.createObjectURL(blob)
        const a = document.createElement("a")
        a.href = url
        a.download = `audit-logs.${format}`
        a.click()
        URL.revokeObjectURL(url)
      } catch {
        // Error handling via UI toast would go here
      }
    },
    [filter]
  )

  const total = data?.total ?? 0
  const logs = data?.audit_logs ?? []
  const hasNext = offset + PAGE_SIZE < total
  const hasPrev = offset > 0

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Audit Logs</h2>
        <p className="text-sm text-muted-foreground">
          Track all actions performed in your workspace
        </p>
      </div>

      <AuditLogFilters
        filter={filter}
        onFilterChange={handleFilterChange}
        onExport={handleExport}
      />

      {selectedLog ? (
        <AuditLogDetail
          log={selectedLog}
          onClose={() => setSelectedLog(null)}
        />
      ) : (
        <>
          {isLoading ? (
            <div className="py-12 text-center text-muted-foreground">
              Loading audit logs...
            </div>
          ) : (
            <AuditLogTable logs={logs} onViewDetail={setSelectedLog} />
          )}

          {total > PAGE_SIZE && (
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                Showing {offset + 1}-{Math.min(offset + PAGE_SIZE, total)} of{" "}
                {total}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!hasPrev}
                  onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={!hasNext}
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
