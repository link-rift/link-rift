import { useState, useEffect } from "react"
import { useUpdateLink } from "@/hooks/useLinks"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import type { Link, UpdateLinkRequest } from "@/types/link"

interface EditLinkModalProps {
  link: Link | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function EditLinkModal({ link, open, onOpenChange }: EditLinkModalProps) {
  const updateLink = useUpdateLink()
  const [form, setForm] = useState<UpdateLinkRequest>({})

  useEffect(() => {
    if (link) {
      setForm({
        url: link.url,
        title: link.title || "",
        description: link.description || "",
        is_active: link.is_active,
      })
    }
  }, [link])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!link) return

    updateLink.mutate(
      { id: link.id, data: form },
      {
        onSuccess: () => {
          onOpenChange(false)
        },
      }
    )
  }

  if (!link) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Edit Link</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-url">Destination URL</Label>
            <Input
              id="edit-url"
              type="url"
              value={form.url || ""}
              onChange={(e) => setForm({ ...form, url: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-title">Title</Label>
            <Input
              id="edit-title"
              value={form.title || ""}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-description">Description</Label>
            <Textarea
              id="edit-description"
              value={form.description || ""}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>

          <div className="flex items-center gap-2">
            <Switch
              checked={form.is_active ?? true}
              onCheckedChange={(checked) => setForm({ ...form, is_active: checked })}
              id="edit-active"
            />
            <Label htmlFor="edit-active">Active</Label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-expires">Expiration Date</Label>
              <Input
                id="edit-expires"
                type="datetime-local"
                value={form.expires_at || ""}
                onChange={(e) =>
                  setForm({
                    ...form,
                    expires_at: e.target.value ? new Date(e.target.value).toISOString() : "",
                  })
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-max-clicks">Max Clicks</Label>
              <Input
                id="edit-max-clicks"
                type="number"
                min="1"
                placeholder="Unlimited"
                value={form.max_clicks || ""}
                onChange={(e) =>
                  setForm({
                    ...form,
                    max_clicks: e.target.value ? parseInt(e.target.value) : undefined,
                  })
                }
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-password">New Password (leave empty to keep current)</Label>
            <Input
              id="edit-password"
              type="password"
              placeholder="Enter new password"
              value={form.password || ""}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
            />
          </div>

          {updateLink.isError && (
            <p className="text-sm text-destructive">
              {updateLink.error instanceof Error ? updateLink.error.message : "Failed to update link"}
            </p>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateLink.isPending}>
              {updateLink.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
