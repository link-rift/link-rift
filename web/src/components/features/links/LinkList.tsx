import { Skeleton } from "@/components/ui/skeleton"
import LinkRow from "./LinkRow"
import type { Link } from "@/types/link"

interface LinkListProps {
  links: Link[]
  isLoading: boolean
  selectedIds: Set<string>
  onSelect: (id: string) => void
  onEdit: (link: Link) => void
}

export default function LinkList({ links, isLoading, selectedIds, onSelect, onEdit }: LinkListProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-3 rounded-lg border p-3">
            <Skeleton className="h-4 w-4" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-4 w-48" />
              <Skeleton className="h-3 w-64" />
            </div>
            <Skeleton className="h-8 w-16" />
          </div>
        ))}
      </div>
    )
  }

  if (links.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12">
        <svg className="mb-4 h-12 w-12 text-muted-foreground/50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
        <h3 className="text-lg font-medium">No links yet</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create your first shortened link to get started.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {links.map((link) => (
        <LinkRow
          key={link.id}
          link={link}
          selected={selectedIds.has(link.id)}
          onSelect={onSelect}
          onEdit={onEdit}
        />
      ))}
    </div>
  )
}
