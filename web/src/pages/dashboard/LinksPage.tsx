import { useState, useMemo } from "react"
import { useLinks } from "@/hooks/useLinks"
import { useLinkStore } from "@/stores/linkStore"
import { Button } from "@/components/ui/button"
import LinkList from "@/components/features/links/LinkList"
import LinkFilters from "@/components/features/links/LinkFilters"
import BulkActions from "@/components/features/links/BulkActions"
import CreateLinkModal from "@/components/features/links/CreateLinkModal"
import EditLinkModal from "@/components/features/links/EditLinkModal"
import type { Link } from "@/types/link"

const PAGE_SIZE = 20

export default function LinksPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [editLink, setEditLink] = useState<Link | null>(null)
  const [search, setSearch] = useState("")
  const [activeFilter, setActiveFilter] = useState<boolean | undefined>(undefined)
  const [offset, setOffset] = useState(0)

  const { selectedIds, toggleSelected, clearSelected } = useLinkStore()

  const queryParams = useMemo(
    () => ({
      search: search || undefined,
      is_active: activeFilter,
      limit: PAGE_SIZE,
      offset,
    }),
    [search, activeFilter, offset]
  )

  const { data, isLoading } = useLinks(queryParams)
  const links = data?.links ?? []
  const total = data?.total ?? 0
  const hasMore = offset + PAGE_SIZE < total

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Links</h1>
          <p className="text-sm text-muted-foreground">
            Manage your shortened links
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create Link
        </Button>
      </div>

      <LinkFilters
        search={search}
        onSearchChange={(v) => {
          setSearch(v)
          setOffset(0)
        }}
        showActive={activeFilter}
        onActiveChange={(v) => {
          setActiveFilter(v)
          setOffset(0)
        }}
      />

      <BulkActions
        selectedCount={selectedIds.size}
        selectedIds={selectedIds}
        onClear={clearSelected}
      />

      <LinkList
        links={links}
        isLoading={isLoading}
        selectedIds={selectedIds}
        onSelect={toggleSelected}
        onEdit={(link) => setEditLink(link)}
      />

      {total > PAGE_SIZE && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {offset + 1}-{Math.min(offset + PAGE_SIZE, total)} of {total} links
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={offset === 0}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={!hasMore}
              onClick={() => setOffset(offset + PAGE_SIZE)}
            >
              Next
            </Button>
          </div>
        </div>
      )}

      <CreateLinkModal open={createOpen} onOpenChange={setCreateOpen} />
      <EditLinkModal
        link={editLink}
        open={!!editLink}
        onOpenChange={(open) => {
          if (!open) setEditLink(null)
        }}
      />
    </div>
  )
}
