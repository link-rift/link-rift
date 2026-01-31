import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useLicenseStore } from "@/stores/licenseStore"
import { useAuthStore } from "@/stores/authStore"
import * as licenseService from "@/services/license"
import type { ActivateLicenseRequest } from "@/types/license"

export function useLicense() {
  const { setLicense, clearLicense } = useLicenseStore()
  const { isAuthenticated } = useAuthStore()

  return useQuery({
    queryKey: ["license"],
    queryFn: async () => {
      try {
        const license = await licenseService.getLicense()
        setLicense(license)
        return license
      } catch {
        clearLicense()
        return null
      }
    },
    enabled: isAuthenticated,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}

export function useActivateLicense() {
  const { setLicense } = useLicenseStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: ActivateLicenseRequest) =>
      licenseService.activateLicense(data),
    onSuccess: (license) => {
      setLicense(license)
      queryClient.setQueryData(["license"], license)
    },
  })
}

export function useDeactivateLicense() {
  const { setLicense } = useLicenseStore()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => licenseService.deactivateLicense(),
    onSuccess: (license) => {
      setLicense(license)
      queryClient.setQueryData(["license"], license)
    },
  })
}
