import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import {
  useUpdateWorkspace,
  useDeleteWorkspace,
  useWorkspaceMembers,
  useTransferOwnership,
} from "@/hooks/useWorkspace"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export default function WorkspaceSettingsPage() {
  const navigate = useNavigate()
  const { currentWorkspace, canManageMembers, isOwner } = useWorkspaceStore()
  const updateWorkspace = useUpdateWorkspace()
  const deleteWorkspace = useDeleteWorkspace()
  const transferOwnership = useTransferOwnership()
  const { data: members } = useWorkspaceMembers(currentWorkspace?.id ?? "")

  const [name, setName] = useState("")
  const [slug, setSlug] = useState("")
  const [transferTarget, setTransferTarget] = useState("")

  useEffect(() => {
    if (currentWorkspace) {
      setName(currentWorkspace.name)
      setSlug(currentWorkspace.slug)
    }
  }, [currentWorkspace])

  function handleSave() {
    if (!currentWorkspace) return
    updateWorkspace.mutate({
      id: currentWorkspace.id,
      data: { name: name.trim(), slug: slug.trim() },
    })
  }

  function handleDelete() {
    if (!currentWorkspace) return
    if (!confirm(`Delete workspace "${currentWorkspace.name}"? This cannot be undone.`)) return
    deleteWorkspace.mutate(currentWorkspace.id, {
      onSuccess: () => navigate("/"),
    })
  }

  function handleTransfer() {
    if (!currentWorkspace || !transferTarget) return
    if (!confirm("Transfer ownership? You will become an admin.")) return
    transferOwnership.mutate(
      { workspaceId: currentWorkspace.id, data: { new_owner_id: transferTarget } },
      {
        onSuccess: () => setTransferTarget(""),
      }
    )
  }

  const otherMembers = members?.filter((m) => m.role !== "owner") ?? []

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">Workspace Settings</h2>
        <p className="text-sm text-muted-foreground">
          Manage your workspace configuration.
        </p>
      </div>

      {canManageMembers() && (
        <Card>
          <CardHeader>
            <CardTitle>General</CardTitle>
            <CardDescription>Update workspace name and URL slug.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="ws-name">Name</Label>
              <Input
                id="ws-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ws-slug">Slug</Label>
              <Input
                id="ws-slug"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
              />
            </div>
            <Button
              onClick={handleSave}
              disabled={updateWorkspace.isPending}
            >
              {updateWorkspace.isPending ? "Saving..." : "Save Changes"}
            </Button>
            {updateWorkspace.isError && (
              <p className="text-sm text-destructive">
                {updateWorkspace.error instanceof Error
                  ? updateWorkspace.error.message
                  : "Failed to update workspace"}
              </p>
            )}
            {updateWorkspace.isSuccess && (
              <p className="text-sm text-green-600">Saved.</p>
            )}
          </CardContent>
        </Card>
      )}

      {isOwner() && (
        <>
          <Separator />
          <Card className="border-destructive/50">
            <CardHeader>
              <CardTitle className="text-destructive">Danger Zone</CardTitle>
              <CardDescription>
                Irreversible actions. Proceed with caution.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-3">
                <h4 className="text-sm font-medium">Transfer Ownership</h4>
                <p className="text-xs text-muted-foreground">
                  Transfer this workspace to another member. You will become an admin.
                </p>
                {otherMembers.length > 0 ? (
                  <div className="flex items-center gap-2">
                    <Select value={transferTarget} onValueChange={setTransferTarget}>
                      <SelectTrigger className="w-64">
                        <SelectValue placeholder="Select a member" />
                      </SelectTrigger>
                      <SelectContent>
                        {otherMembers.map((m) => (
                          <SelectItem key={m.user_id} value={m.user_id}>
                            {m.name || m.email}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button
                      variant="outline"
                      onClick={handleTransfer}
                      disabled={!transferTarget || transferOwnership.isPending}
                    >
                      {transferOwnership.isPending ? "Transferring..." : "Transfer"}
                    </Button>
                  </div>
                ) : (
                  <p className="text-xs text-muted-foreground">
                    Add members before transferring ownership.
                  </p>
                )}
                {transferOwnership.isError && (
                  <p className="text-sm text-destructive">
                    {transferOwnership.error instanceof Error
                      ? transferOwnership.error.message
                      : "Failed to transfer ownership"}
                  </p>
                )}
              </div>

              <Separator />

              <div className="space-y-3">
                <h4 className="text-sm font-medium">Delete Workspace</h4>
                <p className="text-xs text-muted-foreground">
                  Permanently delete this workspace and all its data.
                </p>
                <Button
                  variant="destructive"
                  onClick={handleDelete}
                  disabled={deleteWorkspace.isPending}
                >
                  {deleteWorkspace.isPending ? "Deleting..." : "Delete Workspace"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
