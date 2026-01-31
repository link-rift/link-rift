import { useState } from "react"
import { useWebhooks } from "@/hooks/useWebhooks"
import { FeatureGate } from "@/components/ui/FeatureGate"
import { Button } from "@/components/ui/button"
import WebhookList from "@/components/features/webhooks/WebhookList"
import CreateWebhookModal from "@/components/features/webhooks/CreateWebhookModal"

export default function WebhooksPage() {
  return (
    <FeatureGate feature="webhooks">
      <WebhooksContent />
    </FeatureGate>
  )
}

function WebhooksContent() {
  const [createOpen, setCreateOpen] = useState(false)
  const { data: webhooks, isLoading } = useWebhooks()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Webhooks</h1>
          <p className="text-sm text-muted-foreground">
            Receive real-time event notifications via HTTP webhooks
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
          Create Webhook
        </Button>
      </div>

      <WebhookList webhooks={webhooks ?? []} isLoading={isLoading} />

      <CreateWebhookModal open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  )
}
