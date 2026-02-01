import { useBranding, useUpdateBranding, useDeleteBranding } from "@/hooks/useBranding"
import { FeatureGate } from "@/components/ui/FeatureGate"
import BrandingForm from "@/components/features/branding/BrandingForm"
import BrandingPreview from "@/components/features/branding/BrandingPreview"
import { Button } from "@/components/ui/button"
import type { UpdateBrandingInput } from "@/types/branding"

export default function BrandingPage() {
  return (
    <FeatureGate feature="white_label" tier="enterprise">
      <BrandingContent />
    </FeatureGate>
  )
}

function BrandingContent() {
  const { data: branding, isLoading } = useBranding()
  const updateMutation = useUpdateBranding()
  const deleteMutation = useDeleteBranding()

  const handleSubmit = (input: UpdateBrandingInput) => {
    updateMutation.mutate(input)
  }

  const handleReset = () => {
    if (confirm("Are you sure you want to reset all branding to defaults?")) {
      deleteMutation.mutate()
    }
  }

  if (isLoading) {
    return (
      <div className="py-12 text-center text-muted-foreground">
        Loading branding settings...
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Branding</h2>
          <p className="text-sm text-muted-foreground">
            Customize the appearance of your public pages
          </p>
        </div>
        {branding && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleReset}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? "Resetting..." : "Reset to Defaults"}
          </Button>
        )}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <BrandingForm
            branding={branding}
            onSubmit={handleSubmit}
            isLoading={updateMutation.isPending}
          />
        </div>
        <div>
          <BrandingPreview branding={branding} />
        </div>
      </div>
    </div>
  )
}
