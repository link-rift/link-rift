import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useCreateAPIKey } from "@/hooks/useAPIKeys"
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
import { API_KEY_SCOPES } from "@/types/apikey"

const schema = z.object({
  name: z.string().min(1, "Name is required").max(100),
  scopes: z.array(z.string()).min(1, "Select at least one scope"),
  expires_at: z.string().optional(),
})

type FormValues = z.infer<typeof schema>

interface CreateAPIKeyModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function CreateAPIKeyModal({
  open,
  onOpenChange,
}: CreateAPIKeyModalProps) {
  const createKey = useCreateAPIKey()
  const [createdKey, setCreatedKey] = useState<string | null>(null)
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
    defaultValues: { name: "", scopes: [], expires_at: "" },
  })

  const selectedScopes = watch("scopes")

  const onSubmit = async (data: FormValues) => {
    try {
      const result = await createKey.mutateAsync({
        name: data.name,
        scopes: data.scopes,
        expires_at: data.expires_at || undefined,
      })
      setCreatedKey(result.key)
    } catch {
      // Error handled by mutation state
    }
  }

  const handleClose = () => {
    reset()
    setCreatedKey(null)
    setCopied(false)
    onOpenChange(false)
  }

  const handleCopy = async () => {
    if (createdKey) {
      await navigator.clipboard.writeText(createdKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const toggleScope = (scope: string) => {
    const current = selectedScopes || []
    if (current.includes(scope)) {
      setValue("scopes", current.filter((s) => s !== scope), { shouldValidate: true })
    } else {
      setValue("scopes", [...current, scope], { shouldValidate: true })
    }
  }

  if (createdKey) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>API Key Created</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Copy your API key now. You won't be able to see it again.
            </p>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted p-3 text-xs break-all">
                {createdKey}
              </code>
              <Button size="sm" variant="outline" onClick={handleCopy}>
                {copied ? "Copied!" : "Copy"}
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Use this key in the <code className="text-xs">X-API-Key</code> header for API requests.
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
          <DialogTitle>Create API Key</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              placeholder="e.g., Production API Key"
              {...register("name")}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label>Scopes</Label>
            <div className="grid gap-2 rounded-md border p-3 max-h-48 overflow-y-auto">
              {API_KEY_SCOPES.map((scope) => (
                <label
                  key={scope.value}
                  className="flex items-center gap-2 cursor-pointer"
                >
                  <Checkbox
                    checked={selectedScopes?.includes(scope.value) ?? false}
                    onCheckedChange={() => toggleScope(scope.value)}
                  />
                  <span className="text-sm font-medium">{scope.label}</span>
                  <span className="text-xs text-muted-foreground">
                    - {scope.description}
                  </span>
                </label>
              ))}
            </div>
            {errors.scopes && (
              <p className="text-sm text-destructive">{errors.scopes.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="expires_at">Expiration (Optional)</Label>
            <Input
              id="expires_at"
              type="datetime-local"
              {...register("expires_at")}
            />
          </div>

          {createKey.isError && (
            <p className="text-sm text-destructive">
              {createKey.error instanceof Error
                ? createKey.error.message
                : "Failed to create API key"}
            </p>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={createKey.isPending}>
              {createKey.isPending ? "Creating..." : "Create Key"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
