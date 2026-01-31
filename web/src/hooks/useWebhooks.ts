import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as webhookService from "@/services/webhooks"
import type { CreateWebhookRequest } from "@/types/webhook"

export function useWebhooks() {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["webhooks", wsId],
    queryFn: () => webhookService.getWebhooks(),
    staleTime: 30 * 1000,
    enabled: !!wsId,
  })
}

export function useCreateWebhook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateWebhookRequest) => webhookService.createWebhook(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] })
    },
  })
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => webhookService.deleteWebhook(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] })
    },
  })
}

export function useWebhookDeliveries(
  webhookId: string,
  params?: { limit?: number; offset?: number }
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["webhook-deliveries", wsId, webhookId, params],
    queryFn: () => webhookService.getWebhookDeliveries(webhookId, params),
    staleTime: 15 * 1000,
    enabled: !!wsId && !!webhookId,
  })
}
