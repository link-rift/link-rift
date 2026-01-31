import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from "vitest"
import { screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { render } from "@/test/utils"
import CreateLinkModal from "../CreateLinkModal"
import { server } from "@/test/mocks/server"

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe("CreateLinkModal", () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
  }

  it("renders modal title", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByText("Create New Link")).toBeInTheDocument()
  })

  it("renders required URL field", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByLabelText("Destination URL")).toBeInTheDocument()
    expect(screen.getByPlaceholderText("https://example.com/long-url")).toBeInTheDocument()
  })

  it("renders optional short code field", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByLabelText("Custom Short Code (optional)")).toBeInTheDocument()
  })

  it("renders optional title field", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByLabelText("Title (optional)")).toBeInTheDocument()
  })

  it("shows cancel and create buttons", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Create Link" })).toBeInTheDocument()
  })

  it("disables Create Link button when URL is empty", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.getByRole("button", { name: "Create Link" })).toBeDisabled()
  })

  it("enables Create Link button when URL is provided", async () => {
    render(<CreateLinkModal {...defaultProps} />)

    const urlInput = screen.getByPlaceholderText("https://example.com/long-url")
    await userEvent.type(urlInput, "https://example.com")

    expect(screen.getByRole("button", { name: "Create Link" })).toBeEnabled()
  })

  it("calls onOpenChange when Cancel is clicked", async () => {
    const onOpenChange = vi.fn()
    render(<CreateLinkModal {...defaultProps} onOpenChange={onOpenChange} />)

    await userEvent.click(screen.getByRole("button", { name: "Cancel" }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it("does not show advanced options by default", () => {
    render(<CreateLinkModal {...defaultProps} />)
    expect(screen.queryByLabelText("Description")).not.toBeInTheDocument()
    expect(screen.queryByLabelText("Password Protection")).not.toBeInTheDocument()
  })

  it("shows advanced options when toggle is clicked", async () => {
    render(<CreateLinkModal {...defaultProps} />)

    const toggle = screen.getByRole("switch")
    await userEvent.click(toggle)

    expect(screen.getByLabelText("Description")).toBeInTheDocument()
    expect(screen.getByLabelText("Password Protection")).toBeInTheDocument()
    expect(screen.getByLabelText("Expiration Date")).toBeInTheDocument()
    expect(screen.getByLabelText("Max Clicks")).toBeInTheDocument()
  })

  it("submits form and closes modal on success", async () => {
    const onOpenChange = vi.fn()
    render(<CreateLinkModal {...defaultProps} onOpenChange={onOpenChange} />)

    const urlInput = screen.getByPlaceholderText("https://example.com/long-url")
    await userEvent.type(urlInput, "https://example.com")
    await userEvent.click(screen.getByRole("button", { name: "Create Link" }))

    await waitFor(() => {
      expect(onOpenChange).toHaveBeenCalledWith(false)
    })
  })

  it("is hidden when open is false", () => {
    render(<CreateLinkModal open={false} onOpenChange={vi.fn()} />)
    expect(screen.queryByText("Create New Link")).not.toBeInTheDocument()
  })
})
