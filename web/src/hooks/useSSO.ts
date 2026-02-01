import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as ssoService from "@/services/sso"
import type { CreateSSOConfigInput, UpdateSSOConfigInput } from "@/types/sso"

export function useSSOConfig() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["sso-config", wsId],
    queryFn: () => ssoService.getSSOConfig(),
    enabled: !!wsId,
    retry: false,
  })
}

export function useCreateSSOConfig() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: (input: CreateSSOConfigInput) =>
      ssoService.createSSOConfig(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-config", wsId] })
    },
  })
}

export function useUpdateSSOConfig() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: (input: UpdateSSOConfigInput) =>
      ssoService.updateSSOConfig(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-config", wsId] })
    },
  })
}

export function useDeleteSSOConfig() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: () => ssoService.deleteSSOConfig(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sso-config", wsId] })
    },
  })
}

export function useSSOIdentities() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["sso-identities", wsId],
    queryFn: () => ssoService.listSSOIdentities(),
    enabled: !!wsId,
  })
}
