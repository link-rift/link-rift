import { useState } from "react"
import { useVerifyDomain, useDeleteDomain } from "@/hooks/useDomains"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import DNSInstructions from "./DNSInstructions"
import type { Domain } from "@/types/domain"

interface DomainListProps {
  domains: Domain[]
  isLoading: boolean
}

export default function DomainList({ domains, isLoading }: DomainListProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </div>
    )
  }

  if (domains.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-md border border-dashed py-12">
        <p className="text-sm text-muted-foreground">
          No custom domains configured yet.
        </p>
        <p className="text-xs text-muted-foreground mt-1">
          Add a domain to start using branded short links.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {domains.map((domain) => (
        <DomainCard key={domain.id} domain={domain} />
      ))}
    </div>
  )
}

function DomainCard({ domain }: { domain: Domain }) {
  const [dnsOpen, setDnsOpen] = useState(false)
  const verifyDomain = useVerifyDomain()
  const deleteDomain = useDeleteDomain()

  const handleVerify = async () => {
    try {
      await verifyDomain.mutateAsync(domain.id)
    } catch {
      // Error state is available via verifyDomain.error
    }
  }

  const handleDelete = async () => {
    if (!window.confirm(`Remove domain "${domain.domain}"? This cannot be undone.`)) {
      return
    }
    try {
      await deleteDomain.mutateAsync(domain.id)
    } catch {
      // Error state is available via deleteDomain.error
    }
  }

  return (
    <>
      <div className="flex items-center justify-between rounded-md border p-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="font-medium">{domain.domain}</span>
            {domain.is_verified ? (
              <Badge variant="default" className="bg-green-600 hover:bg-green-600">
                Verified
              </Badge>
            ) : (
              <Badge variant="secondary">Unverified</Badge>
            )}
            {domain.is_verified && (
              <SSLBadge status={domain.ssl_status} />
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            Added {new Date(domain.created_at).toLocaleDateString()}
          </p>
          {verifyDomain.isError && verifyDomain.variables === domain.id && (
            <p className="text-xs text-destructive">
              {verifyDomain.error instanceof Error
                ? verifyDomain.error.message
                : "Verification failed"}
            </p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setDnsOpen(true)}
          >
            DNS Records
          </Button>
          {!domain.is_verified && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleVerify}
              disabled={verifyDomain.isPending}
            >
              {verifyDomain.isPending && verifyDomain.variables === domain.id
                ? "Verifying..."
                : "Verify"}
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={handleDelete}
            disabled={deleteDomain.isPending}
          >
            {deleteDomain.isPending && deleteDomain.variables === domain.id
              ? "Removing..."
              : "Remove"}
          </Button>
        </div>
      </div>
      <DNSInstructions
        domainId={domain.id}
        domainName={domain.domain}
        open={dnsOpen}
        onOpenChange={setDnsOpen}
      />
    </>
  )
}

function SSLBadge({ status }: { status: string }) {
  switch (status) {
    case "active":
      return (
        <Badge variant="outline" className="border-green-600 text-green-600">
          SSL Active
        </Badge>
      )
    case "pending":
      return (
        <Badge variant="outline" className="border-yellow-600 text-yellow-600">
          SSL Pending
        </Badge>
      )
    case "failed":
      return (
        <Badge variant="outline" className="border-destructive text-destructive">
          SSL Failed
        </Badge>
      )
    default:
      return null
  }
}
