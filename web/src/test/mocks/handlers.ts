import { http, HttpResponse } from "msw"
import type { Link, LinkQuickStats } from "@/types/link"

export const mockLink: Link = {
  id: "11111111-1111-1111-1111-111111111111",
  user_id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  workspace_id: "00000000-0000-0000-0000-000000000000",
  url: "https://example.com/very-long-url",
  short_code: "abc123",
  short_url: "https://lnk.rf/abc123",
  title: "Example Link",
  is_active: true,
  has_password: false,
  total_clicks: 42,
  unique_clicks: 30,
  created_at: "2025-01-15T10:00:00Z",
  updated_at: "2025-01-15T10:00:00Z",
}

export const mockLinks: Link[] = [
  mockLink,
  {
    ...mockLink,
    id: "22222222-2222-2222-2222-222222222222",
    short_code: "def456",
    short_url: "https://lnk.rf/def456",
    url: "https://another.com/page",
    title: "Another Link",
    has_password: true,
    is_active: false,
    total_clicks: 10,
    unique_clicks: 5,
  },
  {
    ...mockLink,
    id: "33333333-3333-3333-3333-333333333333",
    short_code: "ghi789",
    short_url: "https://lnk.rf/ghi789",
    url: "https://expired.com",
    title: "Expired Link",
    expires_at: "2024-01-01T00:00:00Z",
    total_clicks: 100,
    unique_clicks: 80,
    max_clicks: 100,
  },
]

export const mockStats: LinkQuickStats = {
  total_clicks: 42,
  unique_clicks: 30,
  clicks_24h: 5,
  clicks_7d: 20,
  created_at: "2025-01-15T10:00:00Z",
}

export const handlers = [
  http.get("/api/v1/links", () => {
    return HttpResponse.json({
      success: true,
      data: mockLinks,
      meta: { total: mockLinks.length, limit: 20, offset: 0 },
    })
  }),

  http.get("/api/v1/links/:id", ({ params }) => {
    const link = mockLinks.find((l) => l.id === params.id)
    if (!link) {
      return HttpResponse.json(
        { success: false, error: { code: "NOT_FOUND", message: "Link not found" } },
        { status: 404 },
      )
    }
    return HttpResponse.json({ success: true, data: link })
  }),

  http.post("/api/v1/links", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>
    const newLink: Link = {
      ...mockLink,
      id: "44444444-4444-4444-4444-444444444444",
      url: (body.url as string) || mockLink.url,
      short_code: (body.short_code as string) || "new123",
      short_url: `https://lnk.rf/${(body.short_code as string) || "new123"}`,
      title: (body.title as string) || null,
      total_clicks: 0,
      unique_clicks: 0,
    }
    return HttpResponse.json({ success: true, data: newLink }, { status: 201 })
  }),

  http.put("/api/v1/links/:id", async ({ params, request }) => {
    const body = (await request.json()) as Record<string, unknown>
    const link = mockLinks.find((l) => l.id === params.id)
    if (!link) {
      return HttpResponse.json(
        { success: false, error: { code: "NOT_FOUND", message: "Link not found" } },
        { status: 404 },
      )
    }
    return HttpResponse.json({
      success: true,
      data: { ...link, ...body, updated_at: new Date().toISOString() },
    })
  }),

  http.delete("/api/v1/links/:id", ({ params }) => {
    const link = mockLinks.find((l) => l.id === params.id)
    if (!link) {
      return HttpResponse.json(
        { success: false, error: { code: "NOT_FOUND", message: "Link not found" } },
        { status: 404 },
      )
    }
    return HttpResponse.json({ success: true, data: { message: "Link deleted" } })
  }),

  http.post("/api/v1/links/bulk", () => {
    return HttpResponse.json({ success: true, data: [mockLink] }, { status: 201 })
  }),

  http.get("/api/v1/links/:id/stats", ({ params }) => {
    const link = mockLinks.find((l) => l.id === params.id)
    if (!link) {
      return HttpResponse.json(
        { success: false, error: { code: "NOT_FOUND", message: "Link not found" } },
        { status: 404 },
      )
    }
    return HttpResponse.json({ success: true, data: mockStats })
  }),
]
