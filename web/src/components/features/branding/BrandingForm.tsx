import { useState, useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import type { WorkspaceBranding, UpdateBrandingInput } from "@/types/branding"

interface BrandingFormProps {
  branding?: WorkspaceBranding | null
  onSubmit: (input: UpdateBrandingInput) => void
  isLoading: boolean
}

export default function BrandingForm({
  branding,
  onSubmit,
  isLoading,
}: BrandingFormProps) {
  const [logoUrl, setLogoUrl] = useState("")
  const [logoDarkUrl, setLogoDarkUrl] = useState("")
  const [faviconUrl, setFaviconUrl] = useState("")
  const [primaryColor, setPrimaryColor] = useState("")
  const [secondaryColor, setSecondaryColor] = useState("")
  const [accentColor, setAccentColor] = useState("")
  const [customCss, setCustomCss] = useState("")
  const [hidePoweredBy, setHidePoweredBy] = useState(false)
  const [customFooterText, setCustomFooterText] = useState("")
  const [customFooterUrl, setCustomFooterUrl] = useState("")

  useEffect(() => {
    if (branding) {
      setLogoUrl(branding.logo_url ?? "")
      setLogoDarkUrl(branding.logo_dark_url ?? "")
      setFaviconUrl(branding.favicon_url ?? "")
      setPrimaryColor(branding.primary_color ?? "")
      setSecondaryColor(branding.secondary_color ?? "")
      setAccentColor(branding.accent_color ?? "")
      setCustomCss(branding.custom_css ?? "")
      setHidePoweredBy(branding.hide_powered_by)
      setCustomFooterText(branding.custom_footer_text ?? "")
      setCustomFooterUrl(branding.custom_footer_url ?? "")
    }
  }, [branding])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      logo_url: logoUrl || null,
      logo_dark_url: logoDarkUrl || null,
      favicon_url: faviconUrl || null,
      primary_color: primaryColor || null,
      secondary_color: secondaryColor || null,
      accent_color: accentColor || null,
      custom_css: customCss || null,
      hide_powered_by: hidePoweredBy,
      custom_footer_text: customFooterText || null,
      custom_footer_url: customFooterUrl || null,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid gap-6 sm:grid-cols-2">
        <div className="space-y-4">
          <h3 className="text-sm font-medium">Logos</h3>
          <div>
            <Label htmlFor="logo-url">Logo URL</Label>
            <Input
              id="logo-url"
              placeholder="https://..."
              value={logoUrl}
              onChange={(e) => setLogoUrl(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="logo-dark-url">Logo URL (Dark Mode)</Label>
            <Input
              id="logo-dark-url"
              placeholder="https://..."
              value={logoDarkUrl}
              onChange={(e) => setLogoDarkUrl(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="favicon-url">Favicon URL</Label>
            <Input
              id="favicon-url"
              placeholder="https://..."
              value={faviconUrl}
              onChange={(e) => setFaviconUrl(e.target.value)}
            />
          </div>
        </div>

        <div className="space-y-4">
          <h3 className="text-sm font-medium">Colors</h3>
          <div>
            <Label htmlFor="primary-color">Primary Color</Label>
            <div className="flex gap-2">
              <Input
                id="primary-color"
                placeholder="#000000"
                value={primaryColor}
                onChange={(e) => setPrimaryColor(e.target.value)}
              />
              {primaryColor && (
                <div
                  className="h-9 w-9 shrink-0 rounded border"
                  style={{ backgroundColor: primaryColor }}
                />
              )}
            </div>
          </div>
          <div>
            <Label htmlFor="secondary-color">Secondary Color</Label>
            <div className="flex gap-2">
              <Input
                id="secondary-color"
                placeholder="#000000"
                value={secondaryColor}
                onChange={(e) => setSecondaryColor(e.target.value)}
              />
              {secondaryColor && (
                <div
                  className="h-9 w-9 shrink-0 rounded border"
                  style={{ backgroundColor: secondaryColor }}
                />
              )}
            </div>
          </div>
          <div>
            <Label htmlFor="accent-color">Accent Color</Label>
            <div className="flex gap-2">
              <Input
                id="accent-color"
                placeholder="#000000"
                value={accentColor}
                onChange={(e) => setAccentColor(e.target.value)}
              />
              {accentColor && (
                <div
                  className="h-9 w-9 shrink-0 rounded border"
                  style={{ backgroundColor: accentColor }}
                />
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-medium">Custom CSS</h3>
        <textarea
          className="w-full rounded-md border bg-background p-3 font-mono text-sm"
          rows={6}
          placeholder=".bio-page { /* your styles */ }"
          value={customCss}
          onChange={(e) => setCustomCss(e.target.value)}
        />
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-medium">Footer</h3>
        <div className="flex items-center gap-3">
          <Switch
            checked={hidePoweredBy}
            onCheckedChange={setHidePoweredBy}
          />
          <Label>Hide "Powered by Linkrift" badge</Label>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <Label htmlFor="footer-text">Custom Footer Text</Label>
            <Input
              id="footer-text"
              placeholder="Powered by Your Brand"
              value={customFooterText}
              onChange={(e) => setCustomFooterText(e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="footer-url">Custom Footer URL</Label>
            <Input
              id="footer-url"
              placeholder="https://yourbrand.com"
              value={customFooterUrl}
              onChange={(e) => setCustomFooterUrl(e.target.value)}
            />
          </div>
        </div>
      </div>

      <Button type="submit" disabled={isLoading}>
        {isLoading ? "Saving..." : "Save Branding"}
      </Button>
    </form>
  )
}
