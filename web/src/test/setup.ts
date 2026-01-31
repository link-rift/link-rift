import "@testing-library/jest-dom/vitest"

// Polyfill ResizeObserver for jsdom (required by Radix UI)
globalThis.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
