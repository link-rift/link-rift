import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import * as linkService from "@/services/links"
import type { CreateLinkRequest, UpdateLinkRequest } from "@/types/link"

export function useLinks(params?: {
  search?: string
  is_active?: boolean
  limit?: number
  offset?: number
}) {
  return useQuery({
    queryKey: ["links", params],
    queryFn: () => linkService.getLinks(params),
    staleTime: 30 * 1000,
  })
}

export function useLink(id: string) {
  return useQuery({
    queryKey: ["links", id],
    queryFn: () => linkService.getLink(id),
    enabled: !!id,
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
  return useQuery({
    queryKey: ["links", id, "stats"],
    queryFn: () => linkService.getLinkStats(id),
    enabled: !!id,
    staleTime: 60 * 1000,
  })
}
