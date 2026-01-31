import { create } from "zustand"
import type { Feature, LicenseInfo, Tier } from "@/types/license"

const tierLevels: Record<Tier, number> = {
  free: 1,
  pro: 2,
  business: 3,
  enterprise: 4,
}

const defaultCELicense: LicenseInfo = {
  type: "subscription",
  tier: "free",
  plan: {
    tier: "free",
    name: "Community",
    description: "Free self-hosted edition with core features",
    price: "Free",
  },
  features: ["link_expiration"],
  limits: {
    max_users: 1,
    max_domains: 0,
    max_links_per_month: 100,
    max_clicks_per_month: 10000,
    max_workspaces: 1,
    max_api_requests_per_min: 10,
  },
  is_community: true,
}

interface LicenseState {
  license: LicenseInfo
  isLoading: boolean
  setLicense: (license: LicenseInfo) => void
  clearLicense: () => void
  hasFeature: (feature: Feature) => boolean
  hasTier: (tier: Tier) => boolean
  checkLimit: (
    type: keyof LicenseInfo["limits"],
    current: number
  ) => boolean
}

export const useLicenseStore = create<LicenseState>((set, get) => ({
  license: defaultCELicense,
  isLoading: true,

  setLicense: (license) => set({ license, isLoading: false }),

  clearLicense: () => set({ license: defaultCELicense, isLoading: false }),

  hasFeature: (feature) => {
    const { license } = get()
    return license.features.includes(feature)
  },

  hasTier: (tier) => {
    const { license } = get()
    return tierLevels[license.tier] >= tierLevels[tier]
  },

  checkLimit: (type, current) => {
    const { license } = get()
    const limit = license.limits[type]
    if (limit < 0) return true // unlimited
    return current < limit
  },
}))
