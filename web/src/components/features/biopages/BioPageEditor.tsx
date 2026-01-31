import { useState } from "react"
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core"
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable"
import {
  useAddBioPageLink,
  useDeleteBioPageLink,
  useReorderBioPageLinks,
  useUpdateBioPageLink,
} from "@/hooks/useBioPages"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import SortableLinkItem from "./SortableLinkItem"
import type { BioPage, BioPageLink } from "@/types/biopage"

interface BioPageEditorProps {
  page: BioPage
  links: BioPageLink[]
  isLoading: boolean
}

export default function BioPageEditor({ page, links, isLoading }: BioPageEditorProps) {
  const [addingLink, setAddingLink] = useState(false)
  const [newTitle, setNewTitle] = useState("")
  const [newUrl, setNewUrl] = useState("")
  const addLink = useAddBioPageLink()
  const deleteLink = useDeleteBioPageLink()
  const updateLink = useUpdateBioPageLink()
  const reorderLinks = useReorderBioPageLinks()

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  )

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = links.findIndex((l) => l.id === active.id)
    const newIndex = links.findIndex((l) => l.id === over.id)
    const reordered = arrayMove(links, oldIndex, newIndex)

    reorderLinks.mutate({
      pageId: page.id,
      data: { link_ids: reordered.map((l) => l.id) },
    })
  }

  const handleAddLink = async () => {
    if (!newTitle.trim() || !newUrl.trim()) return
    try {
      await addLink.mutateAsync({
        pageId: page.id,
        data: { title: newTitle.trim(), url: newUrl.trim() },
      })
      setNewTitle("")
      setNewUrl("")
      setAddingLink(false)
    } catch {
      // handled by mutation
    }
  }

  const handleDeleteLink = (linkId: string) => {
    if (window.confirm("Delete this link?")) {
      deleteLink.mutate({ pageId: page.id, linkId })
    }
  }

  const handleToggleVisibility = (link: BioPageLink) => {
    updateLink.mutate({
      pageId: page.id,
      linkId: link.id,
      data: { is_visible: !link.is_visible },
    })
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0">
        <CardTitle className="text-base">Links</CardTitle>
        <Button size="sm" onClick={() => setAddingLink(true)}>
          Add Link
        </Button>
      </CardHeader>
      <CardContent className="space-y-3">
        {addingLink && (
          <div className="space-y-2 rounded-lg border p-3">
            <Input
              placeholder="Link title"
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
            />
            <Input
              placeholder="https://example.com"
              value={newUrl}
              onChange={(e) => setNewUrl(e.target.value)}
            />
            <div className="flex gap-2">
              <Button size="sm" onClick={handleAddLink} disabled={addLink.isPending}>
                {addLink.isPending ? "Adding..." : "Add"}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setAddingLink(false)}>
                Cancel
              </Button>
            </div>
          </div>
        )}

        {isLoading ? (
          <div className="py-8 text-center text-sm text-muted-foreground">Loading links...</div>
        ) : links.length === 0 ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No links yet. Add your first link above.
          </div>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={links.map((l) => l.id)} strategy={verticalListSortingStrategy}>
              {links.map((link) => (
                <SortableLinkItem
                  key={link.id}
                  link={link}
                  onDelete={() => handleDeleteLink(link.id)}
                  onToggleVisibility={() => handleToggleVisibility(link)}
                />
              ))}
            </SortableContext>
          </DndContext>
        )}
      </CardContent>
    </Card>
  )
}
