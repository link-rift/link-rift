import { useState } from "react"
import { useDNSRecords } from "@/hooks/useDomains"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Skeleton } from "@/components/ui/skeleton"

interface DNSInstructionsProps {
  domainId: string
  domainName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function DNSInstructions({
  domainId,
  domainName,
  open,
  onOpenChange,
}: DNSInstructionsProps) {
  const { data: instructions, isLoading } = useDNSRecords(
    open ? domainId : ""
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>DNS Configuration for {domainName}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Add the following DNS records to your domain provider to verify
            ownership and point traffic to Linkrift.
          </p>
          {isLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-20 w-full" />
            </div>
          ) : instructions?.records ? (
            <div className="space-y-4">
              {instructions.records.map((record, idx) => (
                <DNSRecordRow
                  key={idx}
                  type={record.type}
                  host={record.host}
                  value={record.value}
                />
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              No DNS records available.
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

function DNSRecordRow({
  type,
  host,
  value,
}: {
  type: string
  host: string
  value: string
}) {
  const [copied, setCopied] = useState<"host" | "value" | null>(null)

  const copyToClipboard = (text: string, field: "host" | "value") => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(field)
      setTimeout(() => setCopied(null), 2000)
    })
  }

  return (
    <div className="rounded-md border p-3 space-y-2">
      <div className="flex items-center gap-2">
        <span className="rounded bg-muted px-2 py-0.5 text-xs font-mono font-semibold">
          {type}
        </span>
      </div>
      <div className="space-y-1">
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">Host</p>
            <p className="truncate font-mono text-sm">{host}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="shrink-0"
            onClick={() => copyToClipboard(host, "host")}
          >
            {copied === "host" ? "Copied" : "Copy"}
          </Button>
        </div>
        <div className="flex items-center justify-between gap-2">
          <div className="min-w-0 flex-1">
            <p className="text-xs text-muted-foreground">Value</p>
            <p className="truncate font-mono text-sm">{value}</p>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="shrink-0"
            onClick={() => copyToClipboard(value, "value")}
          >
            {copied === "value" ? "Copied" : "Copy"}
          </Button>
        </div>
      </div>
    </div>
  )
}
