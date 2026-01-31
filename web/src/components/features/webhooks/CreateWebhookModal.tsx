import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useCreateWebhook } from "@/hooks/useWebhooks"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { WEBHOOK_EVENTS } from "@/types/webhook"

const schema = z.object({
  url: z
    .string()
    .min(1, "URL is required")
    .url("Enter a valid URL")
    .refine((url) => url.startsWith("https://"), "URL must use HTTPS"),
  events: z.array(z.string()).min(1, "Select at least one event"),
})

type FormValues = z.infer<typeof schema>

interface CreateWebhookModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function CreateWebhookModal({
  open,
  onOpenChange,
}: CreateWebhookModalProps) {
  const createWebhook = useCreateWebhook()
  const [createdSecret, setCreatedSecret] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { url: "", events: [] },
  })

  const selectedEvents = watch("events")

  const onSubmit = async (data: FormValues) => {
    try {
      const result = await createWebhook.mutateAsync({
        url: data.url,
        events: data.events,
      })
      setCreatedSecret(result.secret)
    } catch {
      // Error handled by mutation state
    }
  }

  const handleClose = () => {
    reset()
    setCreatedSecret(null)
    setCopied(false)
    onOpenChange(false)
  }

  const handleCopy = async () => {
    if (createdSecret) {
      await navigator.clipboard.writeText(createdSecret)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const toggleEvent = (event: string) => {
    const current = selectedEvents || []
    if (current.includes(event)) {
      setValue("events", current.filter((e) => e !== event), { shouldValidate: true })
    } else {
      setValue("events", [...current, event], { shouldValidate: true })
    }
  }

  // Group events by category
  const categories = WEBHOOK_EVENTS.reduce(
    (acc, event) => {
      if (!acc[event.category]) acc[event.category] = []
      acc[event.category].push(event)
      return acc
    },
    {} as Record<string, typeof WEBHOOK_EVENTS[number][]>
  )

  if (createdSecret) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Webhook Created</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Copy your webhook signing secret now. You won't be able to see it again.
            </p>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted p-3 text-xs break-all">
                {createdSecret}
              </code>
              <Button size="sm" variant="outline" onClick={handleCopy}>
                {copied ? "Copied!" : "Copy"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Use this secret to verify webhook signatures using HMAC-SHA256.
              The signature is sent in the <code className="text-xs">X-Linkrift-Signature</code> header.
            </p>
          </div>
          <DialogFooter>
            <Button onClick={handleClose}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Create Webhook</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="url">Endpoint URL</Label>
            <Input
              id="url"
              placeholder="https://api.example.com/webhooks"
              {...register("url")}
            />
            {errors.url && (
              <p className="text-sm text-destructive">{errors.url.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Events</Label>
            <div className="rounded-md border p-3 max-h-60 overflow-y-auto space-y-4">
              {Object.entries(categories).map(([category, events]) => (
                <div key={category}>
                  <p className="mb-2 text-xs font-semibold uppercase text-muted-foreground">
                    {category}
                  </p>
                  <div className="grid gap-2">
                    {events.map((event) => (
                      <label
                        key={event.value}
                        className="flex items-center gap-2 cursor-pointer"
                      >
                        <Checkbox
                          checked={selectedEvents?.includes(event.value) ?? false}
                          onCheckedChange={() => toggleEvent(event.value)}
                        />
                        <span className="text-sm">{event.label}</span>
                        <code className="text-xs text-muted-foreground">
                          {event.value}
                        </code>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
            {errors.events && (
              <p className="text-sm text-destructive">{errors.events.message}</p>
            )}
          </div>

          {createWebhook.isError && (
            <p className="text-sm text-destructive">
              {createWebhook.error instanceof Error
                ? createWebhook.error.message
                : "Failed to create webhook"}
            </p>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={createWebhook.isPending}>
              {createWebhook.isPending ? "Creating..." : "Create Webhook"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
