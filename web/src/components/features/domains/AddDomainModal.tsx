import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useCreateDomain } from "@/hooks/useDomains"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"

const schema = z.object({
  domain: z
    .string()
    .min(1, "Domain is required")
    .regex(
      /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/i,
      "Enter a valid domain (e.g., links.example.com)"
    ),
})

type FormValues = z.infer<typeof schema>

interface AddDomainModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function AddDomainModal({
  open,
  onOpenChange,
}: AddDomainModalProps) {
  const createDomain = useCreateDomain()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { domain: "" },
  })

  const onSubmit = async (data: FormValues) => {
    try {
      await createDomain.mutateAsync({ domain: data.domain.toLowerCase() })
      reset()
      onOpenChange(false)
    } catch {
      // Error is handled by mutation state
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) reset()
        onOpenChange(isOpen)
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Custom Domain</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="domain">Domain</Label>
            <Input
              id="domain"
              placeholder="links.example.com"
              {...register("domain")}
            />
            {errors.domain && (
              <p className="text-sm text-destructive">{errors.domain.message}</p>
            )}
            {createDomain.isError && (
              <p className="text-sm text-destructive">
                {createDomain.error instanceof Error
                  ? createDomain.error.message
                  : "Failed to add domain"}
              </p>
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            After adding your domain, you will need to configure DNS records to
            verify ownership.
          </p>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createDomain.isPending}>
              {createDomain.isPending ? "Adding..." : "Add Domain"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
