import { FeatureGate } from "@/components/ui/FeatureGate"
import SCIMTokenList from "@/components/features/scim/SCIMTokenList"
import { useSCIMSyncLogs } from "@/hooks/useSCIM"
import { useWorkspaceStore } from "@/stores/workspaceStore"

export default function SCIMSettingsPage() {
  return (
    <FeatureGate feature="scim" tier="enterprise">
      <SCIMSettingsContent />
    </FeatureGate>
  )
}

function SCIMSettingsContent() {
  const { currentWorkspace } = useWorkspaceStore()
  const { data: syncLogs, isLoading: logsLoading } = useSCIMSyncLogs()

  const wsId = currentWorkspace?.id
  const baseUrl = window.location.origin

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">
          SCIM Provisioning
        </h2>
        <p className="text-sm text-muted-foreground">
          Automate user provisioning and deprovisioning via SCIM 2.0
        </p>
      </div>

      {wsId && (
        <div className="rounded-lg border bg-muted/50 p-4">
          <h3 className="mb-2 text-sm font-medium">SCIM Configuration</h3>
          <p className="mb-1 text-xs text-muted-foreground">
            Use these values to configure your Identity Provider:
          </p>
          <div className="space-y-1 font-mono text-xs">
            <div>
              <span className="text-muted-foreground">SCIM Base URL: </span>
              {baseUrl}/scim/v2/{wsId}
            </div>
            <div>
              <span className="text-muted-foreground">Users Endpoint: </span>
              {baseUrl}/scim/v2/{wsId}/Users
            </div>
            <div>
              <span className="text-muted-foreground">Groups Endpoint: </span>
              {baseUrl}/scim/v2/{wsId}/Groups
            </div>
          </div>
        </div>
      )}

      <div>
        <h3 className="mb-3 text-lg font-semibold">Bearer Tokens</h3>
        <SCIMTokenList />
      </div>

      <div>
        <h3 className="mb-3 text-lg font-semibold">Sync Log</h3>
        {logsLoading ? (
          <div className="py-4 text-sm text-muted-foreground">
            Loading sync logs...
          </div>
        ) : syncLogs && syncLogs.data.length > 0 ? (
          <div className="overflow-hidden rounded-lg border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  <th className="px-3 py-2 text-left font-medium">Time</th>
                  <th className="px-3 py-2 text-left font-medium">Operation</th>
                  <th className="px-3 py-2 text-left font-medium">Resource</th>
                  <th className="px-3 py-2 text-left font-medium">Status</th>
                  <th className="px-3 py-2 text-left font-medium">Details</th>
                </tr>
              </thead>
              <tbody>
                {syncLogs.data.map((log) => (
                  <tr key={log.id} className="border-t">
                    <td className="px-3 py-2 text-xs text-muted-foreground">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-3 py-2 font-mono text-xs">
                      {log.operation}
                    </td>
                    <td className="px-3 py-2 text-xs">
                      {log.resource_type}
                      {log.external_id && (
                        <span className="ml-1 text-muted-foreground">
                          ({log.external_id})
                        </span>
                      )}
                    </td>
                    <td className="px-3 py-2">
                      <span
                        className={`rounded px-1.5 py-0.5 text-xs ${
                          log.status === "success"
                            ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                            : "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                        }`}
                      >
                        {log.status}
                      </span>
                    </td>
                    <td className="max-w-xs truncate px-3 py-2 font-mono text-xs text-muted-foreground">
                      {JSON.stringify(log.details)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="py-4 text-center text-sm text-muted-foreground">
            No sync activity yet. SCIM operations will appear here.
          </p>
        )}
      </div>
    </div>
  )
}
