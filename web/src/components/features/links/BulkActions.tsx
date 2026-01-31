import { useDeleteLink } from "@/hooks/useLinks"
import { Button } from "@/components/ui/button"

interface BulkActionsProps {
  selectedCount: number
  selectedIds: Set<string>
  onClear: () => void
}

export default function BulkActions({ selectedCount, selectedIds, onClear }: BulkActionsProps) {
  const deleteLink = useDeleteLink()

  if (selectedCount === 0) return null

  function handleBulkDelete() {
    if (!window.confirm(`Are you sure you want to delete ${selectedCount} link(s)?`)) {
      return
    }

    const ids = Array.from(selectedIds)
    ids.forEach((id) => deleteLink.mutate(id))
    onClear()
  }

  return (
    <div className="flex items-center gap-3 rounded-lg border bg-muted/50 p-3">
      <span className="text-sm font-medium">
        {selectedCount} link{selectedCount > 1 ? "s" : ""} selected
      </span>
      <div className="flex gap-2">
        <Button
          variant="destructive"
          size="sm"
          onClick={handleBulkDelete}
          disabled={deleteLink.isPending}
        >
          Delete Selected
        </Button>
        <Button variant="ghost" size="sm" onClick={onClear}>
          Clear Selection
        </Button>
      </div>
    </div>
  )
}
