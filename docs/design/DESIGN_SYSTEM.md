# Design System

> Last Updated: 2025-01-24

Visual design language, principles, and guidelines for Linkrift.

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
- [Color System](#color-system)
- [Typography](#typography)
- [Spacing and Layout](#spacing-and-layout)
- [Component Principles](#component-principles)
- [Micro-interactions](#micro-interactions)
- [Dark Mode](#dark-mode)
- [Accessibility](#accessibility)
- [Design Inspiration](#design-inspiration)

---

## Design Philosophy

### Core Principles

1. **Clarity over decoration** — Every element serves a purpose
2. **Data-first design** — Optimize for scanning and understanding data
3. **Consistent but not rigid** — Follow patterns, allow for context
4. **Fast feels good** — Perceived performance matters

### Brand Personality

| Attribute | Expression |
|-----------|------------|
| **Professional** | Clean layouts, restrained colors, quality typography |
| **Technical** | Monospace accents, precise spacing, data-rich interfaces |
| **Confident** | Bold headings, clear hierarchy, decisive CTAs |
| **Approachable** | Rounded corners, warm neutrals, helpful empty states |

---

## Anti-Patterns to Avoid

> ⚠️ These patterns create a generic "AI-generated" look. Avoid them.

### Color Anti-Patterns

```
❌ Generic purple-to-blue gradients everywhere
❌ Rainbow gradient borders
❌ Gradient text on every heading
❌ Neon glow effects
```

### Layout Anti-Patterns

```
❌ Overly rounded corners on everything (rounded-2xl abuse)
❌ Cookie-cutter card layouts with identical spacing
❌ Excessive whitespace without purpose
❌ Floating mockups in hero sections
```

### Visual Anti-Patterns

```
❌ Glassmorphism overuse with excessive blur
❌ Overuse of drop shadows on every element
❌ Generic blob/abstract illustrations
❌ Generic stock illustrations without customization
❌ Emoji as design elements in professional UI
```

### What Makes Linkrift Different

```
✅ Purposeful use of color (brand accent for CTAs only)
✅ Subtle shadows and borders that feel intentional
✅ Data-dense layouts optimized for power users
✅ Custom iconography and consistent visual language
✅ Monospace fonts for codes and technical data
```

---

## Color System

### Primary Palette

```css
/* Brand Colors - Distinctive teal/cyan, not default Tailwind */
--color-brand-50: #ecfeff;
--color-brand-100: #cffafe;
--color-brand-200: #a5f3fc;
--color-brand-300: #67e8f9;
--color-brand-400: #22d3ee;
--color-brand-500: #06b6d4;  /* Primary */
--color-brand-600: #0891b2;  /* Hover */
--color-brand-700: #0e7490;
--color-brand-800: #155e75;
--color-brand-900: #164e63;
```

### Neutral Palette

```css
/* Warm neutrals with slight warmth */
--color-gray-50: #fafaf9;
--color-gray-100: #f5f5f4;
--color-gray-200: #e7e5e4;
--color-gray-300: #d6d3d1;
--color-gray-400: #a8a29e;
--color-gray-500: #78716c;
--color-gray-600: #57534e;
--color-gray-700: #44403c;
--color-gray-800: #292524;
--color-gray-900: #1c1917;
--color-gray-950: #0c0a09;
```

### Semantic Colors

```css
/* Success - Green */
--color-success-50: #f0fdf4;
--color-success-500: #22c55e;
--color-success-600: #16a34a;

/* Warning - Amber */
--color-warning-50: #fffbeb;
--color-warning-500: #f59e0b;
--color-warning-600: #d97706;

/* Error - Red */
--color-error-50: #fef2f2;
--color-error-500: #ef4444;
--color-error-600: #dc2626;

/* Info - Blue */
--color-info-50: #eff6ff;
--color-info-500: #3b82f6;
--color-info-600: #2563eb;
```

### Color Usage Ratios

```
60% — Neutral backgrounds and text
30% — Secondary accents, borders, dividers
10% — Primary brand color for CTAs and key actions
```

### Tailwind Configuration

```typescript
// tailwind.config.ts
export default {
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#ecfeff',
          100: '#cffafe',
          200: '#a5f3fc',
          300: '#67e8f9',
          400: '#22d3ee',
          500: '#06b6d4',
          600: '#0891b2',
          700: '#0e7490',
          800: '#155e75',
          900: '#164e63',
        },
        // Override gray with warm neutrals
        gray: {
          50: '#fafaf9',
          100: '#f5f5f4',
          200: '#e7e5e4',
          300: '#d6d3d1',
          400: '#a8a29e',
          500: '#78716c',
          600: '#57534e',
          700: '#44403c',
          800: '#292524',
          900: '#1c1917',
          950: '#0c0a09',
        },
      },
    },
  },
}
```

---

## Typography

### Font Stack

```css
/* Display font - Headings */
--font-display: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;

/* Body font - Content */
--font-body: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;

/* Monospace - Code, short codes, technical */
--font-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
```

### Type Scale

| Level | Size | Line Height | Weight | Usage |
|-------|------|-------------|--------|-------|
| `text-xs` | 12px | 16px | 400 | Labels, metadata |
| `text-sm` | 14px | 20px | 400 | Body small, table cells |
| `text-base` | 16px | 24px | 400 | Body default |
| `text-lg` | 18px | 28px | 500 | Subheadings |
| `text-xl` | 20px | 28px | 600 | Section titles |
| `text-2xl` | 24px | 32px | 600 | Page titles |
| `text-3xl` | 30px | 36px | 700 | Hero headings |

### Typography Classes

```css
/* Heading styles */
.heading-1 {
  @apply text-3xl font-bold tracking-tight text-gray-900;
}

.heading-2 {
  @apply text-2xl font-semibold tracking-tight text-gray-900;
}

.heading-3 {
  @apply text-xl font-semibold text-gray-900;
}

/* Body styles */
.body-default {
  @apply text-base text-gray-700 leading-relaxed;
}

.body-small {
  @apply text-sm text-gray-600;
}

/* Technical/Code */
.code-inline {
  @apply font-mono text-sm bg-gray-100 px-1.5 py-0.5 rounded;
}

.short-code {
  @apply font-mono text-brand-600 font-medium;
}
```

---

## Spacing and Layout

### Spacing Scale

Based on 4px grid:

```css
--spacing-0: 0;
--spacing-1: 4px;    /* 0.25rem */
--spacing-2: 8px;    /* 0.5rem */
--spacing-3: 12px;   /* 0.75rem */
--spacing-4: 16px;   /* 1rem */
--spacing-5: 20px;   /* 1.25rem */
--spacing-6: 24px;   /* 1.5rem */
--spacing-8: 32px;   /* 2rem */
--spacing-10: 40px;  /* 2.5rem */
--spacing-12: 48px;  /* 3rem */
--spacing-16: 64px;  /* 4rem */
--spacing-20: 80px;  /* 5rem */
```

### Container Widths

```css
--container-sm: 640px;
--container-md: 768px;
--container-lg: 1024px;
--container-xl: 1280px;
--container-2xl: 1536px;
```

### Layout Patterns

```
┌─────────────────────────────────────────────────────────────┐
│ Header (h-16, border-b)                                     │
├──────────────┬──────────────────────────────────────────────┤
│              │                                              │
│  Sidebar     │  Main Content                                │
│  (w-64)      │  (flex-1, p-6)                               │
│              │                                              │
│              │  ┌────────────────────────────────────────┐  │
│              │  │ Page Header (mb-6)                     │  │
│              │  ├────────────────────────────────────────┤  │
│              │  │                                        │  │
│              │  │ Content Area                           │  │
│              │  │                                        │  │
│              │  └────────────────────────────────────────┘  │
│              │                                              │
└──────────────┴──────────────────────────────────────────────┘
```

---

## Component Principles

### Buttons

```tsx
// Solid (Primary actions)
<Button variant="default">Create Link</Button>
// → bg-brand-500 text-white hover:bg-brand-600

// Outline (Secondary actions)
<Button variant="outline">Cancel</Button>
// → border-gray-300 text-gray-700 hover:bg-gray-50

// Ghost (Tertiary actions)
<Button variant="ghost">Learn more</Button>
// → text-gray-600 hover:bg-gray-100

// Destructive
<Button variant="destructive">Delete</Button>
// → bg-error-500 text-white hover:bg-error-600
```

### Cards

```css
/* Default card */
.card {
  @apply bg-white rounded-lg border border-gray-200;
}

/* Elevated card (for dialogs, dropdowns) */
.card-elevated {
  @apply bg-white rounded-lg shadow-lg border border-gray-100;
}

/* Interactive card */
.card-interactive {
  @apply bg-white rounded-lg border border-gray-200
         hover:border-gray-300 hover:shadow-sm
         transition-all duration-150;
}
```

### Inputs

```css
/* Default input */
.input {
  @apply w-full px-3 py-2 text-sm
         border border-gray-300 rounded-md
         placeholder:text-gray-400
         focus:outline-none focus:ring-2 focus:ring-brand-500/20 focus:border-brand-500
         transition-colors duration-150;
}

/* Error state */
.input-error {
  @apply border-error-500 focus:ring-error-500/20 focus:border-error-500;
}
```

### Tables

```css
/* Table container */
.table-container {
  @apply overflow-hidden rounded-lg border border-gray-200;
}

/* Table header */
.table-header {
  @apply bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider;
}

/* Table row */
.table-row {
  @apply border-t border-gray-100 hover:bg-gray-50/50;
}

/* Table cell */
.table-cell {
  @apply px-4 py-3 text-sm text-gray-700;
}
```

---

## Micro-interactions

### Timing and Easing

```css
/* Duration tokens */
--duration-fast: 100ms;
--duration-normal: 150ms;
--duration-slow: 300ms;

/* Easing - Never use linear for UI */
--ease-out: cubic-bezier(0.33, 1, 0.68, 1);
--ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
```

### Hover States

```css
/* Button hover */
.btn:hover {
  transform: translateY(-1px);
  transition: transform var(--duration-fast) var(--ease-out);
}

/* Card hover */
.card:hover {
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.05);
  transition: box-shadow var(--duration-normal) var(--ease-out);
}

/* Link hover */
.link:hover {
  color: var(--color-brand-600);
  transition: color var(--duration-fast);
}
```

### Loading States

```tsx
// Skeleton loading (preferred over spinners)
<div className="animate-pulse">
  <div className="h-4 bg-gray-200 rounded w-3/4" />
  <div className="h-4 bg-gray-200 rounded w-1/2 mt-2" />
</div>

// Shimmer effect
.skeleton {
  background: linear-gradient(
    90deg,
    var(--color-gray-200) 0%,
    var(--color-gray-100) 50%,
    var(--color-gray-200) 100%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

### Success/Error Feedback

```tsx
// Toast notification
<Toast variant="success">
  Link created successfully
</Toast>

// Inline validation
<span className="text-error-500 text-sm mt-1">
  This field is required
</span>

// Success checkmark animation
.success-check {
  animation: checkmark 0.3s ease-out;
}

@keyframes checkmark {
  0% { transform: scale(0); }
  50% { transform: scale(1.2); }
  100% { transform: scale(1); }
}
```

---

## Dark Mode

### Strategy

Dark mode is not just color inversion:

1. **Reduce contrast** — Less harsh on eyes
2. **Elevate with lightness** — Raised elements are lighter, not darker
3. **Adjust shadows** — Use darker shadows, not drop shadows
4. **Preserve brand** — Brand colors stay consistent

### Dark Mode Palette

```css
/* Dark mode surfaces */
--dark-bg-base: #0c0a09;        /* gray-950 */
--dark-bg-elevated: #1c1917;    /* gray-900 */
--dark-bg-elevated-2: #292524;  /* gray-800 */

/* Dark mode text */
--dark-text-primary: #fafaf9;   /* gray-50 */
--dark-text-secondary: #a8a29e; /* gray-400 */
--dark-text-muted: #78716c;     /* gray-500 */

/* Dark mode borders */
--dark-border-default: #44403c; /* gray-700 */
--dark-border-subtle: #292524;  /* gray-800 */
```

### Implementation

```tsx
// Tailwind dark mode classes
<div className="bg-white dark:bg-gray-900">
  <h1 className="text-gray-900 dark:text-gray-50">
    Dashboard
  </h1>
  <p className="text-gray-600 dark:text-gray-400">
    Welcome back
  </p>
</div>

// Card in dark mode
<div className="
  bg-white dark:bg-gray-800
  border-gray-200 dark:border-gray-700
  shadow-sm dark:shadow-none
">
  Card content
</div>
```

---

## Accessibility

### Color Contrast

All text must meet WCAG 2.1 AA standards:

| Text Type | Minimum Ratio |
|-----------|---------------|
| Normal text (< 18px) | 4.5:1 |
| Large text (≥ 18px or 14px bold) | 3:1 |
| UI components | 3:1 |

### Focus Indicators

```css
/* Custom focus ring (not browser default) */
.focus-ring {
  @apply focus:outline-none focus:ring-2 focus:ring-brand-500 focus:ring-offset-2;
}

/* Visible focus for keyboard navigation */
:focus-visible {
  outline: 2px solid var(--color-brand-500);
  outline-offset: 2px;
}

/* Remove focus ring for mouse users */
:focus:not(:focus-visible) {
  outline: none;
}
```

### Touch Targets

```css
/* Minimum touch target: 44x44px */
.touch-target {
  min-width: 44px;
  min-height: 44px;
}

/* Icon buttons must have adequate size */
.icon-button {
  @apply p-2.5; /* 10px padding + 24px icon = 44px */
}
```

### Screen Reader Support

```tsx
// Visually hidden but accessible
<span className="sr-only">Open menu</span>

// Accessible loading state
<div aria-busy="true" aria-live="polite">
  Loading...
</div>

// Accessible error message
<input aria-invalid="true" aria-describedby="error-email" />
<span id="error-email" role="alert">
  Please enter a valid email
</span>
```

---

## Design Inspiration

### Reference Applications

| Application | What to Learn |
|-------------|---------------|
| **Linear** | Clean data tables, keyboard-first design, subtle animations |
| **Vercel** | Minimal dashboard, monospace accents, dark mode execution |
| **Stripe** | Data visualization, professional tone, documentation quality |
| **Raycast** | Command palette, speed-focused UI, keyboard shortcuts |
| **Figma** | Collaborative features, real-time indicators, efficient toolbars |

### What Makes Linkrift Unique

1. **Short code prominence** — Monospace display, easy copy
2. **Analytics-first** — Dashboard optimized for data scanning
3. **QR code previews** — Inline QR generation and preview
4. **Domain badges** — Visual indication of custom domains
5. **Real-time updates** — Live click counters, activity indicators

---

## Related Documentation

- [UI Components](UI_COMPONENTS.md) — Component specifications
- [Frontend Architecture](../architecture/FRONTEND_ARCHITECTURE.md) — React patterns
- [Accessibility](../reference/ACCESSIBILITY.md) — WCAG compliance
