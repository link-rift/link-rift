import type { WorkspaceBranding } from "@/types/branding"

interface BrandingPreviewProps {
  branding?: WorkspaceBranding | null
}

export default function BrandingPreview({ branding }: BrandingPreviewProps) {
  if (!branding) {
    return (
      <div className="rounded-lg border bg-muted/50 p-8 text-center text-sm text-muted-foreground">
        No branding configured. Save branding settings to see a preview.
      </div>
    )
  }

  return (
    <div className="rounded-lg border p-4">
      <h3 className="mb-3 text-sm font-medium">Preview</h3>
      <div
        className="rounded-md border p-6"
        style={{
          backgroundColor: branding.primary_color || undefined,
        }}
      >
        {branding.logo_url && (
          <div className="mb-4 flex justify-center">
            <img
              src={branding.logo_url}
              alt="Logo preview"
              className="h-10 max-w-[200px] object-contain"
            />
          </div>
        )}

        <div className="space-y-2 text-center">
          <div
            className="text-lg font-bold"
            style={{ color: branding.secondary_color || undefined }}
          >
            Sample Page Title
          </div>
          <div
            className="text-sm opacity-70"
            style={{ color: branding.secondary_color || undefined }}
          >
            This is how your branding will look on public pages.
          </div>
        </div>

        <div className="mt-4 flex justify-center">
          <div
            className="rounded-md px-6 py-2 text-sm font-medium text-white"
            style={{
              backgroundColor: branding.accent_color || "#1a1a1a",
            }}
          >
            Sample Button
          </div>
        </div>

        <div className="mt-6 text-center text-xs opacity-50">
          {branding.hide_powered_by
            ? branding.custom_footer_text || "(Powered by badge hidden)"
            : "Powered by Linkrift"}
        </div>
      </div>
    </div>
  )
}
