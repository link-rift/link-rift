import { useState, lazy, Suspense } from "react"
import { useNavigate } from "react-router-dom"
import { useDeleteLink } from "@/hooks/useLinks"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Link } from "@/types/link"

const QRCodeModal = lazy(() => import("@/components/features/qrcodes/QRCodeModal"))

interface LinkRowProps {
  link: Link
  selected: boolean
  onSelect: (id: string) => void
  onEdit: (link: Link) => void
}

export default function LinkRow({ link, selected, onSelect, onEdit }: LinkRowProps) {
  const navigate = useNavigate()
  const deleteLink = useDeleteLink()
  const [copied, setCopied] = useState(false)
  const [showQRModal, setShowQRModal] = useState(false)

  function copyShortUrl() {
    navigator.clipboard.writeText(link.short_url)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function handleDelete() {
    if (window.confirm("Are you sure you want to delete this link?")) {
      deleteLink.mutate(link.id)
    }
  }

  const isExpired = link.expires_at && new Date(link.expires_at) < new Date()
  const isOverLimit = link.max_clicks != null && link.total_clicks >= link.max_clicks

  return (
    <div className="flex items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted/50">
      <Checkbox
        checked={selected}
        onCheckedChange={() => onSelect(link.id)}
      />

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <button
            onClick={copyShortUrl}
            className="text-sm font-medium text-primary hover:underline"
            title="Click to copy"
          >
            {link.short_url}
          </button>
          {copied && <span className="text-xs text-muted-foreground">Copied!</span>}
          {link.has_password && (
            <Badge variant="outline" className="text-xs">
              Protected
            </Badge>
          )}
          {!link.is_active && (
            <Badge variant="secondary" className="text-xs">
              Disabled
            </Badge>
          )}
          {isExpired && (
            <Badge variant="destructive" className="text-xs">
              Expired
            </Badge>
          )}
          {isOverLimit && (
            <Badge variant="destructive" className="text-xs">
              Limit Reached
            </Badge>
          )}
        </div>
        <p className="truncate text-sm text-muted-foreground" title={link.url}>
          {link.title || link.url}
        </p>
      </div>

      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <div className="text-right">
          <p className="font-medium text-foreground">{link.total_clicks}</p>
          <p className="text-xs">clicks</p>
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01" />
              </svg>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={copyShortUrl}>
              Copy Short URL
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => window.open(link.url, "_blank")}>
              Open Destination
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => navigate(`/analytics/${link.id}`)}>
              View Analytics
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setShowQRModal(true)}>
              QR Code
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => onEdit(link)}>
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleDelete}
              className="text-destructive focus:text-destructive"
            >
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {showQRModal && (
        <Suspense fallback={null}>
          <QRCodeModal
            link={link}
            open={showQRModal}
            onClose={() => setShowQRModal(false)}
          />
        </Suspense>
      )}
    </div>
  )
}
