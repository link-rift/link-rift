import { useNavigate } from "react-router-dom"
import { useDeleteBioPage } from "@/hooks/useBioPages"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import type { BioPage } from "@/types/biopage"

interface BioPageListProps {
  pages: BioPage[]
  isLoading: boolean
}

export default function BioPageList({ pages, isLoading }: BioPageListProps) {
  const navigate = useNavigate()

  if (isLoading) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i}>
            <CardHeader><Skeleton className="h-5 w-32" /></CardHeader>
            <CardContent><Skeleton className="h-4 w-48" /></CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (pages.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-12 text-center">
        <h3 className="text-lg font-medium">No bio pages yet</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Create your first link-in-bio page to get started.
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {pages.map((page) => (
        <BioPageCard key={page.id} page={page} onClick={() => navigate(`/bio-pages/${page.id}`)} />
      ))}
    </div>
  )
}

function BioPageCard({ page, onClick }: { page: BioPage; onClick: () => void }) {
  const deleteMutation = useDeleteBioPage()

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (window.confirm(`Delete "${page.title}"? This cannot be undone.`)) {
      deleteMutation.mutate(page.id)
    }
  }

  return (
    <Card className="cursor-pointer transition-shadow hover:shadow-md" onClick={onClick}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium">{page.title}</CardTitle>
        <Badge variant={page.is_published ? "default" : "secondary"} className="text-xs">
          {page.is_published ? "Published" : "Draft"}
        </Badge>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">/b/{page.slug}</p>
        <div className="mt-2 flex items-center justify-between">
          <span className="text-xs text-muted-foreground">
            {page.link_count ?? 0} links
          </span>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="h-7 px-2 text-xs text-destructive hover:text-destructive"
          >
            Delete
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
