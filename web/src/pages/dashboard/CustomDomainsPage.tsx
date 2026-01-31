import { useState } from "react"
import { useDomains } from "@/hooks/useDomains"
import { FeatureGate } from "@/components/ui/FeatureGate"
import { Button } from "@/components/ui/button"
import DomainList from "@/components/features/domains/DomainList"
import AddDomainModal from "@/components/features/domains/AddDomainModal"

export default function CustomDomainsPage() {
  return (
    <FeatureGate feature="custom_domains">
      <DomainsContent />
    </FeatureGate>
  )
}

function DomainsContent() {
  const [addOpen, setAddOpen] = useState(false)
  const { data: domains, isLoading } = useDomains()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Custom Domains</h1>
          <p className="text-sm text-muted-foreground">
            Connect your own domains for branded short links
          </p>
        </div>
        <Button onClick={() => setAddOpen(true)}>
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
          Add Domain
        </Button>
      </div>

      <DomainList domains={domains ?? []} isLoading={isLoading} />

      <AddDomainModal open={addOpen} onOpenChange={setAddOpen} />
    </div>
  )
}
