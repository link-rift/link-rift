import { useState } from "react"
import { useBioPages } from "@/hooks/useBioPages"
import { FeatureGate } from "@/components/ui/FeatureGate"
import { Button } from "@/components/ui/button"
import BioPageList from "@/components/features/biopages/BioPageList"
import CreateBioPageModal from "@/components/features/biopages/CreateBioPageModal"

export default function BioPagesPage() {
  return (
    <FeatureGate feature="bio_pages">
      <BioPagesContent />
    </FeatureGate>
  )
}

function BioPagesContent() {
  const [createOpen, setCreateOpen] = useState(false)
  const { data: pages, isLoading } = useBioPages()

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Bio Pages</h1>
          <p className="text-sm text-muted-foreground">
            Create and manage link-in-bio pages
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <svg className="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Create Bio Page
        </Button>
      </div>
      <BioPageList pages={pages ?? []} isLoading={isLoading} />
      <CreateBioPageModal open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  )
}
