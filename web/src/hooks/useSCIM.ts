import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as scimService from "@/services/scim"
import type { CreateSCIMTokenInput } from "@/types/scim"

export function useSCIMTokens() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["scim-tokens", wsId],
    queryFn: () => scimService.listSCIMTokens(),
    enabled: !!wsId,
  })
}

export function useCreateSCIMToken() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: (input: CreateSCIMTokenInput) =>
      scimService.createSCIMToken(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scim-tokens", wsId] })
    },
  })
}

export function useRevokeSCIMToken() {
  const queryClient = useQueryClient()
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useMutation({
    mutationFn: (tokenId: string) => scimService.revokeSCIMToken(tokenId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["scim-tokens", wsId] })
    },
  })
}

export function useSCIMSyncLogs(limit = 50, offset = 0) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["scim-sync-logs", wsId, limit, offset],
    queryFn: () => scimService.listSCIMSyncLogs(limit, offset),
    enabled: !!wsId,
  })
}
