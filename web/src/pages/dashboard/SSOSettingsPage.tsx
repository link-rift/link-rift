import {
  useSSOConfig,
  useCreateSSOConfig,
  useUpdateSSOConfig,
  useDeleteSSOConfig,
} from "@/hooks/useSSO"
import { FeatureGate } from "@/components/ui/FeatureGate"
import SSOConfigForm from "@/components/features/sso/SSOConfigForm"
import { Button } from "@/components/ui/button"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { CreateSSOConfigInput } from "@/types/sso"

export default function SSOSettingsPage() {
  return (
    <FeatureGate feature="saml" tier="enterprise">
      <SSOSettingsContent />
    </FeatureGate>
  )
}

function SSOSettingsContent() {
  const { data: config, isLoading } = useSSOConfig()
  const createMutation = useCreateSSOConfig()
  const updateMutation = useUpdateSSOConfig()
  const deleteMutation = useDeleteSSOConfig()
  const { currentWorkspace } = useWorkspaceStore()

  const handleSubmit = (input: CreateSSOConfigInput) => {
    if (config) {
      updateMutation.mutate(input)
    } else {
      createMutation.mutate(input)
    }
  }

  const handleDelete = () => {
    if (confirm("Are you sure you want to remove the SSO configuration?")) {
      deleteMutation.mutate()
    }
  }

  if (isLoading) {
    return (
      <div className="py-12 text-center text-muted-foreground">
        Loading SSO settings...
      </div>
    )
  }

  const baseUrl = window.location.origin
  const wsId = currentWorkspace?.id

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            SSO / SAML Settings
          </h2>
          <p className="text-sm text-muted-foreground">
            Configure Single Sign-On with your Identity Provider
          </p>
        </div>
        {config && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? "Removing..." : "Remove SSO"}
          </Button>
        )}
      </div>

      {wsId && (
        <div className="rounded-lg border bg-muted/50 p-4">
          <h3 className="mb-2 text-sm font-medium">
            Service Provider (SP) Details
          </h3>
          <p className="mb-1 text-xs text-muted-foreground">
            Use these values to configure your Identity Provider:
          </p>
          <div className="space-y-1 font-mono text-xs">
            <div>
              <span className="text-muted-foreground">SP Entity ID: </span>
              {baseUrl}/api/v1/auth/sso/metadata/{wsId}
            </div>
            <div>
              <span className="text-muted-foreground">ACS URL: </span>
              {baseUrl}/api/v1/auth/sso/callback/{wsId}
            </div>
            <div>
              <span className="text-muted-foreground">Metadata URL: </span>
              {baseUrl}/api/v1/auth/sso/metadata/{wsId}
            </div>
          </div>
        </div>
      )}

      <SSOConfigForm
        config={config}
        onSubmit={handleSubmit}
        isLoading={createMutation.isPending || updateMutation.isPending}
      />
    </div>
  )
}
