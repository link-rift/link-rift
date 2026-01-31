import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as domainService from "@/services/domains"
import type { CreateDomainRequest } from "@/types/domain"

export function useDomains() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["domains", wsId],
    queryFn: () => domainService.getDomains(),
    staleTime: 30 * 1000,
    enabled: !!wsId,
  })
}

export function useDomain(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["domains", wsId, id],
    queryFn: () => domainService.getDomain(id),
    enabled: !!id && !!wsId,
  })
}

export function useCreateDomain() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateDomainRequest) => domainService.createDomain(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["domains"] })
    },
  })
}

export function useVerifyDomain() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => domainService.verifyDomain(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["domains"] })
    },
  })
}

export function useDeleteDomain() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => domainService.deleteDomain(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["domains"] })
    },
  })
}

export function useDNSRecords(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["domains", wsId, id, "dns-records"],
    queryFn: () => domainService.getDNSRecords(id),
    enabled: !!id && !!wsId,
  })
}
