import { apiRequest } from "./api"
import type { LicenseInfo, ActivateLicenseRequest } from "@/types/license"

export async function getLicense(): Promise<LicenseInfo> {
  const res = await apiRequest<LicenseInfo>("/license")
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to get license info")
  }
  return res.data
}

export async function activateLicense(
  data: ActivateLicenseRequest
): Promise<LicenseInfo> {
  const res = await apiRequest<LicenseInfo>("/license", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to activate license")
  }
  return res.data
}

export async function deactivateLicense(): Promise<LicenseInfo> {
  const res = await apiRequest<LicenseInfo>("/license", {
    method: "DELETE",
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to deactivate license")
  }
  return res.data
}
