import { useParams, useNavigate } from "react-router-dom"
import { useBioPage, useBioPageLinks, usePublishBioPage, useUnpublishBioPage } from "@/hooks/useBioPages"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import BioPageEditor from "@/components/features/biopages/BioPageEditor"
import BioPagePreview from "@/components/features/biopages/BioPagePreview"

export default function BioPageEditorPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: page, isLoading: pageLoading } = useBioPage(id ?? "")
  const { data: links, isLoading: linksLoading } = useBioPageLinks(id ?? "")
  const publishMutation = usePublishBioPage()
  const unpublishMutation = useUnpublishBioPage()

  if (pageLoading || !page) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  const handleTogglePublish = () => {
    if (page.is_published) {
      unpublishMutation.mutate(page.id)
    } else {
      publishMutation.mutate(page.id)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => navigate("/bio-pages")}>
            <svg className="mr-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back
          </Button>
          <h1 className="text-xl font-bold">{page.title}</h1>
          <Badge variant={page.is_published ? "default" : "secondary"}>
            {page.is_published ? "Published" : "Draft"}
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          {page.is_published && (
            <Button variant="outline" size="sm" asChild>
              <a href={`/b/${page.slug}`} target="_blank" rel="noopener noreferrer">
                View Page
              </a>
            </Button>
          )}
          <Button
            size="sm"
            variant={page.is_published ? "outline" : "default"}
            onClick={handleTogglePublish}
            disabled={publishMutation.isPending || unpublishMutation.isPending}
          >
            {page.is_published ? "Unpublish" : "Publish"}
          </Button>
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <BioPageEditor page={page} links={links ?? []} isLoading={linksLoading} />
        <BioPagePreview page={page} links={links ?? []} />
      </div>
    </div>
  )
}
