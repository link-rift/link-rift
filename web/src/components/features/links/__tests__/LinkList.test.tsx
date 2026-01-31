import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from "vitest"
import { screen } from "@testing-library/react"
import { render } from "@/test/utils"
import LinkList from "../LinkList"
import { server } from "@/test/mocks/server"
import type { Link } from "@/types/link"

const makeLink = (overrides: Partial<Link> = {}): Link => ({
  id: "11111111-1111-1111-1111-111111111111",
  user_id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  workspace_id: "00000000-0000-0000-0000-000000000000",
  url: "https://example.com",
  short_code: "abc123",
  short_url: "https://lnk.rf/abc123",
  title: "Test Link",
  is_active: true,
  has_password: false,
  total_clicks: 10,
  unique_clicks: 5,
  created_at: "2025-01-15T10:00:00Z",
  updated_at: "2025-01-15T10:00:00Z",
  ...overrides,
})

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe("LinkList", () => {
  const defaultProps = {
    links: [] as Link[],
    isLoading: false,
    selectedIds: new Set<string>(),
    onSelect: vi.fn(),
    onEdit: vi.fn(),
  }

  it("shows loading skeletons when isLoading is true", () => {
    const { container } = render(<LinkList {...defaultProps} isLoading={true} />)
    // Skeletons render 5 placeholder items
    const skeletons = container.querySelectorAll("[class*='animate-pulse'], [data-slot='skeleton']")
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it("shows empty state when no links", () => {
    render(<LinkList {...defaultProps} />)
    expect(screen.getByText("No links yet")).toBeInTheDocument()
    expect(screen.getByText(/Create your first shortened link/)).toBeInTheDocument()
  })

  it("renders link rows for provided links", () => {
    const links = [
      makeLink({ id: "1", short_url: "https://lnk.rf/aaa", title: "Link A" }),
      makeLink({ id: "2", short_url: "https://lnk.rf/bbb", title: "Link B" }),
    ]
    render(<LinkList {...defaultProps} links={links} />)
    expect(screen.getByText("https://lnk.rf/aaa")).toBeInTheDocument()
    expect(screen.getByText("https://lnk.rf/bbb")).toBeInTheDocument()
    expect(screen.getByText("Link A")).toBeInTheDocument()
    expect(screen.getByText("Link B")).toBeInTheDocument()
  })

  it("passes selected state to link rows", () => {
    const links = [
      makeLink({ id: "1", short_url: "https://lnk.rf/aaa" }),
      makeLink({ id: "2", short_url: "https://lnk.rf/bbb" }),
    ]
    const selectedIds = new Set(["1"])
    render(<LinkList {...defaultProps} links={links} selectedIds={selectedIds} />)

    const checkboxes = screen.getAllByRole("checkbox")
    expect(checkboxes).toHaveLength(2)
    // First should be checked, second unchecked
    expect(checkboxes[0]).toHaveAttribute("data-state", "checked")
    expect(checkboxes[1]).toHaveAttribute("data-state", "unchecked")
  })
})
