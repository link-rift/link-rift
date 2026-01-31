import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useCreateBioPage } from "@/hooks/useBioPages"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"

const schema = z.object({
  title: z.string().min(1, "Title is required").max(255),
  slug: z
    .string()
    .min(1, "Slug is required")
    .max(100)
    .regex(/^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$/, "Slug must be lowercase alphanumeric with hyphens"),
  bio: z.string().max(500).optional(),
})

type FormValues = z.infer<typeof schema>

interface CreateBioPageModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function CreateBioPageModal({ open, onOpenChange }: CreateBioPageModalProps) {
  const navigate = useNavigate()
  const createBioPage = useCreateBioPage()
  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: "", slug: "", bio: "" },
  })

  const title = watch("title")

  useEffect(() => {
    const slug = title
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "")
    setValue("slug", slug)
  }, [title, setValue])

  const onSubmit = async (data: FormValues) => {
    try {
      const page = await createBioPage.mutateAsync({
        title: data.title,
        slug: data.slug,
        bio: data.bio || undefined,
      })
      reset()
      onOpenChange(false)
      navigate(`/bio-pages/${page.id}`)
    } catch {
      // Error handled by mutation state
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
          <DialogTitle>Create Bio Page</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="title">Title</Label>
            <Input id="title" placeholder="My Bio Page" {...register("title")} />
            {errors.title && (
              <p className="text-xs text-destructive">{errors.title.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="slug">Slug</Label>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">/b/</span>
              <Input id="slug" placeholder="my-bio-page" {...register("slug")} />
            </div>
            {errors.slug && (
              <p className="text-xs text-destructive">{errors.slug.message}</p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="bio">Bio</Label>
            <Textarea
              id="bio"
              placeholder="A short description about you..."
              rows={3}
              {...register("bio")}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createBioPage.isPending}>
              {createBioPage.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>
          {createBioPage.isError && (
            <p className="text-xs text-destructive">
              {createBioPage.error?.message || "Failed to create bio page"}
            </p>
          )}
        </form>
      </DialogContent>
    </Dialog>
  )
}
