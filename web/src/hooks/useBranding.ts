import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as brandingService from "@/services/branding"
import type { UpdateBrandingInput } from "@/types/branding"

export function useBranding() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["branding", wsId],
    queryFn: () => brandingService.getBranding(),
    enabled: !!wsId,
    retry: false,
  })
}

export function useUpdateBranding() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: (input: UpdateBrandingInput) =>
      brandingService.updateBranding(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["branding", wsId] })
    },
  })
}

export function useDeleteBranding() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: () => brandingService.deleteBranding(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["branding", wsId] })
    },
  })
}
