import { describe, it, expect, beforeEach } from "vitest"
import { useLinkStore } from "../linkStore"

describe("linkStore", () => {
  beforeEach(() => {
    useLinkStore.setState({
      selectedIds: new Set(),
      filter: {},
    })
  })

  it("starts with empty selectedIds", () => {
    const state = useLinkStore.getState()
    expect(state.selectedIds.size).toBe(0)
  })

  it("starts with empty filter", () => {
    const state = useLinkStore.getState()
    expect(state.filter).toEqual({})
  })

  describe("toggleSelected", () => {
    it("adds id when not selected", () => {
      useLinkStore.getState().toggleSelected("link-1")
      expect(useLinkStore.getState().selectedIds.has("link-1")).toBe(true)
    })

    it("removes id when already selected", () => {
      useLinkStore.getState().toggleSelected("link-1")
      useLinkStore.getState().toggleSelected("link-1")
      expect(useLinkStore.getState().selectedIds.has("link-1")).toBe(false)
    })

    it("handles multiple ids independently", () => {
      useLinkStore.getState().toggleSelected("link-1")
      useLinkStore.getState().toggleSelected("link-2")
      const ids = useLinkStore.getState().selectedIds
      expect(ids.has("link-1")).toBe(true)
      expect(ids.has("link-2")).toBe(true)
      expect(ids.size).toBe(2)
    })
  })

  describe("selectAll", () => {
    it("replaces selectedIds with provided ids", () => {
      useLinkStore.getState().toggleSelected("old-id")
      useLinkStore.getState().selectAll(["a", "b", "c"])
      const ids = useLinkStore.getState().selectedIds
      expect(ids.size).toBe(3)
      expect(ids.has("a")).toBe(true)
      expect(ids.has("b")).toBe(true)
      expect(ids.has("c")).toBe(true)
      expect(ids.has("old-id")).toBe(false)
    })
  })

  describe("clearSelected", () => {
    it("empties selectedIds", () => {
      useLinkStore.getState().toggleSelected("link-1")
      useLinkStore.getState().toggleSelected("link-2")
      useLinkStore.getState().clearSelected()
      expect(useLinkStore.getState().selectedIds.size).toBe(0)
    })
  })

  describe("isSelected", () => {
    it("returns true for selected id", () => {
      useLinkStore.getState().toggleSelected("link-1")
      expect(useLinkStore.getState().isSelected("link-1")).toBe(true)
    })

    it("returns false for unselected id", () => {
      expect(useLinkStore.getState().isSelected("nonexistent")).toBe(false)
    })
  })

  describe("setFilter", () => {
    it("sets search filter", () => {
      useLinkStore.getState().setFilter({ search: "test" })
      expect(useLinkStore.getState().filter).toEqual({ search: "test" })
    })

    it("sets is_active filter", () => {
      useLinkStore.getState().setFilter({ is_active: true })
      expect(useLinkStore.getState().filter).toEqual({ is_active: true })
    })

    it("replaces entire filter object", () => {
      useLinkStore.getState().setFilter({ search: "old" })
      useLinkStore.getState().setFilter({ is_active: false })
      expect(useLinkStore.getState().filter).toEqual({ is_active: false })
    })
  })
})
