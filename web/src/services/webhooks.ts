import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { Webhook, WebhookDelivery, CreateWebhookRequest, CreateWebhookResponse } from "@/types/webhook"

function wsBase(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return `/workspaces/${ws.id}/webhooks`
}

export async function createWebhook(data: CreateWebhookRequest): Promise<CreateWebhookResponse> {
  const res = await apiRequest<CreateWebhookResponse>(wsBase(), {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create webhook")
  }
  return res.data
}

export async function getWebhooks(): Promise<Webhook[]> {
  const res = await apiRequest<Webhook[]>(wsBase())
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch webhooks")
  }
  return res.data
}

export async function deleteWebhook(id: string): Promise<void> {
  const res = await apiRequest<{ message: string }>(`${wsBase()}/${id}`, {
    method: "DELETE",
  })
  if (!res.success) {
    throw new Error(res.error?.message || "Failed to delete webhook")
  }
}

export async function getWebhookDeliveries(
  webhookId: string,
  params?: { limit?: number; offset?: number }
): Promise<{ deliveries: WebhookDelivery[]; total: number }> {
  const searchParams = new URLSearchParams()
  if (params?.limit) searchParams.set("limit", String(params.limit))
  if (params?.offset) searchParams.set("offset", String(params.offset))

  const qs = searchParams.toString()
  const url = qs ? `${wsBase()}/${webhookId}/deliveries?${qs}` : `${wsBase()}/${webhookId}/deliveries`
  const res = await apiRequest<WebhookDelivery[]>(url)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch webhook deliveries")
  }
  return {
    deliveries: res.data,
    total: res.meta?.total ?? res.data.length,
  }
}
