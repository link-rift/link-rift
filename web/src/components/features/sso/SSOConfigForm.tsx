import { useState, useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import type { SSOConfig, CreateSSOConfigInput } from "@/types/sso"

interface SSOConfigFormProps {
  config?: SSOConfig | null
  onSubmit: (input: CreateSSOConfigInput) => void
  isLoading: boolean
}

export default function SSOConfigForm({
  config,
  onSubmit,
  isLoading,
}: SSOConfigFormProps) {
  const [entityId, setEntityId] = useState("")
  const [ssoUrl, setSsoUrl] = useState("")
  const [sloUrl, setSloUrl] = useState("")
  const [certificate, setCertificate] = useState("")
  const [idpMetadataUrl, setIdpMetadataUrl] = useState("")
  const [isEnabled, setIsEnabled] = useState(false)
  const [enforceSso, setEnforceSso] = useState(false)
  const [allowedDomains, setAllowedDomains] = useState("")

  useEffect(() => {
    if (config) {
      setEntityId(config.entity_id)
      setSsoUrl(config.sso_url)
      setSloUrl(config.slo_url ?? "")
      setCertificate(config.certificate)
      setIdpMetadataUrl(config.idp_metadata_url ?? "")
      setIsEnabled(config.is_enabled)
      setEnforceSso(config.enforce_sso)
      setAllowedDomains(config.allowed_domains?.join(", ") ?? "")
    }
  }, [config])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      provider: "saml",
      entity_id: entityId,
      sso_url: ssoUrl,
      slo_url: sloUrl || undefined,
      certificate,
      idp_metadata_url: idpMetadataUrl || undefined,
      is_enabled: isEnabled,
      enforce_sso: enforceSso,
      allowed_domains: allowedDomains
        ? allowedDomains.split(",").map((d) => d.trim()).filter(Boolean)
        : [],
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <h3 className="text-sm font-medium">Identity Provider Settings</h3>
        <div>
          <Label htmlFor="entity-id">IdP Entity ID</Label>
          <Input
            id="entity-id"
            placeholder="https://idp.example.com/..."
            value={entityId}
            onChange={(e) => setEntityId(e.target.value)}
            required
          />
        </div>
        <div>
          <Label htmlFor="sso-url">SSO URL (Sign-in URL)</Label>
          <Input
            id="sso-url"
            placeholder="https://idp.example.com/sso/saml"
            value={ssoUrl}
            onChange={(e) => setSsoUrl(e.target.value)}
            required
          />
        </div>
        <div>
          <Label htmlFor="slo-url">SLO URL (Sign-out URL)</Label>
          <Input
            id="slo-url"
            placeholder="https://idp.example.com/slo/saml"
            value={sloUrl}
            onChange={(e) => setSloUrl(e.target.value)}
          />
        </div>
        <div>
          <Label htmlFor="idp-metadata">IdP Metadata URL</Label>
          <Input
            id="idp-metadata"
            placeholder="https://idp.example.com/metadata"
            value={idpMetadataUrl}
            onChange={(e) => setIdpMetadataUrl(e.target.value)}
          />
        </div>
        <div>
          <Label htmlFor="certificate">IdP Certificate (PEM or Base64)</Label>
          <textarea
            id="certificate"
            className="w-full rounded-md border bg-background p-3 font-mono text-xs"
            rows={6}
            placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
            value={certificate}
            onChange={(e) => setCertificate(e.target.value)}
            required
          />
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-sm font-medium">Access Control</h3>
        <div>
          <Label htmlFor="allowed-domains">
            Allowed Email Domains (comma-separated)
          </Label>
          <Input
            id="allowed-domains"
            placeholder="example.com, company.org"
            value={allowedDomains}
            onChange={(e) => setAllowedDomains(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-3">
          <Switch checked={isEnabled} onCheckedChange={setIsEnabled} />
          <Label>Enable SSO</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch checked={enforceSso} onCheckedChange={setEnforceSso} />
          <Label>Enforce SSO (require SSO for all workspace members)</Label>
        </div>
      </div>

      <Button type="submit" disabled={isLoading}>
        {isLoading ? "Saving..." : config ? "Update SSO Config" : "Configure SSO"}
      </Button>
    </form>
  )
}
