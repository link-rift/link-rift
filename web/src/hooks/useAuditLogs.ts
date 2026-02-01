import { useQuery } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as auditLogService from "@/services/auditlogs"
import type { AuditLogFilter } from "@/types/auditlog"

export function useAuditLogs(
  filter?: AuditLogFilter & { limit?: number; offset?: number }
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["audit-logs", wsId, filter],
    queryFn: () => auditLogService.getAuditLogs(filter),
    staleTime: 30 * 1000,
    enabled: !!wsId,
  })
}

export function useAuditLog(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["audit-logs", wsId, id],
    queryFn: () => auditLogService.getAuditLog(id),
    enabled: !!id && !!wsId,
  })
}
