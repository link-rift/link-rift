import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from "vitest"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { render } from "@/test/utils"
import BulkActions from "../BulkActions"
import { server } from "@/test/mocks/server"

beforeAll(() => server.listen({ onUnhandledRequest: "bypass" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

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

  it("prompts confirm before deleting", async () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false)
    const onClear = vi.fn()

    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={onClear} />)

    await userEvent.click(screen.getByRole("button", { name: "Delete Selected" }))
    expect(confirmSpy).toHaveBeenCalledWith("Are you sure you want to delete 1 link(s)?")
    // Should not call onClear since confirm was false
    expect(onClear).not.toHaveBeenCalled()
    confirmSpy.mockRestore()
  })

  it("deletes and clears when confirm is accepted", async () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true)
    const onClear = vi.fn()

    render(<BulkActions selectedCount={1} selectedIds={new Set(["a"])} onClear={onClear} />)

    await userEvent.click(screen.getByRole("button", { name: "Delete Selected" }))
    expect(confirmSpy).toHaveBeenCalled()
    expect(onClear).toHaveBeenCalledOnce()
    confirmSpy.mockRestore()
  })
})
