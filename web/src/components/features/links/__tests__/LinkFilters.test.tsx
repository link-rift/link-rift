import { describe, it, expect, vi } from "vitest"
import { screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { render } from "@/test/utils"
import LinkFilters from "../LinkFilters"

describe("LinkFilters", () => {
  const defaultProps = {
    search: "",
    onSearchChange: vi.fn(),
    showActive: undefined as boolean | undefined,
    onActiveChange: vi.fn(),
  }

  it("renders search input with placeholder", () => {
    render(<LinkFilters {...defaultProps} />)
    expect(screen.getByPlaceholderText("Search links...")).toBeInTheDocument()
  })

  it("renders filter buttons", () => {
    render(<LinkFilters {...defaultProps} />)
    expect(screen.getByRole("button", { name: "All" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Active" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Inactive" })).toBeInTheDocument()
  })

  it("displays current search value", () => {
    render(<LinkFilters {...defaultProps} search="test query" />)
    expect(screen.getByDisplayValue("test query")).toBeInTheDocument()
  })

  it("calls onSearchChange when typing", async () => {
    const onSearchChange = vi.fn()
    render(<LinkFilters {...defaultProps} onSearchChange={onSearchChange} />)

    const input = screen.getByPlaceholderText("Search links...")
    await userEvent.type(input, "hello")
    expect(onSearchChange).toHaveBeenCalledTimes(5) // one per character
  })

  it("calls onActiveChange(undefined) when All is clicked", async () => {
    const onActiveChange = vi.fn()
    render(<LinkFilters {...defaultProps} onActiveChange={onActiveChange} showActive={true} />)

    await userEvent.click(screen.getByRole("button", { name: "All" }))
    expect(onActiveChange).toHaveBeenCalledWith(undefined)
  })

  it("calls onActiveChange(true) when Active is clicked", async () => {
    const onActiveChange = vi.fn()
    render(<LinkFilters {...defaultProps} onActiveChange={onActiveChange} />)

    await userEvent.click(screen.getByRole("button", { name: "Active" }))
    expect(onActiveChange).toHaveBeenCalledWith(true)
  })

  it("calls onActiveChange(false) when Inactive is clicked", async () => {
    const onActiveChange = vi.fn()
    render(<LinkFilters {...defaultProps} onActiveChange={onActiveChange} />)

    await userEvent.click(screen.getByRole("button", { name: "Inactive" }))
    expect(onActiveChange).toHaveBeenCalledWith(false)
  })
})
