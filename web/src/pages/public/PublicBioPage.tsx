import { useParams } from "react-router-dom"
import { usePublicBioPage } from "@/hooks/useBioPages"
import { trackBioLinkClick } from "@/services/biopages"
import type { PublicBioPage as PublicBioPageType, ThemeStyles } from "@/types/biopage"

export default function PublicBioPage() {
  const { slug } = useParams<{ slug: string }>()
  const { data: page, isLoading, error } = usePublicBioPage(slug ?? "")

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (error || !page) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-center">
          <h1 className="text-2xl font-bold">Page Not Found</h1>
          <p className="mt-2 text-muted-foreground">This bio page doesn't exist or isn't published.</p>
        </div>
      </div>
    )
  }

  return <BioPageRenderer page={page} />
}

function BioPageRenderer({ page }: { page: PublicBioPageType }) {
  const styles = page.theme?.styles
  const bgStyle = getBackgroundStyle(styles)

  const handleLinkClick = (linkId: string, url: string) => {
    trackBioLinkClick(page.slug, linkId)
    window.open(url, "_blank", "noopener,noreferrer")
  }

  return (
    <div className="flex min-h-screen flex-col items-center" style={bgStyle}>
      {/* Custom CSS is user-provided content for their own bio page styling */}
      {page.custom_css && <style dangerouslySetInnerHTML={{ __html: page.custom_css }} />}
      <div className="w-full max-w-md px-4 py-12">
        {page.avatar_url && (
          <div className="mb-4 flex justify-center">
            <img
              src={page.avatar_url}
              alt={page.title}
              className="h-24 w-24 rounded-full object-cover"
            />
          </div>
        )}
        <h1
          className="mb-2 text-center text-2xl font-bold"
          style={{ color: styles?.text_color }}
        >
          {page.title}
        </h1>
        {page.bio && (
          <p
            className="mb-8 text-center text-sm opacity-80"
            style={{ color: styles?.text_color }}
          >
            {page.bio}
          </p>
        )}
        <div className="space-y-3">
          {page.links.map((link) => (
            <button
              key={link.id}
              onClick={() => handleLinkClick(link.id, link.url)}
              className="w-full px-4 py-3 text-center font-medium transition-transform hover:scale-[1.02] active:scale-[0.98]"
              style={{
                backgroundColor: styles?.button_color || "#1a1a1a",
                color: styles?.button_text_color || "#ffffff",
                borderRadius: styles?.button_style === "pill" ? "9999px" : "0.5rem",
                fontFamily: styles?.font_family,
              }}
            >
              {link.icon && <span className="mr-2">{link.icon}</span>}
              {link.title}
            </button>
          ))}
        </div>
        <div className="mt-12 text-center">
          <span className="text-xs opacity-50" style={{ color: styles?.text_color }}>
            Powered by Linkrift
          </span>
        </div>
      </div>
    </div>
  )
}

function getBackgroundStyle(styles?: ThemeStyles | null): React.CSSProperties {
  if (!styles) return { backgroundColor: "#ffffff" }
  if (styles.gradient) {
    return {
      background: `linear-gradient(${styles.gradient.direction}, ${styles.gradient.from}, ${styles.gradient.to})`,
    }
  }
  return { backgroundColor: styles.background_color }
}
