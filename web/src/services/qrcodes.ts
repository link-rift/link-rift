import { apiRequest } from "./api"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type {
  QRCode,
  CreateQRCodeRequest,
  BulkQRCodeRequest,
  QRStyleTemplate,
} from "@/types/qrcode"

function getWorkspaceId(): string {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")
  return ws.id
}

function wsBase(): string {
  return `/workspaces/${getWorkspaceId()}`
}

export async function createQRCode(
  linkId: string,
  data: CreateQRCodeRequest
): Promise<QRCode> {
  const res = await apiRequest<QRCode>(`${wsBase()}/links/${linkId}/qr`, {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to create QR code")
  }
  return res.data
}

export async function getQRCodeForLink(linkId: string): Promise<QRCode> {
  const res = await apiRequest<QRCode>(`${wsBase()}/links/${linkId}/qr`)
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch QR code")
  }
  return res.data
}

export async function downloadQRCode(
  linkId: string,
  format: "png" | "svg" = "png"
): Promise<Blob> {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")

  const token = localStorage.getItem("access_token")
  const url = `/api/v1/workspaces/${ws.id}/links/${linkId}/qr/download?format=${format}`

  const response = await fetch(url, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })

  if (!response.ok) {
    throw new Error(`Failed to download QR code: ${response.statusText}`)
  }

  return response.blob()
}

export async function bulkGenerateQRCodes(
  data: BulkQRCodeRequest
): Promise<Blob> {
  const ws = useWorkspaceStore.getState().currentWorkspace
  if (!ws) throw new Error("No workspace selected")

  const token = localStorage.getItem("access_token")
  const url = `/api/v1/workspaces/${ws.id}/qr/bulk`

  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(data),
  })

  if (!response.ok) {
    throw new Error(`Failed to generate QR codes: ${response.statusText}`)
  }

  return response.blob()
}

export async function getStyleTemplates(): Promise<
  Record<string, QRStyleTemplate>
> {
  const res = await apiRequest<Record<string, QRStyleTemplate>>(
    `${wsBase()}/qr/templates`
  )
  if (!res.success || !res.data) {
    throw new Error(res.error?.message || "Failed to fetch style templates")
  }
  return res.data
}
