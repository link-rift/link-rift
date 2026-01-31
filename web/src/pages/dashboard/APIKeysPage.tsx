import { useState } from "react"
import { useAPIKeys } from "@/hooks/useAPIKeys"
import { FeatureGate } from "@/components/ui/FeatureGate"
import { Button } from "@/components/ui/button"
import APIKeyList from "@/components/features/apikeys/APIKeyList"
import CreateAPIKeyModal from "@/components/features/apikeys/CreateAPIKeyModal"

export default function APIKeysPage() {
  return (
    <FeatureGate feature="api_access">
      <APIKeysContent />
    </FeatureGate>
  )
}

function APIKeysContent() {
  const [createOpen, setCreateOpen] = useState(false)
  const { data: apiKeys, isLoading } = useAPIKeys()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">API Keys</h1>
          <p className="text-sm text-muted-foreground">
            Manage API keys for programmatic access to the Linkrift API
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <svg
            className="mr-2 h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          Create API Key
        </Button>
      </div>

      <APIKeyList apiKeys={apiKeys ?? []} isLoading={isLoading} />

      <CreateAPIKeyModal open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  )
}
