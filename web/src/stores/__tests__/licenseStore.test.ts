import { describe, it, expect, beforeEach } from "vitest"
import { useLicenseStore } from "../licenseStore"
import type { LicenseInfo } from "@/types/license"

function makeProLicense(): LicenseInfo {
  return {
    type: "subscription",
    tier: "pro",
    plan: {
      tier: "pro",
      name: "Pro",
      description: "Professional plan",
      price: "$19/mo",
    },
    features: [
      "link_expiration",
      "custom_domains",
      "qr_customization",
      "api_access",
      "advanced_analytics",
    ],
    limits: {
      max_users: 5,
      max_domains: 10,
      max_links_per_month: 10000,
      max_clicks_per_month: 1000000,
      max_workspaces: 5,
      max_api_requests_per_min: 100,
    },
    is_community: false,
  }
}

describe("licenseStore", () => {
  beforeEach(() => {
    useLicenseStore.setState({
      license: {
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
      },
      isLoading: true,
    })
  })

  it("starts with free community license", () => {
    const state = useLicenseStore.getState()
    expect(state.license.tier).toBe("free")
    expect(state.license.is_community).toBe(true)
  })

  describe("setLicense", () => {
    it("updates license and stops loading", () => {
      const pro = makeProLicense()
      useLicenseStore.getState().setLicense(pro)

      const state = useLicenseStore.getState()
      expect(state.license.tier).toBe("pro")
      expect(state.isLoading).toBe(false)
    })
  })

  describe("clearLicense", () => {
    it("resets to default free license", () => {
      useLicenseStore.getState().setLicense(makeProLicense())
      useLicenseStore.getState().clearLicense()

      const state = useLicenseStore.getState()
      expect(state.license.tier).toBe("free")
      expect(state.license.is_community).toBe(true)
    })
  })

  describe("hasFeature", () => {
    it("returns true for included feature", () => {
      useLicenseStore.getState().setLicense(makeProLicense())
      expect(useLicenseStore.getState().hasFeature("custom_domains")).toBe(true)
    })

    it("returns false for missing feature", () => {
      expect(useLicenseStore.getState().hasFeature("custom_domains")).toBe(false)
    })

    it("returns true for free feature", () => {
      expect(useLicenseStore.getState().hasFeature("link_expiration")).toBe(true)
    })
  })

  describe("hasTier", () => {
    it("free has free tier", () => {
      expect(useLicenseStore.getState().hasTier("free")).toBe(true)
    })

    it("free does not have pro tier", () => {
      expect(useLicenseStore.getState().hasTier("pro")).toBe(false)
    })

    it("pro has free and pro tier", () => {
      useLicenseStore.getState().setLicense(makeProLicense())
      expect(useLicenseStore.getState().hasTier("free")).toBe(true)
      expect(useLicenseStore.getState().hasTier("pro")).toBe(true)
    })

    it("pro does not have business tier", () => {
      useLicenseStore.getState().setLicense(makeProLicense())
      expect(useLicenseStore.getState().hasTier("business")).toBe(false)
    })
  })

  describe("checkLimit", () => {
    it("returns true when under limit", () => {
      expect(
        useLicenseStore.getState().checkLimit("max_links_per_month", 50)
      ).toBe(true)
    })

    it("returns false when at limit", () => {
      expect(
        useLicenseStore.getState().checkLimit("max_links_per_month", 100)
      ).toBe(false)
    })

    it("returns false when over limit", () => {
      expect(
        useLicenseStore.getState().checkLimit("max_links_per_month", 150)
      ).toBe(false)
    })

    it("returns true for unlimited (-1)", () => {
      const license = makeProLicense()
      license.limits.max_links_per_month = -1
      useLicenseStore.getState().setLicense(license)
      expect(
        useLicenseStore.getState().checkLimit("max_links_per_month", 999999)
      ).toBe(true)
    })
  })
})
