import type { BioPage, BioPageLink } from "@/types/biopage"

interface BioPagePreviewProps {
  page: BioPage
  links: BioPageLink[]
}

export default function BioPagePreview({ page, links }: BioPagePreviewProps) {
  const visibleLinks = links.filter((l) => l.is_visible)

  return (
    <div className="flex justify-center">
      <div className="w-[375px]">
        <div className="rounded-[2rem] border-4 border-muted bg-background p-1 shadow-lg">
          <div className="rounded-[1.5rem] bg-white dark:bg-zinc-950 overflow-hidden" style={{ minHeight: 600 }}>
            <div className="flex flex-col items-center px-6 py-10">
              {page.avatar_url && (
                <img
                  src={page.avatar_url}
                  alt={page.title}
                  className="mb-4 h-20 w-20 rounded-full object-cover"
                />
              )}
              <h2 className="mb-1 text-lg font-bold">{page.title}</h2>
              {page.bio && (
                <p className="mb-6 text-center text-xs text-muted-foreground">{page.bio}</p>
              )}
              <div className="w-full space-y-2.5">
                {visibleLinks.map((link) => (
                  <div
                    key={link.id}
                    className="w-full rounded-lg bg-primary px-4 py-2.5 text-center text-sm font-medium text-primary-foreground"
                  >
                    {link.title}
                  </div>
                ))}
                {visibleLinks.length === 0 && (
                  <p className="py-8 text-center text-xs text-muted-foreground">
                    Add links to see them here
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
