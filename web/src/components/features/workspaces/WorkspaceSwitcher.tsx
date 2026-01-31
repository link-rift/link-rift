import { useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { useCreateWorkspace } from "@/hooks/useWorkspace"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { Workspace } from "@/types/workspace"

export default function WorkspaceSwitcher() {
  const { workspaces, currentWorkspace, setCurrentWorkspace } = useWorkspaceStore()
  const queryClient = useQueryClient()
  const createWorkspace = useCreateWorkspace()
  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState("")

  function handleSwitch(workspace: Workspace) {
    setCurrentWorkspace(workspace)
    queryClient.invalidateQueries({ queryKey: ["links"] })
  }

  function handleCreate() {
    if (!newName.trim()) return
    createWorkspace.mutate(
      { name: newName.trim() },
      {
        onSuccess: (ws) => {
          setCurrentWorkspace(ws)
          queryClient.invalidateQueries({ queryKey: ["links"] })
          setCreateOpen(false)
          setNewName("")
        },
      }
    )
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" className="max-w-[180px] truncate">
            {currentWorkspace?.name ?? "Select workspace"}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-56">
          {workspaces.map((ws) => (
            <DropdownMenuItem
              key={ws.id}
              onClick={() => handleSwitch(ws)}
              className={ws.id === currentWorkspace?.id ? "bg-muted" : ""}
            >
              <span className="truncate">{ws.name}</span>
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => setCreateOpen(true)}>
            + Create Workspace
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Create Workspace</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ws-name">Workspace Name</Label>
              <Input
                id="ws-name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="My Workspace"
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={!newName.trim() || createWorkspace.isPending}
              >
                {createWorkspace.isPending ? "Creating..." : "Create"}
              </Button>
            </div>
            {createWorkspace.isError && (
              <p className="text-sm text-destructive">
                {createWorkspace.error instanceof Error
                  ? createWorkspace.error.message
                  : "Failed to create workspace"}
              </p>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
