import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as qrService from "@/services/qrcodes"
import type { CreateQRCodeRequest, BulkQRCodeRequest } from "@/types/qrcode"

export function useQRCodeForLink(linkId: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["qrcodes", wsId, linkId],
    queryFn: () => qrService.getQRCodeForLink(linkId),
    enabled: !!linkId && !!wsId,
    retry: (failureCount, error) => {
      // Don't retry on 404 (no QR code exists yet)
      if (error instanceof Error && error.message.includes("not found")) {
        return false
      }
      return failureCount < 3
    },
  })
}

export function useCreateQRCode() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      linkId,
      data,
    }: {
      linkId: string
      data: CreateQRCodeRequest
    }) => qrService.createQRCode(linkId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["qrcodes"] })
    },
  })
}

export function useDownloadQRCode() {
  return useMutation({
    mutationFn: ({
      linkId,
      format,
    }: {
      linkId: string
      format: "png" | "svg"
    }) => qrService.downloadQRCode(linkId, format),
  })
}

export function useBulkGenerateQRCodes() {
  return useMutation({
    mutationFn: (data: BulkQRCodeRequest) =>
      qrService.bulkGenerateQRCodes(data),
  })
}

export function useStyleTemplates() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["qr-templates", wsId],
    queryFn: () => qrService.getStyleTemplates(),
    staleTime: Infinity,
    enabled: !!wsId,
  })
}
