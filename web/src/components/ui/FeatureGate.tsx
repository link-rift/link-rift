import * as React from "react"
import { useLicenseStore } from "@/stores/licenseStore"
import { UpgradePrompt } from "@/components/ui/UpgradePrompt"
import type { Feature, Tier } from "@/types/license"
import { FEATURE_DISPLAY_NAMES } from "@/types/license"

interface FeatureGateProps {
  feature?: Feature
  tier?: Tier
  fallback?: React.ReactNode
  upgradeVariant?: "inline" | "card" | "banner"
  children: React.ReactNode
}

export function FeatureGate({
  feature,
  tier,
  fallback,
  upgradeVariant = "card",
  children,
}: FeatureGateProps) {
  const { hasFeature, hasTier } = useLicenseStore()

  const hasAccess = React.useMemo(() => {
    if (feature && !hasFeature(feature)) return false
    if (tier && !hasTier(tier)) return false
    return true
  }, [feature, tier, hasFeature, hasTier])

  if (hasAccess) {
    return <>{children}</>
  }

  if (fallback) {
    return <>{fallback}</>
  }

  const displayName = feature ? FEATURE_DISPLAY_NAMES[feature] : undefined
  const requiredTier = tier ?? "pro"

  return (
    <UpgradePrompt
      feature={displayName}
      requiredTier={requiredTier}
      variant={upgradeVariant}
    />
  )
}
