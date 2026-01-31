import { useState } from "react"
import { useCreateLink } from "@/hooks/useLinks"
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
import type { CreateLinkRequest } from "@/types/link"

interface CreateLinkModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function CreateLinkModal({ open, onOpenChange }: CreateLinkModalProps) {
  const createLink = useCreateLink()
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [form, setForm] = useState<CreateLinkRequest>({
    url: "",
  })

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    const data: CreateLinkRequest = { url: form.url }
    if (form.short_code) data.short_code = form.short_code
    if (form.title) data.title = form.title
    if (form.description) data.description = form.description
    if (form.password) data.password = form.password
    if (form.expires_at) data.expires_at = form.expires_at
    if (form.max_clicks) data.max_clicks = form.max_clicks
    if (form.utm_source) data.utm_source = form.utm_source
    if (form.utm_medium) data.utm_medium = form.utm_medium
    if (form.utm_campaign) data.utm_campaign = form.utm_campaign
    if (form.utm_term) data.utm_term = form.utm_term
    if (form.utm_content) data.utm_content = form.utm_content

    createLink.mutate(data, {
      onSuccess: () => {
        onOpenChange(false)
        setForm({ url: "" })
        setShowAdvanced(false)
      },
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Create New Link</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="url">Destination URL</Label>
            <Input
              id="url"
              type="url"
              placeholder="https://example.com/long-url"
              value={form.url}
              onChange={(e) => setForm({ ...form, url: e.target.value })}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="short_code">Custom Short Code (optional)</Label>
            <Input
              id="short_code"
              placeholder="my-link"
              value={form.short_code || ""}
              onChange={(e) => setForm({ ...form, short_code: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="title">Title (optional)</Label>
            <Input
              id="title"
              placeholder="Link title"
              value={form.title || ""}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </div>

          <div className="flex items-center gap-2">
            <Switch
              checked={showAdvanced}
              onCheckedChange={setShowAdvanced}
              id="advanced"
            />
            <Label htmlFor="advanced">Advanced options</Label>
          </div>

          {showAdvanced && (
            <div className="space-y-4 rounded-lg border p-4">
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Link description"
                  value={form.description || ""}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="password">Password Protection</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Enter password"
                  value={form.password || ""}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="expires_at">Expiration Date</Label>
                  <Input
                    id="expires_at"
                    type="datetime-local"
                    value={form.expires_at || ""}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        expires_at: e.target.value ? new Date(e.target.value).toISOString() : undefined,
                      })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="max_clicks">Max Clicks</Label>
                  <Input
                    id="max_clicks"
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
                <Label className="text-sm font-medium">UTM Parameters</Label>
                <div className="grid grid-cols-2 gap-2">
                  <Input
                    placeholder="utm_source"
                    value={form.utm_source || ""}
                    onChange={(e) => setForm({ ...form, utm_source: e.target.value })}
                  />
                  <Input
                    placeholder="utm_medium"
                    value={form.utm_medium || ""}
                    onChange={(e) => setForm({ ...form, utm_medium: e.target.value })}
                  />
                  <Input
                    placeholder="utm_campaign"
                    value={form.utm_campaign || ""}
                    onChange={(e) => setForm({ ...form, utm_campaign: e.target.value })}
                  />
                  <Input
                    placeholder="utm_term"
                    value={form.utm_term || ""}
                    onChange={(e) => setForm({ ...form, utm_term: e.target.value })}
                  />
                  <Input
                    placeholder="utm_content"
                    value={form.utm_content || ""}
                    onChange={(e) => setForm({ ...form, utm_content: e.target.value })}
                    className="col-span-2"
                  />
                </div>
              </div>
            </div>
          )}

          {createLink.isError && (
            <p className="text-sm text-destructive">
              {createLink.error instanceof Error ? createLink.error.message : "Failed to create link"}
            </p>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createLink.isPending || !form.url}>
              {createLink.isPending ? "Creating..." : "Create Link"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
