import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as linkService from "@/services/links"
import type { CreateLinkRequest, UpdateLinkRequest } from "@/types/link"

export function useLinks(params?: {
  search?: string
  is_active?: boolean
  limit?: number
  offset?: number
}) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["links", wsId, params],
    queryFn: () => linkService.getLinks(params),
    staleTime: 30 * 1000,
    enabled: !!wsId,
  })
}

export function useLink(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["links", wsId, id],
    queryFn: () => linkService.getLink(id),
    enabled: !!id && !!wsId,
  })
}

export function useCreateLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateLinkRequest) => linkService.createLink(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] })
    },
  })
}

export function useUpdateLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateLinkRequest }) =>
      linkService.updateLink(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] })
    },
  })
}

export function useDeleteLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => linkService.deleteLink(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] })
    },
  })
}

export function useBulkCreateLinks() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (links: CreateLinkRequest[]) =>
      linkService.bulkCreateLinks({ links }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["links"] })
    },
  })
}

export function useLinkStats(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["links", wsId, id, "stats"],
    queryFn: () => linkService.getLinkStats(id),
    enabled: !!id && !!wsId,
    staleTime: 60 * 1000,
  })
}
