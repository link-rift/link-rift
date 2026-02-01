# Accessibility Guide

> Last Updated: 2025-01-24

Comprehensive guide for ensuring Linkrift meets WCAG 2.1 AA accessibility standards and provides an inclusive experience for all users.

---

## Table of Contents

- [Overview](#overview)
- [WCAG 2.1 AA Compliance](#wcag-21-aa-compliance)
  - [Perceivable](#perceivable)
  - [Operable](#operable)
  - [Understandable](#understandable)
  - [Robust](#robust)
- [Keyboard Navigation](#keyboard-navigation)
  - [Focus Management](#focus-management)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
  - [Skip Links](#skip-links)
- [Screen Reader Support](#screen-reader-support)
  - [ARIA Labels](#aria-labels)
  - [Live Regions](#live-regions)
  - [Semantic HTML](#semantic-html)
- [Visual Accessibility](#visual-accessibility)
  - [Color Contrast](#color-contrast)
  - [Text Sizing](#text-sizing)
  - [Motion and Animation](#motion-and-animation)
- [Forms and Inputs](#forms-and-inputs)
- [Testing](#testing)
  - [Automated Testing with axe-core](#automated-testing-with-axe-core)
  - [Manual Testing](#manual-testing)
  - [Screen Reader Testing](#screen-reader-testing)
- [Component Patterns](#component-patterns)
- [Accessibility Checklist](#accessibility-checklist)

---

## Overview

Linkrift is committed to digital accessibility. We strive to ensure our platform is usable by everyone, including people who:

- Are blind or have low vision
- Are deaf or hard of hearing
- Have motor impairments
- Have cognitive disabilities
- Use assistive technologies

**Our Commitment:**
- WCAG 2.1 Level AA compliance
- Regular accessibility audits
- Continuous improvement based on user feedback
- Accessible documentation

---

## WCAG 2.1 AA Compliance

### Perceivable

**1.1 Text Alternatives**

All non-text content must have text alternatives:

```tsx
// Images
<img src="/logo.png" alt="Linkrift logo" />

// Decorative images
<img src="/decoration.png" alt="" role="presentation" />

// Icons with meaning
<button aria-label="Delete link">
  <TrashIcon aria-hidden="true" />
</button>

// Complex images
<figure>
  <img
    src="/analytics-chart.png"
    alt="Bar chart showing link clicks over time"
    aria-describedby="chart-desc"
  />
  <figcaption id="chart-desc">
    Link clicks increased by 25% in the last 30 days,
    from 1,000 to 1,250 clicks.
  </figcaption>
</figure>
```

**1.3 Adaptable**

Content must be presented in different ways without losing meaning:

```tsx
// Use semantic HTML
<nav aria-label="Main navigation">
  <ul>
    <li><a href="/dashboard">Dashboard</a></li>
    <li><a href="/analytics">Analytics</a></li>
  </ul>
</nav>

// Data tables with proper headers
<table>
  <caption>Link Performance Summary</caption>
  <thead>
    <tr>
      <th scope="col">Short Code</th>
      <th scope="col">Clicks</th>
      <th scope="col">Created</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>abc123</td>
      <td>1,234</td>
      <td>Jan 15, 2025</td>
    </tr>
  </tbody>
</table>

// Reading order matches visual order
<article>
  <header>
    <h1>Article Title</h1>
    <p>Published on January 24, 2025</p>
  </header>
  <main>Content here...</main>
  <footer>Author information</footer>
</article>
```

**1.4 Distinguishable**

Make content easy to see and hear:

```css
/* Minimum contrast ratio 4.5:1 for normal text */
.text-primary {
  color: #1a1a2e; /* On white background: 14.5:1 */
}

/* Minimum 3:1 for large text (18pt+) */
.heading-large {
  color: #4a4a6a; /* On white background: 7.5:1 */
}

/* Focus indicators must be visible */
:focus-visible {
  outline: 2px solid #2563eb;
  outline-offset: 2px;
}

/* Don't rely on color alone */
.error-field {
  border-color: #ef4444;
  border-width: 2px; /* Visual indicator beyond color */
}
.error-field::before {
  content: '⚠ '; /* Icon indicator */
}
```

### Operable

**2.1 Keyboard Accessible**

All functionality must be available via keyboard:

```tsx
// Custom interactive elements need keyboard support
function CustomButton({ onClick, children }: CustomButtonProps) {
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick();
    }
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={handleKeyDown}
    >
      {children}
    </div>
  );
}

// Better: Use native button element
function Button({ onClick, children }: ButtonProps) {
  return <button onClick={onClick}>{children}</button>;
}
```

**2.2 Enough Time**

Users must have enough time to read and use content:

```tsx
// Allow users to extend time limits
function SessionTimeout() {
  const [timeRemaining, setTimeRemaining] = useState(300);
  const [showWarning, setShowWarning] = useState(false);

  useEffect(() => {
    if (timeRemaining === 60) {
      setShowWarning(true);
    }
  }, [timeRemaining]);

  return (
    showWarning && (
      <AlertDialog>
        <AlertDialogContent>
          <AlertDialogTitle>Session Expiring</AlertDialogTitle>
          <AlertDialogDescription>
            Your session will expire in {timeRemaining} seconds.
          </AlertDialogDescription>
          <AlertDialogAction onClick={extendSession}>
            Extend Session
          </AlertDialogAction>
        </AlertDialogContent>
      </AlertDialog>
    )
  );
}
```

**2.3 Seizures and Physical Reactions**

Avoid content that could cause seizures:

```tsx
// Respect reduced motion preferences
const prefersReducedMotion = window.matchMedia(
  '(prefers-reduced-motion: reduce)'
).matches;

function AnimatedComponent() {
  return (
    <motion.div
      animate={{ opacity: 1 }}
      transition={{
        duration: prefersReducedMotion ? 0 : 0.3,
      }}
    />
  );
}
```

```css
/* CSS approach */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

**2.4 Navigable**

Help users navigate and find content:

```tsx
// Skip link for keyboard users
function SkipLink() {
  return (
    <a
      href="#main-content"
      className="skip-link"
    >
      Skip to main content
    </a>
  );
}

// Page title updates
function Dashboard() {
  useEffect(() => {
    document.title = 'Dashboard - Linkrift';
  }, []);
}

// Clear focus order
<form>
  <input tabIndex={0} /> {/* Natural order */}
  <input tabIndex={0} />
  <button tabIndex={0} />
</form>
```

### Understandable

**3.1 Readable**

Make text content readable:

```html
<!-- Set language -->
<html lang="en">

<!-- Mark language changes -->
<p>The French word <span lang="fr">bonjour</span> means hello.</p>
```

**3.2 Predictable**

Web pages should appear and operate predictably:

```tsx
// Don't change context on focus
// Bad
<select onChange={(e) => window.location.href = e.target.value}>

// Good
<select onChange={(e) => setSelectedValue(e.target.value)}>
<button onClick={navigateToValue}>Go</button>
```

**3.3 Input Assistance**

Help users avoid and correct mistakes:

```tsx
function URLInput() {
  const [url, setUrl] = useState('');
  const [error, setError] = useState('');

  const validateURL = (value: string) => {
    try {
      new URL(value);
      setError('');
      return true;
    } catch {
      setError('Please enter a valid URL starting with http:// or https://');
      return false;
    }
  };

  return (
    <div>
      <label htmlFor="url-input">
        Enter URL <span aria-hidden="true">*</span>
        <span className="sr-only">(required)</span>
      </label>
      <input
        id="url-input"
        type="url"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        onBlur={() => validateURL(url)}
        aria-invalid={!!error}
        aria-describedby={error ? 'url-error' : 'url-hint'}
        required
      />
      <p id="url-hint" className="hint">
        Example: https://example.com/page
      </p>
      {error && (
        <p id="url-error" className="error" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
```

### Robust

**4.1 Compatible**

Maximize compatibility with assistive technologies:

```tsx
// Use valid HTML
// Bad
<div onclick="...">Click me</div>

// Good
<button onClick={...}>Click me</button>

// Proper ARIA usage
<div
  role="tablist"
  aria-label="Dashboard sections"
>
  <button
    role="tab"
    aria-selected={activeTab === 'links'}
    aria-controls="links-panel"
    id="links-tab"
  >
    Links
  </button>
</div>
<div
  role="tabpanel"
  id="links-panel"
  aria-labelledby="links-tab"
  hidden={activeTab !== 'links'}
>
  {/* Tab content */}
</div>
```

---

## Keyboard Navigation

### Focus Management

```tsx
// Focus trap for modals
import { FocusTrap } from '@radix-ui/react-focus-trap';

function Modal({ isOpen, onClose, children }) {
  const closeButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (isOpen) {
      closeButtonRef.current?.focus();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <FocusTrap>
      <div role="dialog" aria-modal="true" aria-labelledby="modal-title">
        <h2 id="modal-title">Modal Title</h2>
        {children}
        <button ref={closeButtonRef} onClick={onClose}>
          Close
        </button>
      </div>
    </FocusTrap>
  );
}

// Restore focus after modal closes
function ParentComponent() {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const [isOpen, setIsOpen] = useState(false);

  const handleClose = () => {
    setIsOpen(false);
    triggerRef.current?.focus(); // Return focus to trigger
  };

  return (
    <>
      <button ref={triggerRef} onClick={() => setIsOpen(true)}>
        Open Modal
      </button>
      <Modal isOpen={isOpen} onClose={handleClose}>
        Modal content
      </Modal>
    </>
  );
}
```

### Keyboard Shortcuts

```tsx
// Document keyboard shortcuts
function KeyboardShortcuts() {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger shortcuts when typing in inputs
      if (e.target instanceof HTMLInputElement ||
          e.target instanceof HTMLTextAreaElement) {
        return;
      }

      // Ctrl/Cmd + K: Open search
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        openSearch();
      }

      // N: New link (when not in input)
      if (e.key === 'n') {
        openNewLinkModal();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return null;
}

// Keyboard shortcuts help
function ShortcutsHelp() {
  return (
    <div role="region" aria-label="Keyboard shortcuts">
      <h2>Keyboard Shortcuts</h2>
      <dl>
        <dt><kbd>Ctrl</kbd> + <kbd>K</kbd></dt>
        <dd>Open search</dd>
        <dt><kbd>N</kbd></dt>
        <dd>Create new link</dd>
        <dt><kbd>?</kbd></dt>
        <dd>Show this help</dd>
      </dl>
    </div>
  );
}
```

### Skip Links

```tsx
// Skip link component
function SkipLinks() {
  return (
    <nav aria-label="Skip links" className="skip-links">
      <a href="#main-content">Skip to main content</a>
      <a href="#navigation">Skip to navigation</a>
      <a href="#search">Skip to search</a>
    </nav>
  );
}
```

```css
.skip-links a {
  position: absolute;
  left: -10000px;
  top: auto;
  width: 1px;
  height: 1px;
  overflow: hidden;
}

.skip-links a:focus {
  position: fixed;
  top: 0;
  left: 0;
  width: auto;
  height: auto;
  padding: 1rem;
  background: #1a1a2e;
  color: white;
  z-index: 9999;
}
```

---

## Screen Reader Support

### ARIA Labels

```tsx
// Labeling interactive elements
<button aria-label="Close modal">
  <XIcon aria-hidden="true" />
</button>

// Labeling regions
<nav aria-label="Main navigation">...</nav>
<nav aria-label="Footer navigation">...</nav>

// Describing state
<button aria-expanded={isOpen} aria-controls="dropdown-menu">
  Menu
</button>
<ul id="dropdown-menu" hidden={!isOpen}>
  ...
</ul>

// Loading states
<button aria-busy={isLoading} disabled={isLoading}>
  {isLoading ? 'Saving...' : 'Save'}
</button>

// Current page in navigation
<nav>
  <a href="/dashboard" aria-current="page">Dashboard</a>
  <a href="/analytics">Analytics</a>
</nav>
```

### Live Regions

```tsx
// Announce dynamic content changes
function Notifications() {
  const [message, setMessage] = useState('');

  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className="sr-only"
    >
      {message}
    </div>
  );
}

// Announce urgent messages
function ErrorAlert({ error }: { error: string }) {
  return (
    <div role="alert" aria-live="assertive">
      Error: {error}
    </div>
  );
}

// Progress updates
function UploadProgress({ progress }: { progress: number }) {
  return (
    <div
      role="progressbar"
      aria-valuenow={progress}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label="Upload progress"
    >
      <div style={{ width: `${progress}%` }} />
      <span aria-live="polite">{progress}% complete</span>
    </div>
  );
}
```

### Semantic HTML

```tsx
// Use proper heading hierarchy
<main>
  <h1>Dashboard</h1>
  <section>
    <h2>Your Links</h2>
    <article>
      <h3>Link Details</h3>
    </article>
  </section>
  <section>
    <h2>Analytics</h2>
  </section>
</main>

// Use landmark elements
<header role="banner">...</header>
<nav role="navigation">...</nav>
<main role="main">...</main>
<aside role="complementary">...</aside>
<footer role="contentinfo">...</footer>

// Lists for groups
<ul aria-label="Recent links">
  <li>Link 1</li>
  <li>Link 2</li>
</ul>
```

---

## Visual Accessibility

### Color Contrast

```typescript
// Color contrast utilities
function getContrastRatio(color1: string, color2: string): number {
  const lum1 = getLuminance(color1);
  const lum2 = getLuminance(color2);
  const lighter = Math.max(lum1, lum2);
  const darker = Math.min(lum1, lum2);
  return (lighter + 0.05) / (darker + 0.05);
}

// Ensure 4.5:1 for normal text, 3:1 for large text
const colors = {
  text: {
    primary: '#1a1a2e',    // 14.5:1 on white
    secondary: '#4a5568',  // 7.0:1 on white
    muted: '#718096',      // 4.5:1 on white (minimum)
  },
  background: {
    primary: '#ffffff',
    secondary: '#f7fafc',
  },
};
```

```css
/* High contrast mode support */
@media (prefers-contrast: high) {
  :root {
    --color-text: #000000;
    --color-background: #ffffff;
    --color-border: #000000;
  }

  button {
    border: 2px solid currentColor;
  }
}
```

### Text Sizing

```css
/* Use relative units */
html {
  font-size: 100%; /* Respects user preferences */
}

body {
  font-size: 1rem; /* 16px default */
  line-height: 1.5;
}

h1 { font-size: 2rem; }
h2 { font-size: 1.5rem; }
p { font-size: 1rem; }
small { font-size: 0.875rem; }

/* Support 200% zoom */
@media (min-width: 320px) {
  .container {
    max-width: 100%;
    padding: 1rem;
  }
}

/* Minimum touch target size: 44x44px */
button, a {
  min-height: 44px;
  min-width: 44px;
}
```

### Motion and Animation

```tsx
// Hook for reduced motion
function useReducedMotion() {
  const [reducedMotion, setReducedMotion] = useState(
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  );

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handleChange = () => setReducedMotion(mediaQuery.matches);
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  return reducedMotion;
}

// Usage
function AnimatedCard() {
  const reducedMotion = useReducedMotion();

  return (
    <motion.div
      initial={{ opacity: 0, y: reducedMotion ? 0 : 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: reducedMotion ? 0 : 0.3 }}
    >
      Card content
    </motion.div>
  );
}
```

---

## Forms and Inputs

```tsx
function AccessibleForm() {
  return (
    <form aria-label="Create new link">
      {/* Required field */}
      <div className="field">
        <label htmlFor="url">
          URL
          <span aria-hidden="true" className="required">*</span>
          <span className="sr-only">(required)</span>
        </label>
        <input
          id="url"
          type="url"
          required
          aria-required="true"
          aria-describedby="url-hint url-error"
        />
        <p id="url-hint" className="hint">
          Enter the full URL including https://
        </p>
        <p id="url-error" className="error" role="alert">
          {/* Error message appears here */}
        </p>
      </div>

      {/* Optional field with character count */}
      <div className="field">
        <label htmlFor="custom-code">
          Custom short code (optional)
        </label>
        <input
          id="custom-code"
          type="text"
          maxLength={20}
          aria-describedby="code-hint code-count"
        />
        <p id="code-hint" className="hint">
          Letters, numbers, and hyphens only
        </p>
        <p id="code-count" aria-live="polite">
          {charCount}/20 characters
        </p>
      </div>

      {/* Checkbox */}
      <div className="field">
        <input
          id="track-clicks"
          type="checkbox"
          aria-describedby="track-hint"
        />
        <label htmlFor="track-clicks">
          Enable click tracking
        </label>
        <p id="track-hint" className="hint">
          Track anonymous click statistics
        </p>
      </div>

      {/* Submit */}
      <button type="submit">Create Link</button>
    </form>
  );
}
```

---

## Testing

### Automated Testing with axe-core

```typescript
// vitest setup
import { configureAxe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

const axe = configureAxe({
  rules: {
    region: { enabled: false }, // Disable specific rules if needed
  },
});

// Component test
import { render } from '@testing-library/react';
import { axe } from 'jest-axe';

describe('LinkForm', () => {
  it('should have no accessibility violations', async () => {
    const { container } = render(<LinkForm />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
```

```typescript
// Playwright accessibility testing
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Accessibility', () => {
  test('dashboard has no violations', async ({ page }) => {
    await page.goto('/dashboard');

    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
      .analyze();

    expect(results.violations).toEqual([]);
  });

  test('create link modal has no violations', async ({ page }) => {
    await page.goto('/dashboard');
    await page.click('button:has-text("New Link")');

    const results = await new AxeBuilder({ page })
      .include('[role="dialog"]')
      .analyze();

    expect(results.violations).toEqual([]);
  });
});
```

### Manual Testing

**Keyboard Testing Checklist:**
- [ ] Can reach all interactive elements with Tab
- [ ] Focus order is logical
- [ ] Focus indicator is visible
- [ ] Can activate buttons with Enter/Space
- [ ] Can close modals with Escape
- [ ] Can navigate dropdowns with arrow keys
- [ ] Skip link works

**Screen Reader Testing:**
1. Test with VoiceOver (macOS)
2. Test with NVDA (Windows)
3. Test with JAWS (Windows)

### Screen Reader Testing

```bash
# macOS VoiceOver
# Press Cmd + F5 to enable
# Use VO + arrows to navigate
# Use VO + Space to activate

# Windows NVDA
# Download from https://www.nvaccess.org/
# Use NVDA + arrows to navigate
# Use Enter/Space to activate
```

**What to test:**
- Page title is announced
- Headings are properly structured
- Links and buttons are descriptive
- Form labels are associated
- Error messages are announced
- Dynamic content updates are announced

---

## Component Patterns

### Accessible Modal

```tsx
import * as Dialog from '@radix-ui/react-dialog';

function AccessibleModal({ trigger, title, children }) {
  return (
    <Dialog.Root>
      <Dialog.Trigger asChild>
        {trigger}
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="modal-overlay" />
        <Dialog.Content
          className="modal-content"
          aria-describedby="modal-description"
        >
          <Dialog.Title>{title}</Dialog.Title>
          <Dialog.Description id="modal-description">
            {children}
          </Dialog.Description>
          <Dialog.Close asChild>
            <button aria-label="Close">
              <XIcon aria-hidden="true" />
            </button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
```

### Accessible Dropdown

```tsx
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';

function AccessibleDropdown() {
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button aria-label="Options menu">
          <MoreIcon aria-hidden="true" />
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content>
          <DropdownMenu.Item onSelect={handleEdit}>
            Edit
          </DropdownMenu.Item>
          <DropdownMenu.Item onSelect={handleDelete}>
            Delete
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
```

---

## Accessibility Checklist

### Before Every Release

- [ ] Run automated axe-core tests
- [ ] Test keyboard navigation
- [ ] Test with screen reader
- [ ] Check color contrast
- [ ] Verify focus indicators
- [ ] Test at 200% zoom
- [ ] Test with reduced motion enabled
- [ ] Validate HTML

### Monthly Audit

- [ ] Full WCAG 2.1 AA audit
- [ ] User testing with assistive technology users
- [ ] Review and address accessibility issues
- [ ] Update documentation

---

## Resources

- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [axe DevTools Browser Extension](https://www.deque.com/axe/devtools/)
- [ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [Inclusive Components](https://inclusive-components.design/)

For accessibility issues or suggestions, please open a GitHub issue with the `accessibility` label.
