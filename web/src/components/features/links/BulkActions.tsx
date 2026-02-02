import { useState } from "react"
import { useDeleteLink } from "@/hooks/useLinks"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"

interface BulkActionsProps {
  selectedCount: number
  selectedIds: Set<string>
  onClear: () => void
}

export default function BulkActions({ selectedCount, selectedIds, onClear }: BulkActionsProps) {
  const deleteLink = useDeleteLink()
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  if (selectedCount === 0) return null

  function handleBulkDelete() {
    const ids = Array.from(selectedIds)
    ids.forEach((id) => deleteLink.mutate(id))
    onClear()
    setShowDeleteDialog(false)
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
          onClick={() => setShowDeleteDialog(true)}
          disabled={deleteLink.isPending}
        >
          Delete Selected
        </Button>
        <Button variant="ghost" size="sm" onClick={onClear}>
          Clear Selection
        </Button>
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {selectedCount} link{selectedCount > 1 ? "s" : ""}</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete {selectedCount} link{selectedCount > 1 ? "s" : ""}? This action cannot be undone and all associated analytics data will be lost.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleBulkDelete}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
