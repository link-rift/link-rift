# Browser Extension

> Last Updated: 2025-01-24

Comprehensive guide for building and deploying Linkrift browser extensions for Chrome, Firefox, and other Chromium-based browsers.

---

## Table of Contents

- [Overview](#overview)
- [Chrome Extension (Manifest V3)](#chrome-extension-manifest-v3)
  - [Project Structure](#project-structure)
  - [Manifest Configuration](#manifest-configuration)
  - [Background Service Worker](#background-service-worker)
  - [Popup Interface](#popup-interface)
  - [Content Scripts](#content-scripts)
  - [Options Page](#options-page)
- [Firefox Extension](#firefox-extension)
  - [Firefox-Specific Manifest](#firefox-specific-manifest)
  - [Browser API Compatibility](#browser-api-compatibility)
- [Building with Vite](#building-with-vite)
  - [Vite Configuration](#vite-configuration)
  - [Build Scripts](#build-scripts)
  - [Hot Reload Development](#hot-reload-development)
- [Features](#features)
  - [One-Click Shortening](#one-click-shortening)
  - [Context Menu Integration](#context-menu-integration)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
  - [QR Code Generation](#qr-code-generation)
  - [Analytics Popup](#analytics-popup)
- [Authentication](#authentication)
- [Publishing](#publishing)
  - [Chrome Web Store](#chrome-web-store)
  - [Firefox Add-ons](#firefox-add-ons)
  - [Edge Add-ons](#edge-add-ons)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Linkrift browser extension allows users to shorten URLs directly from their browser with a single click. It supports:

- **Chrome** (and Chromium-based browsers: Edge, Brave, Opera)
- **Firefox**
- **Safari** (via Safari Web Extension converter)

**Key features:**
- One-click URL shortening
- Right-click context menu
- Keyboard shortcuts
- Custom short codes
- QR code generation
- Click analytics
- Sync across devices

---

## Chrome Extension (Manifest V3)

### Project Structure

```
browser-extension/
├── src/
│   ├── background/
│   │   └── service-worker.ts    # Background service worker
│   ├── popup/
│   │   ├── Popup.tsx            # Main popup component
│   │   ├── popup.html           # Popup HTML entry
│   │   └── popup.css            # Popup styles
│   ├── content/
│   │   └── content-script.ts    # Page injection script
│   ├── options/
│   │   ├── Options.tsx          # Options page
│   │   └── options.html         # Options HTML entry
│   ├── components/              # Shared React components
│   ├── hooks/                   # Custom React hooks
│   ├── utils/                   # Utility functions
│   └── types/                   # TypeScript types
├── public/
│   ├── icons/                   # Extension icons
│   └── _locales/               # Internationalization
├── manifest.json                # Extension manifest
├── vite.config.ts              # Vite build config
├── package.json
└── tsconfig.json
```

### Manifest Configuration

```json
{
  "manifest_version": 3,
  "name": "Linkrift - URL Shortener",
  "version": "1.0.0",
  "description": "Shorten URLs instantly with Linkrift. Track clicks and share links easily.",

  "permissions": [
    "activeTab",
    "storage",
    "contextMenus",
    "clipboardWrite"
  ],

  "optional_permissions": [
    "tabs"
  ],

  "host_permissions": [
    "https://api.linkrift.io/*",
    "https://*.linkrift.io/*"
  ],

  "background": {
    "service_worker": "src/background/service-worker.js",
    "type": "module"
  },

  "action": {
    "default_popup": "src/popup/popup.html",
    "default_icon": {
      "16": "icons/icon-16.png",
      "32": "icons/icon-32.png",
      "48": "icons/icon-48.png",
      "128": "icons/icon-128.png"
    },
    "default_title": "Shorten this URL"
  },

  "options_ui": {
    "page": "src/options/options.html",
    "open_in_tab": true
  },

  "content_scripts": [
    {
      "matches": ["<all_urls>"],
      "js": ["src/content/content-script.js"],
      "run_at": "document_idle"
    }
  ],

  "commands": {
    "_execute_action": {
      "suggested_key": {
        "default": "Alt+Shift+L",
        "mac": "Command+Shift+L"
      },
      "description": "Open Linkrift popup"
    },
    "shorten-current": {
      "suggested_key": {
        "default": "Alt+Shift+S",
        "mac": "Command+Shift+S"
      },
      "description": "Shorten current page URL"
    }
  },

  "icons": {
    "16": "icons/icon-16.png",
    "32": "icons/icon-32.png",
    "48": "icons/icon-48.png",
    "128": "icons/icon-128.png"
  },

  "default_locale": "en"
}
```

### Background Service Worker

```typescript
// src/background/service-worker.ts
import { LinkriftAPI } from '../utils/api';
import { StorageManager } from '../utils/storage';

// Initialize context menus on install
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: 'shorten-link',
    title: 'Shorten with Linkrift',
    contexts: ['link'],
  });

  chrome.contextMenus.create({
    id: 'shorten-page',
    title: 'Shorten this page',
    contexts: ['page'],
  });

  chrome.contextMenus.create({
    id: 'shorten-selection',
    title: 'Shorten selected URL',
    contexts: ['selection'],
  });
});

// Handle context menu clicks
chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  let url: string | undefined;

  switch (info.menuItemId) {
    case 'shorten-link':
      url = info.linkUrl;
      break;
    case 'shorten-page':
      url = info.pageUrl;
      break;
    case 'shorten-selection':
      url = info.selectionText;
      break;
  }

  if (url) {
    try {
      const result = await shortenUrl(url);
      await copyToClipboard(result.shortUrl);
      showNotification('URL Shortened!', `Copied: ${result.shortUrl}`);
    } catch (error) {
      showNotification('Error', 'Failed to shorten URL');
    }
  }
});

// Handle keyboard shortcuts
chrome.commands.onCommand.addListener(async (command) => {
  if (command === 'shorten-current') {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (tab?.url) {
      try {
        const result = await shortenUrl(tab.url);
        await copyToClipboard(result.shortUrl);

        // Send message to content script to show toast
        if (tab.id) {
          chrome.tabs.sendMessage(tab.id, {
            type: 'SHOW_TOAST',
            message: `Copied: ${result.shortUrl}`,
          });
        }
      } catch (error) {
        console.error('Failed to shorten URL:', error);
      }
    }
  }
});

// Handle messages from popup/content scripts
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'SHORTEN_URL') {
    shortenUrl(message.url)
      .then(sendResponse)
      .catch((error) => sendResponse({ error: error.message }));
    return true; // Keep channel open for async response
  }

  if (message.type === 'GET_ANALYTICS') {
    getAnalytics(message.shortCode)
      .then(sendResponse)
      .catch((error) => sendResponse({ error: error.message }));
    return true;
  }
});

// API functions
async function shortenUrl(url: string): Promise<ShortenResult> {
  const token = await StorageManager.getAuthToken();
  const api = new LinkriftAPI(token);

  const settings = await StorageManager.getSettings();

  return api.createLink({
    url,
    customCode: settings.autoGenerateCode ? undefined : undefined,
  });
}

async function getAnalytics(shortCode: string): Promise<Analytics> {
  const token = await StorageManager.getAuthToken();
  const api = new LinkriftAPI(token);
  return api.getAnalytics(shortCode);
}

async function copyToClipboard(text: string): Promise<void> {
  // Use offscreen document for clipboard access in service worker
  await chrome.offscreen.createDocument({
    url: 'offscreen.html',
    reasons: [chrome.offscreen.Reason.CLIPBOARD],
    justification: 'Copy shortened URL to clipboard',
  });

  await chrome.runtime.sendMessage({
    type: 'COPY_TO_CLIPBOARD',
    text,
  });

  await chrome.offscreen.closeDocument();
}

function showNotification(title: string, message: string): void {
  chrome.notifications.create({
    type: 'basic',
    iconUrl: 'icons/icon-128.png',
    title,
    message,
  });
}
```

### Popup Interface

```tsx
// src/popup/Popup.tsx
import { useState, useEffect } from 'react';
import { useCurrentTab } from '../hooks/useCurrentTab';
import { useShortenUrl } from '../hooks/useShortenUrl';
import { useAnalytics } from '../hooks/useAnalytics';

export function Popup() {
  const { tab, loading: tabLoading } = useCurrentTab();
  const { shorten, result, loading, error } = useShortenUrl();
  const [customCode, setCustomCode] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleShorten = async () => {
    if (tab?.url) {
      await shorten(tab.url, customCode || undefined);
    }
  };

  const handleCopy = async () => {
    if (result?.shortUrl) {
      await navigator.clipboard.writeText(result.shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="popup-container">
      <header className="popup-header">
        <img src="/icons/logo.svg" alt="Linkrift" className="logo" />
        <h1>Linkrift</h1>
      </header>

      <main className="popup-content">
        {tabLoading ? (
          <div className="loading">Loading...</div>
        ) : (
          <>
            <div className="url-preview">
              <span className="url-label">Current URL:</span>
              <span className="url-text" title={tab?.url}>
                {truncateUrl(tab?.url || '', 50)}
              </span>
            </div>

            {result ? (
              <div className="result">
                <div className="short-url">
                  <input
                    type="text"
                    value={result.shortUrl}
                    readOnly
                    className="short-url-input"
                  />
                  <button
                    onClick={handleCopy}
                    className={`copy-btn ${copied ? 'copied' : ''}`}
                  >
                    {copied ? 'Copied!' : 'Copy'}
                  </button>
                </div>

                <div className="result-actions">
                  <button onClick={() => openQRCode(result.shortUrl)}>
                    QR Code
                  </button>
                  <button onClick={() => openAnalytics(result.shortCode)}>
                    Analytics
                  </button>
                </div>
              </div>
            ) : (
              <>
                <button
                  onClick={handleShorten}
                  disabled={loading || !tab?.url}
                  className="shorten-btn"
                >
                  {loading ? 'Shortening...' : 'Shorten URL'}
                </button>

                <button
                  onClick={() => setShowAdvanced(!showAdvanced)}
                  className="advanced-toggle"
                >
                  {showAdvanced ? 'Hide' : 'Show'} Advanced Options
                </button>

                {showAdvanced && (
                  <div className="advanced-options">
                    <label>
                      Custom short code:
                      <input
                        type="text"
                        value={customCode}
                        onChange={(e) => setCustomCode(e.target.value)}
                        placeholder="my-custom-link"
                        pattern="[a-zA-Z0-9-]+"
                      />
                    </label>
                  </div>
                )}
              </>
            )}

            {error && <div className="error">{error}</div>}
          </>
        )}
      </main>

      <footer className="popup-footer">
        <a href="#" onClick={() => chrome.runtime.openOptionsPage()}>
          Settings
        </a>
        <a href="https://linkrift.io/dashboard" target="_blank">
          Dashboard
        </a>
      </footer>
    </div>
  );
}

function truncateUrl(url: string, maxLength: number): string {
  if (url.length <= maxLength) return url;
  return url.substring(0, maxLength - 3) + '...';
}
```

### Content Scripts

```typescript
// src/content/content-script.ts

// Toast notification for in-page feedback
function showToast(message: string, duration = 3000): void {
  const existing = document.getElementById('linkrift-toast');
  if (existing) existing.remove();

  const toast = document.createElement('div');
  toast.id = 'linkrift-toast';
  toast.textContent = message;
  toast.style.cssText = `
    position: fixed;
    bottom: 20px;
    right: 20px;
    background: #1a1a2e;
    color: white;
    padding: 12px 24px;
    border-radius: 8px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 14px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    z-index: 999999;
    animation: slideIn 0.3s ease-out;
  `;

  // Add animation styles
  const style = document.createElement('style');
  style.textContent = `
    @keyframes slideIn {
      from { transform: translateX(100%); opacity: 0; }
      to { transform: translateX(0); opacity: 1; }
    }
    @keyframes slideOut {
      from { transform: translateX(0); opacity: 1; }
      to { transform: translateX(100%); opacity: 0; }
    }
  `;
  document.head.appendChild(style);

  document.body.appendChild(toast);

  setTimeout(() => {
    toast.style.animation = 'slideOut 0.3s ease-in';
    setTimeout(() => toast.remove(), 300);
  }, duration);
}

// Listen for messages from background script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'SHOW_TOAST') {
    showToast(message.message);
    sendResponse({ success: true });
  }
});

// Add hover tooltip for Linkrift short URLs
function addLinkTooltips(): void {
  document.addEventListener('mouseover', (e) => {
    const target = e.target as HTMLElement;
    if (target.tagName === 'A') {
      const href = (target as HTMLAnchorElement).href;
      if (href.includes('lnkr.ft') || href.includes('linkrift.io')) {
        // Could show analytics preview on hover
      }
    }
  });
}

// Initialize
addLinkTooltips();
```

### Options Page

```tsx
// src/options/Options.tsx
import { useState, useEffect } from 'react';
import { StorageManager, Settings } from '../utils/storage';

export function Options() {
  const [settings, setSettings] = useState<Settings>({
    defaultDomain: 'lnkr.ft',
    autoShorten: false,
    showNotifications: true,
    copyToClipboard: true,
    trackAnalytics: true,
    theme: 'system',
  });
  const [saved, setSaved] = useState(false);
  const [apiKey, setApiKey] = useState('');

  useEffect(() => {
    StorageManager.getSettings().then(setSettings);
    StorageManager.getAuthToken().then((token) => {
      if (token) setApiKey(token.substring(0, 20) + '...');
    });
  }, []);

  const handleSave = async () => {
    await StorageManager.saveSettings(settings);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const handleLogin = () => {
    chrome.tabs.create({ url: 'https://linkrift.io/login?extension=true' });
  };

  const handleLogout = async () => {
    await StorageManager.clearAuthToken();
    setApiKey('');
  };

  return (
    <div className="options-container">
      <header>
        <h1>Linkrift Settings</h1>
      </header>

      <section className="settings-section">
        <h2>Account</h2>
        {apiKey ? (
          <div className="account-info">
            <p>Logged in</p>
            <button onClick={handleLogout}>Log Out</button>
          </div>
        ) : (
          <button onClick={handleLogin}>Log In to Linkrift</button>
        )}
      </section>

      <section className="settings-section">
        <h2>Default Domain</h2>
        <select
          value={settings.defaultDomain}
          onChange={(e) =>
            setSettings({ ...settings, defaultDomain: e.target.value })
          }
        >
          <option value="lnkr.ft">lnkr.ft (default)</option>
          <option value="custom">Custom domain</option>
        </select>
      </section>

      <section className="settings-section">
        <h2>Behavior</h2>
        <label>
          <input
            type="checkbox"
            checked={settings.copyToClipboard}
            onChange={(e) =>
              setSettings({ ...settings, copyToClipboard: e.target.checked })
            }
          />
          Automatically copy shortened URL to clipboard
        </label>
        <label>
          <input
            type="checkbox"
            checked={settings.showNotifications}
            onChange={(e) =>
              setSettings({ ...settings, showNotifications: e.target.checked })
            }
          />
          Show desktop notifications
        </label>
        <label>
          <input
            type="checkbox"
            checked={settings.trackAnalytics}
            onChange={(e) =>
              setSettings({ ...settings, trackAnalytics: e.target.checked })
            }
          />
          Enable click analytics
        </label>
      </section>

      <section className="settings-section">
        <h2>Appearance</h2>
        <select
          value={settings.theme}
          onChange={(e) =>
            setSettings({ ...settings, theme: e.target.value as any })
          }
        >
          <option value="system">System default</option>
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </section>

      <section className="settings-section">
        <h2>Keyboard Shortcuts</h2>
        <p>
          Configure shortcuts in{' '}
          <a href="chrome://extensions/shortcuts" target="_blank">
            Chrome Extension Shortcuts
          </a>
        </p>
        <ul>
          <li>
            <kbd>Alt+Shift+L</kbd> - Open popup
          </li>
          <li>
            <kbd>Alt+Shift+S</kbd> - Shorten current page
          </li>
        </ul>
      </section>

      <button onClick={handleSave} className="save-btn">
        {saved ? 'Saved!' : 'Save Settings'}
      </button>
    </div>
  );
}
```

---

## Firefox Extension

### Firefox-Specific Manifest

Firefox requires some modifications to the manifest:

```json
{
  "manifest_version": 3,
  "name": "Linkrift - URL Shortener",
  "version": "1.0.0",

  "browser_specific_settings": {
    "gecko": {
      "id": "extension@linkrift.io",
      "strict_min_version": "109.0"
    }
  },

  "background": {
    "scripts": ["src/background/service-worker.js"],
    "type": "module"
  },

  "permissions": [
    "activeTab",
    "storage",
    "contextMenus",
    "clipboardWrite"
  ]
}
```

### Browser API Compatibility

Use a compatibility layer for cross-browser support:

```typescript
// src/utils/browser.ts

// Unified browser API
export const browser = globalThis.browser || globalThis.chrome;

// Promise-based API wrapper
export function promisify<T>(
  method: (...args: any[]) => void,
  ...args: any[]
): Promise<T> {
  return new Promise((resolve, reject) => {
    method(...args, (result: T) => {
      if (browser.runtime.lastError) {
        reject(browser.runtime.lastError);
      } else {
        resolve(result);
      }
    });
  });
}

// Storage wrapper
export const storage = {
  async get<T>(keys: string | string[]): Promise<T> {
    return browser.storage.sync.get(keys);
  },
  async set(items: Record<string, any>): Promise<void> {
    return browser.storage.sync.set(items);
  },
  async remove(keys: string | string[]): Promise<void> {
    return browser.storage.sync.remove(keys);
  },
};

// Tabs wrapper
export const tabs = {
  async query(queryInfo: chrome.tabs.QueryInfo): Promise<chrome.tabs.Tab[]> {
    return browser.tabs.query(queryInfo);
  },
  async create(createProperties: chrome.tabs.CreateProperties): Promise<chrome.tabs.Tab> {
    return browser.tabs.create(createProperties);
  },
};
```

---

## Building with Vite

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import { copyFileSync, mkdirSync } from 'fs';

export default defineConfig({
  plugins: [
    react(),
    {
      name: 'copy-manifest',
      buildEnd() {
        // Copy manifest.json to dist
        copyFileSync('manifest.json', 'dist/manifest.json');
      },
    },
  ],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        popup: resolve(__dirname, 'src/popup/popup.html'),
        options: resolve(__dirname, 'src/options/options.html'),
        'service-worker': resolve(__dirname, 'src/background/service-worker.ts'),
        'content-script': resolve(__dirname, 'src/content/content-script.ts'),
      },
      output: {
        entryFileNames: 'src/[name]/[name].js',
        chunkFileNames: 'chunks/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      },
    },
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
});
```

### Build Scripts

```json
{
  "scripts": {
    "dev": "vite build --watch",
    "build": "vite build",
    "build:chrome": "TARGET=chrome vite build",
    "build:firefox": "TARGET=firefox vite build",
    "build:all": "npm run build:chrome && npm run build:firefox",
    "lint": "eslint src --ext .ts,.tsx",
    "type-check": "tsc --noEmit",
    "test": "vitest",
    "package:chrome": "npm run build:chrome && cd dist && zip -r ../linkrift-chrome.zip .",
    "package:firefox": "npm run build:firefox && cd dist && zip -r ../linkrift-firefox.zip ."
  }
}
```

### Hot Reload Development

For development with hot reload:

```typescript
// vite.config.dev.ts
import { defineConfig } from 'vite';

export default defineConfig({
  // ... base config
  build: {
    watch: {
      include: 'src/**',
    },
  },
  plugins: [
    {
      name: 'reload-extension',
      writeBundle() {
        // Trigger extension reload
        fetch('http://localhost:8080/reload').catch(() => {});
      },
    },
  ],
});
```

---

## Features

### One-Click Shortening

Users can shorten the current page with a single click on the extension icon.

### Context Menu Integration

Right-click options:
- **Shorten this link** - Shorten any hyperlink
- **Shorten this page** - Shorten the current page URL
- **Shorten selected URL** - Shorten text selected on the page

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Alt+Shift+L` | Open popup |
| `Alt+Shift+S` | Quick shorten current page |
| `Alt+Shift+C` | Copy last shortened URL |

### QR Code Generation

```tsx
// src/components/QRCode.tsx
import { QRCodeSVG } from 'qrcode.react';

interface QRCodeProps {
  url: string;
  size?: number;
}

export function QRCode({ url, size = 200 }: QRCodeProps) {
  const handleDownload = () => {
    const svg = document.getElementById('qr-code');
    if (svg) {
      const svgData = new XMLSerializer().serializeToString(svg);
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      const img = new Image();

      img.onload = () => {
        canvas.width = size;
        canvas.height = size;
        ctx?.drawImage(img, 0, 0);

        const link = document.createElement('a');
        link.download = 'linkrift-qr.png';
        link.href = canvas.toDataURL('image/png');
        link.click();
      };

      img.src = 'data:image/svg+xml;base64,' + btoa(svgData);
    }
  };

  return (
    <div className="qr-container">
      <QRCodeSVG
        id="qr-code"
        value={url}
        size={size}
        level="H"
        includeMargin
      />
      <button onClick={handleDownload}>Download QR Code</button>
    </div>
  );
}
```

### Analytics Popup

Quick view of link performance directly in the extension.

---

## Authentication

OAuth2 flow for connecting to Linkrift account:

```typescript
// src/utils/auth.ts
const CLIENT_ID = 'linkrift-browser-extension';
const REDIRECT_URI = chrome.identity.getRedirectURL();

export async function login(): Promise<string> {
  const authUrl = new URL('https://linkrift.io/oauth/authorize');
  authUrl.searchParams.set('client_id', CLIENT_ID);
  authUrl.searchParams.set('redirect_uri', REDIRECT_URI);
  authUrl.searchParams.set('response_type', 'token');
  authUrl.searchParams.set('scope', 'links:read links:write analytics:read');

  return new Promise((resolve, reject) => {
    chrome.identity.launchWebAuthFlow(
      {
        url: authUrl.toString(),
        interactive: true,
      },
      (redirectUrl) => {
        if (chrome.runtime.lastError) {
          reject(chrome.runtime.lastError);
          return;
        }

        if (redirectUrl) {
          const url = new URL(redirectUrl);
          const token = url.hash.match(/access_token=([^&]+)/)?.[1];

          if (token) {
            resolve(token);
          } else {
            reject(new Error('No token in response'));
          }
        }
      }
    );
  });
}
```

---

## Publishing

### Chrome Web Store

1. **Prepare assets:**
   - Icon: 128x128 PNG
   - Screenshots: 1280x800 or 640x400
   - Promotional images: 440x280 (small), 920x680 (large)

2. **Create developer account:** $5 one-time fee

3. **Package extension:**
   ```bash
   npm run package:chrome
   ```

4. **Submit for review:**
   - Upload ZIP to Chrome Web Store Developer Dashboard
   - Fill in listing details
   - Submit for review (1-3 days typical)

### Firefox Add-ons

1. **Create developer account:** Free

2. **Package extension:**
   ```bash
   npm run package:firefox
   ```

3. **Submit:**
   - Upload to addons.mozilla.org
   - Automated review for most extensions
   - Manual review for extensions requesting sensitive permissions

### Edge Add-ons

1. **Create Microsoft Partner Center account**

2. **Package:** Same as Chrome (Manifest V3 compatible)

3. **Submit through Partner Center**

---

## Testing

```typescript
// src/__tests__/shorten.test.ts
import { describe, it, expect, vi } from 'vitest';
import { shortenUrl } from '../utils/api';

describe('URL Shortening', () => {
  it('should shorten a valid URL', async () => {
    const result = await shortenUrl('https://example.com');
    expect(result.shortUrl).toMatch(/^https:\/\/lnkr\.ft\//);
  });

  it('should reject invalid URLs', async () => {
    await expect(shortenUrl('not-a-url')).rejects.toThrow();
  });
});
```

---

## Troubleshooting

**Extension not loading:**
- Check manifest.json syntax
- Verify all referenced files exist
- Check browser console for errors

**API requests failing:**
- Verify host_permissions in manifest
- Check CORS headers on API
- Ensure authentication token is valid

**Content script not running:**
- Check matches pattern in manifest
- Verify page isn't restricted (chrome://, etc.)
- Check for Content Security Policy issues

---

For more information, visit the [Linkrift Developer Portal](https://developers.linkrift.io).
