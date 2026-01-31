import { create } from "zustand"
import type { LinkFilter } from "@/types/link"

interface LinkState {
  selectedIds: Set<string>
  filter: LinkFilter
  setFilter: (filter: LinkFilter) => void
  toggleSelected: (id: string) => void
  selectAll: (ids: string[]) => void
  clearSelected: () => void
  isSelected: (id: string) => boolean
}

export const useLinkStore = create<LinkState>((set, get) => ({
  selectedIds: new Set(),
  filter: {},

  setFilter: (filter) => set({ filter }),

  toggleSelected: (id) =>
    set((state) => {
      const newSet = new Set(state.selectedIds)
      if (newSet.has(id)) {
        newSet.delete(id)
      } else {
        newSet.add(id)
      }
      return { selectedIds: newSet }
    }),

  selectAll: (ids) => set({ selectedIds: new Set(ids) }),

  clearSelected: () => set({ selectedIds: new Set() }),

  isSelected: (id) => get().selectedIds.has(id),
}))
