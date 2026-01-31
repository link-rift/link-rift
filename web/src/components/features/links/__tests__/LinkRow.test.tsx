import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from "vitest"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { render } from "@/test/utils"
import LinkRow from "../LinkRow"
import { server } from "@/test/mocks/server"
import type { Link } from "@/types/link"

const baseLink: Link = {
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

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe("LinkRow", () => {
  const defaultProps = {
    link: baseLink,
    selected: false,
    onSelect: vi.fn(),
    onEdit: vi.fn(),
  }

  it("renders short URL and title", () => {
    render(<LinkRow {...defaultProps} />)

    expect(screen.getByText("https://lnk.rf/abc123")).toBeInTheDocument()
    expect(screen.getByText("Example Link")).toBeInTheDocument()
  })

  it("renders total clicks", () => {
    render(<LinkRow {...defaultProps} />)
    expect(screen.getByText("42")).toBeInTheDocument()
    expect(screen.getByText("clicks")).toBeInTheDocument()
  })

  it("renders destination URL when title is absent", () => {
    const link = { ...baseLink, title: null }
    render(<LinkRow {...defaultProps} link={link} />)
    expect(screen.getByText("https://example.com/very-long-url")).toBeInTheDocument()
  })

  it("shows Protected badge when link has password", () => {
    const link = { ...baseLink, has_password: true }
    render(<LinkRow {...defaultProps} link={link} />)
    expect(screen.getByText("Protected")).toBeInTheDocument()
  })

  it("shows Disabled badge when link is inactive", () => {
    const link = { ...baseLink, is_active: false }
    render(<LinkRow {...defaultProps} link={link} />)
    expect(screen.getByText("Disabled")).toBeInTheDocument()
  })

  it("shows Expired badge when link is expired", () => {
    const link = { ...baseLink, expires_at: "2024-01-01T00:00:00Z" }
    render(<LinkRow {...defaultProps} link={link} />)
    expect(screen.getByText("Expired")).toBeInTheDocument()
  })

  it("shows Limit Reached badge when clicks >= max_clicks", () => {
    const link = { ...baseLink, max_clicks: 42, total_clicks: 42 }
    render(<LinkRow {...defaultProps} link={link} />)
    expect(screen.getByText("Limit Reached")).toBeInTheDocument()
  })

  it("does not show badges for normal active link", () => {
    render(<LinkRow {...defaultProps} />)
    expect(screen.queryByText("Protected")).not.toBeInTheDocument()
    expect(screen.queryByText("Disabled")).not.toBeInTheDocument()
    expect(screen.queryByText("Expired")).not.toBeInTheDocument()
    expect(screen.queryByText("Limit Reached")).not.toBeInTheDocument()
  })

  it("calls onSelect when checkbox is clicked", async () => {
    const onSelect = vi.fn()
    render(<LinkRow {...defaultProps} onSelect={onSelect} />)

    const checkbox = screen.getByRole("checkbox")
    await userEvent.click(checkbox)
    expect(onSelect).toHaveBeenCalledWith(baseLink.id)
  })

  it("copies short URL to clipboard on short URL click", async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText: writeTextMock } })

    render(<LinkRow {...defaultProps} />)

    await userEvent.click(screen.getByText("https://lnk.rf/abc123"))
    expect(writeTextMock).toHaveBeenCalledWith("https://lnk.rf/abc123")
  })
})
