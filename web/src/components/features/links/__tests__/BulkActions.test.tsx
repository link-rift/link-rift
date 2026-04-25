import { describe, it, expect, vi, beforeAll, afterAll, afterEach, beforeEach } from "vitest"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { render } from "@/test/utils"
import BulkActions from "../BulkActions"
import { server } from "@/test/mocks/server"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { MOCK_WORKSPACE_ID } from "@/test/mocks/handlers"

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

beforeEach(() => {
  useWorkspaceStore.setState({
    currentWorkspace: {
      id: MOCK_WORKSPACE_ID,
      name: "Test",
      slug: "test",
      owner_id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      plan: "free",
      settings: null,
      created_at: "2025-01-15T10:00:00Z",
      updated_at: "2025-01-15T10:00:00Z",
    },
  })
})

describe("BulkActions", () => {
  it("renders nothing when selectedCount is 0", () => {
    const { container } = render(
      <BulkActions selectedCount={0} selectedIds={new Set()} onClear={vi.fn()} />,
    )
    expect(container.firstChild).toBeNull()
  })

  it("shows selection count for single item", () => {
    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={vi.fn()} />)
    expect(screen.getByText("1 link selected")).toBeInTheDocument()
  })

  it("shows pluralized selection count", () => {
    render(
      <BulkActions selectedCount={3} selectedIds={new Set(["a", "b", "c"])} onClear={vi.fn()} />,
    )
    expect(screen.getByText("3 links selected")).toBeInTheDocument()
  })

  it("renders Delete Selected and Clear Selection buttons", () => {
    render(<BulkActions selectedCount={2} selectedIds={new Set(["a", "b"])} onClear={vi.fn()} />)
    expect(screen.getByRole("button", { name: "Delete Selected" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Clear Selection" })).toBeInTheDocument()
  })

  it("calls onClear when Clear Selection is clicked", async () => {
    const onClear = vi.fn()
    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={onClear} />)

    await userEvent.click(screen.getByRole("button", { name: "Clear Selection" }))
    expect(onClear).toHaveBeenCalledOnce()
  })

  it("opens confirmation dialog and does not delete when canceled", async () => {
    const onClear = vi.fn()

    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={onClear} />)

    await userEvent.click(screen.getByRole("button", { name: "Delete Selected" }))

    const dialog = await screen.findByRole("alertdialog")
    expect(dialog).toBeInTheDocument()

    await userEvent.click(screen.getByRole("button", { name: "Cancel" }))
    expect(onClear).not.toHaveBeenCalled()
  })

  it("deletes and clears when confirmed in dialog", async () => {
    const onClear = vi.fn()

    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={onClear} />)

    await userEvent.click(screen.getByRole("button", { name: "Delete Selected" }))

    const dialog = await screen.findByRole("alertdialog")
    const confirmButton = await screen.findByRole("button", { name: "Delete" })
    expect(dialog).toContainElement(confirmButton)

    await userEvent.click(confirmButton)
    expect(onClear).toHaveBeenCalledOnce()
  })
})
