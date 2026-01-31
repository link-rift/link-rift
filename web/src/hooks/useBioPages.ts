import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as bioPageService from "@/services/biopages"
import type {
  CreateBioPageRequest,
  UpdateBioPageRequest,
  CreateBioPageLinkRequest,
  UpdateBioPageLinkRequest,
  ReorderBioLinksRequest,
} from "@/types/biopage"

// Bio Pages

export function useBioPages() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["bio-pages", wsId],
    queryFn: () => bioPageService.getBioPages(),
    staleTime: 30 * 1000,
    enabled: !!wsId,
  })
}

export function useBioPage(id: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["bio-pages", wsId, id],
    queryFn: () => bioPageService.getBioPage(id),
    enabled: !!id && !!wsId,
  })
}

export function useCreateBioPage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBioPageRequest) => bioPageService.createBioPage(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useUpdateBioPage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateBioPageRequest }) =>
      bioPageService.updateBioPage(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useDeleteBioPage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => bioPageService.deleteBioPage(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function usePublishBioPage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => bioPageService.publishBioPage(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useUnpublishBioPage() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => bioPageService.unpublishBioPage(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

// Bio Page Links

export function useBioPageLinks(pageId: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["bio-pages", wsId, pageId, "links"],
    queryFn: () => bioPageService.getBioPageLinks(pageId),
    enabled: !!pageId && !!wsId,
  })
}

export function useAddBioPageLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ pageId, data }: { pageId: string; data: CreateBioPageLinkRequest }) =>
      bioPageService.addBioPageLink(pageId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useUpdateBioPageLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      pageId,
      linkId,
      data,
    }: {
      pageId: string
      linkId: string
      data: UpdateBioPageLinkRequest
    }) => bioPageService.updateBioPageLink(pageId, linkId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useDeleteBioPageLink() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ pageId, linkId }: { pageId: string; linkId: string }) =>
      bioPageService.deleteBioPageLink(pageId, linkId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

export function useReorderBioPageLinks() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ pageId, data }: { pageId: string; data: ReorderBioLinksRequest }) =>
      bioPageService.reorderBioPageLinks(pageId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bio-pages"] })
    },
  })
}

// Themes

export function useBioPageThemes() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["bio-themes", wsId],
    queryFn: () => bioPageService.getBioPageThemes(),
    staleTime: 5 * 60 * 1000, // themes rarely change
    enabled: !!wsId,
  })
}

// Public

export function usePublicBioPage(slug: string) {
  return useQuery({
    queryKey: ["public-bio-page", slug],
    queryFn: () => bioPageService.getPublicBioPage(slug),
    enabled: !!slug,
    staleTime: 60 * 1000,
  })
}
