# Internationalization (i18n)

> Last Updated: 2025-01-24

Guide for implementing and managing internationalization in Linkrift using react-i18next.

---

## Table of Contents

- [Overview](#overview)
- [Setup](#setup)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Project Structure](#project-structure)
- [Basic Usage](#basic-usage)
  - [Using Translations](#using-translations)
  - [Interpolation](#interpolation)
  - [Pluralization](#pluralization)
  - [Formatting](#formatting)
- [Adding New Languages](#adding-new-languages)
- [Translation Workflow](#translation-workflow)
  - [Extracting Strings](#extracting-strings)
  - [Translation Management](#translation-management)
  - [Quality Assurance](#quality-assurance)
- [Advanced Features](#advanced-features)
  - [Namespaces](#namespaces)
  - [Lazy Loading](#lazy-loading)
  - [Context](#context)
  - [Trans Component](#trans-component)
- [Best Practices](#best-practices)
- [RTL Support](#rtl-support)
- [Backend i18n](#backend-i18n)
- [Testing](#testing)

---

## Overview

Linkrift supports multiple languages through react-i18next, providing:

- Automatic language detection
- Lazy loading of translations
- Interpolation and pluralization
- Number, date, and currency formatting
- RTL language support

**Supported Languages:**

| Language | Code | Status |
|----------|------|--------|
| English | en | Complete |
| Spanish | es | Complete |
| French | fr | Complete |
| German | de | Complete |
| Portuguese | pt-BR | Complete |
| Japanese | ja | Complete |
| Chinese (Simplified) | zh-CN | In Progress |
| Arabic | ar | In Progress |
| Russian | ru | Planned |
| Korean | ko | Planned |

---

## Setup

### Installation

```bash
npm install i18next react-i18next i18next-browser-languagedetector i18next-http-backend
```

### Configuration

```typescript
// src/i18n/config.ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import HttpBackend from 'i18next-http-backend';

export const supportedLanguages = {
  en: { name: 'English', nativeName: 'English', dir: 'ltr' },
  es: { name: 'Spanish', nativeName: 'Español', dir: 'ltr' },
  fr: { name: 'French', nativeName: 'Français', dir: 'ltr' },
  de: { name: 'German', nativeName: 'Deutsch', dir: 'ltr' },
  'pt-BR': { name: 'Portuguese (Brazil)', nativeName: 'Português', dir: 'ltr' },
  ja: { name: 'Japanese', nativeName: '日本語', dir: 'ltr' },
  'zh-CN': { name: 'Chinese (Simplified)', nativeName: '简体中文', dir: 'ltr' },
  ar: { name: 'Arabic', nativeName: 'العربية', dir: 'rtl' },
};

i18n
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    // Debug mode in development
    debug: import.meta.env.DEV,

    // Default and fallback language
    fallbackLng: 'en',
    supportedLngs: Object.keys(supportedLanguages),

    // Language detection options
    detection: {
      order: ['querystring', 'localStorage', 'navigator', 'htmlTag'],
      lookupQuerystring: 'lang',
      lookupLocalStorage: 'linkrift-language',
      caches: ['localStorage'],
    },

    // Backend options for loading translations
    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json',
    },

    // Namespaces
    ns: ['common', 'dashboard', 'analytics', 'auth', 'errors'],
    defaultNS: 'common',

    // Interpolation options
    interpolation: {
      escapeValue: false, // React already escapes
      formatSeparator: ',',
    },

    // React options
    react: {
      useSuspense: true,
    },
  });

export default i18n;
```

```tsx
// src/main.tsx
import React, { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './i18n/config';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Suspense fallback={<LoadingSpinner />}>
      <App />
    </Suspense>
  </React.StrictMode>
);
```

### Project Structure

```
src/
├── i18n/
│   ├── config.ts           # i18n configuration
│   └── utils.ts            # i18n utilities
├── locales/                # Or public/locales/
│   ├── en/
│   │   ├── common.json     # Common translations
│   │   ├── dashboard.json  # Dashboard-specific
│   │   ├── analytics.json  # Analytics page
│   │   ├── auth.json       # Authentication
│   │   └── errors.json     # Error messages
│   ├── es/
│   │   ├── common.json
│   │   └── ...
│   └── ...
└── components/
    └── LanguageSwitcher.tsx
```

---

## Basic Usage

### Using Translations

```tsx
// Using the hook
import { useTranslation } from 'react-i18next';

function Dashboard() {
  const { t } = useTranslation();

  return (
    <div>
      <h1>{t('dashboard.title')}</h1>
      <p>{t('dashboard.welcome')}</p>
    </div>
  );
}

// Using with specific namespace
function Analytics() {
  const { t } = useTranslation('analytics');

  return (
    <div>
      <h1>{t('title')}</h1>
      <p>{t('description')}</p>
    </div>
  );
}

// Using multiple namespaces
function Settings() {
  const { t } = useTranslation(['settings', 'common']);

  return (
    <div>
      <h1>{t('settings:title')}</h1>
      <button>{t('common:save')}</button>
    </div>
  );
}
```

**Translation file example:**

```json
// locales/en/common.json
{
  "appName": "Linkrift",
  "navigation": {
    "dashboard": "Dashboard",
    "analytics": "Analytics",
    "settings": "Settings",
    "logout": "Log Out"
  },
  "actions": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "edit": "Edit",
    "copy": "Copy",
    "share": "Share"
  },
  "messages": {
    "success": "Success!",
    "error": "An error occurred",
    "loading": "Loading..."
  }
}
```

### Interpolation

```tsx
// Simple interpolation
t('greeting', { name: 'John' })
// "Hello, {{name}}!" -> "Hello, John!"

// Nested interpolation
t('stats.clicks', { count: link.clicks, name: link.shortCode })
// "{{name}} has {{count}} clicks" -> "abc123 has 42 clicks"
```

```json
// locales/en/common.json
{
  "greeting": "Hello, {{name}}!",
  "stats": {
    "clicks": "{{name}} has {{count}} clicks"
  }
}
```

### Pluralization

```tsx
// Pluralization
t('links.count', { count: links.length })
// count: 0 -> "No links"
// count: 1 -> "1 link"
// count: 5 -> "5 links"
```

```json
// locales/en/common.json
{
  "links": {
    "count_zero": "No links",
    "count_one": "{{count}} link",
    "count_other": "{{count}} links"
  }
}

// For languages with different plural rules (e.g., Russian)
// locales/ru/common.json
{
  "links": {
    "count_zero": "Нет ссылок",
    "count_one": "{{count}} ссылка",
    "count_few": "{{count}} ссылки",
    "count_many": "{{count}} ссылок",
    "count_other": "{{count}} ссылок"
  }
}
```

### Formatting

```tsx
import { useTranslation } from 'react-i18next';

function Analytics() {
  const { t, i18n } = useTranslation();

  // Number formatting
  const formattedNumber = new Intl.NumberFormat(i18n.language).format(1234567);
  // en: "1,234,567"
  // de: "1.234.567"

  // Date formatting
  const formattedDate = new Intl.DateTimeFormat(i18n.language, {
    dateStyle: 'long',
    timeStyle: 'short',
  }).format(new Date());
  // en: "January 24, 2025 at 10:30 AM"
  // de: "24. Januar 2025 um 10:30"

  // Currency formatting
  const formattedCurrency = new Intl.NumberFormat(i18n.language, {
    style: 'currency',
    currency: 'USD',
  }).format(99.99);
  // en: "$99.99"
  // de: "99,99 $"

  return (
    <div>
      <p>{t('analytics.totalClicks', { count: formattedNumber })}</p>
      <p>{t('analytics.lastUpdated', { date: formattedDate })}</p>
    </div>
  );
}
```

---

## Adding New Languages

### Step 1: Create Translation Files

```bash
# Create directory for new language
mkdir -p public/locales/ko

# Copy English files as template
cp public/locales/en/*.json public/locales/ko/
```

### Step 2: Update Configuration

```typescript
// src/i18n/config.ts
export const supportedLanguages = {
  // ... existing languages
  ko: { name: 'Korean', nativeName: '한국어', dir: 'ltr' },
};
```

### Step 3: Translate Content

```json
// public/locales/ko/common.json
{
  "appName": "Linkrift",
  "navigation": {
    "dashboard": "대시보드",
    "analytics": "분석",
    "settings": "설정",
    "logout": "로그아웃"
  },
  "actions": {
    "save": "저장",
    "cancel": "취소",
    "delete": "삭제",
    "edit": "편집",
    "copy": "복사",
    "share": "공유"
  }
}
```

### Step 4: Test

```bash
# Run app with new language
npm run dev

# Navigate to http://localhost:5173?lang=ko
```

---

## Translation Workflow

### Extracting Strings

Use i18next-parser to extract translation keys:

```bash
npm install -D i18next-parser
```

```javascript
// i18next-parser.config.js
module.exports = {
  locales: ['en', 'es', 'fr', 'de', 'pt-BR', 'ja', 'zh-CN', 'ar'],
  output: 'public/locales/$LOCALE/$NAMESPACE.json',
  input: ['src/**/*.{ts,tsx}'],
  defaultNamespace: 'common',
  keySeparator: '.',
  namespaceSeparator: ':',
  createOldCatalogs: false,
  failOnWarnings: false,
  verbose: true,
};
```

```bash
# Extract strings
npx i18next-parser
```

### Translation Management

**Option 1: Crowdin**

```yaml
# crowdin.yml
project_id: 'linkrift'
api_token_env: 'CROWDIN_TOKEN'
base_path: '.'
base_url: 'https://api.crowdin.com'

preserve_hierarchy: true

files:
  - source: '/public/locales/en/*.json'
    translation: '/public/locales/%locale%/%original_file_name%'
```

**Option 2: Lokalise**

```bash
# Push source strings
lokalise2 file upload \
  --project-id $LOKALISE_PROJECT_ID \
  --file public/locales/en/common.json \
  --lang-iso en

# Pull translations
lokalise2 file download \
  --project-id $LOKALISE_PROJECT_ID \
  --format json \
  --dest public/locales
```

**Option 3: Self-Hosted (tolgee)**

```yaml
# docker-compose.yml
services:
  tolgee:
    image: tolgee/tolgee
    ports:
      - "8085:8080"
    environment:
      - SPRING_DATASOURCE_URL=jdbc:postgresql://postgres:5432/tolgee
    volumes:
      - tolgee_data:/data
```

### Quality Assurance

```typescript
// scripts/validate-translations.ts
import fs from 'fs';
import path from 'path';

const localesDir = path.join(__dirname, '../public/locales');
const sourceLocale = 'en';

function validateTranslations() {
  const sourceFiles = fs.readdirSync(path.join(localesDir, sourceLocale));
  const locales = fs.readdirSync(localesDir).filter((d) => d !== sourceLocale);

  const errors: string[] = [];

  for (const file of sourceFiles) {
    const sourceContent = JSON.parse(
      fs.readFileSync(path.join(localesDir, sourceLocale, file), 'utf-8')
    );
    const sourceKeys = getAllKeys(sourceContent);

    for (const locale of locales) {
      const localePath = path.join(localesDir, locale, file);

      if (!fs.existsSync(localePath)) {
        errors.push(`Missing file: ${locale}/${file}`);
        continue;
      }

      const localeContent = JSON.parse(fs.readFileSync(localePath, 'utf-8'));
      const localeKeys = getAllKeys(localeContent);

      // Check for missing keys
      for (const key of sourceKeys) {
        if (!localeKeys.includes(key)) {
          errors.push(`Missing key in ${locale}/${file}: ${key}`);
        }
      }

      // Check for extra keys
      for (const key of localeKeys) {
        if (!sourceKeys.includes(key)) {
          errors.push(`Extra key in ${locale}/${file}: ${key}`);
        }
      }
    }
  }

  if (errors.length > 0) {
    console.error('Translation validation errors:');
    errors.forEach((e) => console.error(`  - ${e}`));
    process.exit(1);
  }

  console.log('All translations valid!');
}

function getAllKeys(obj: any, prefix = ''): string[] {
  const keys: string[] = [];
  for (const key of Object.keys(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof obj[key] === 'object' && obj[key] !== null) {
      keys.push(...getAllKeys(obj[key], fullKey));
    } else {
      keys.push(fullKey);
    }
  }
  return keys;
}

validateTranslations();
```

---

## Advanced Features

### Namespaces

Organize translations into logical groups:

```typescript
// Load specific namespaces
const { t } = useTranslation(['dashboard', 'common']);

// Access namespaced translations
t('dashboard:title')
t('common:actions.save')
```

### Lazy Loading

Load translations on demand:

```typescript
// i18n config
backend: {
  loadPath: '/locales/{{lng}}/{{ns}}.json',
},

// Load namespace when needed
const { t } = useTranslation('analytics', { useSuspense: true });
```

### Context

Handle gendered or contextual translations:

```json
{
  "friend": "A friend",
  "friend_male": "A boyfriend",
  "friend_female": "A girlfriend"
}
```

```tsx
t('friend', { context: 'male' }) // "A boyfriend"
t('friend', { context: 'female' }) // "A girlfriend"
```

### Trans Component

For complex translations with React components:

```tsx
import { Trans } from 'react-i18next';

function Welcome() {
  return (
    <Trans i18nKey="welcome">
      Welcome to <strong>Linkrift</strong>. Click <a href="/docs">here</a> to
      learn more.
    </Trans>
  );
}
```

```json
{
  "welcome": "Welcome to <1>Linkrift</1>. Click <3>here</3> to learn more."
}
```

---

## Best Practices

### 1. Use Meaningful Keys

```json
// Bad
{
  "text1": "Welcome",
  "btn1": "Save"
}

// Good
{
  "dashboard": {
    "welcome": "Welcome",
    "actions": {
      "save": "Save"
    }
  }
}
```

### 2. Keep Translations DRY

```json
// Common terms in common namespace
{
  "common": {
    "save": "Save",
    "cancel": "Cancel"
  }
}

// Reuse in components
t('common:save')
```

### 3. Handle Missing Translations

```typescript
i18n.init({
  missingKeyHandler: (lng, ns, key) => {
    console.warn(`Missing translation: ${lng}/${ns}/${key}`);
    // Report to error tracking service
  },
  saveMissing: true,
});
```

### 4. Avoid String Concatenation

```tsx
// Bad
t('hello') + ' ' + name + '!'

// Good
t('greeting', { name })
```

---

## RTL Support

For Arabic, Hebrew, and other RTL languages:

```tsx
// src/hooks/useDirection.ts
import { useTranslation } from 'react-i18next';
import { supportedLanguages } from '../i18n/config';

export function useDirection() {
  const { i18n } = useTranslation();
  const lang = supportedLanguages[i18n.language as keyof typeof supportedLanguages];
  return lang?.dir || 'ltr';
}

// App.tsx
function App() {
  const dir = useDirection();

  return (
    <div dir={dir} className={dir === 'rtl' ? 'rtl' : 'ltr'}>
      {/* App content */}
    </div>
  );
}
```

```css
/* RTL-specific styles */
[dir='rtl'] .sidebar {
  right: 0;
  left: auto;
}

[dir='rtl'] .icon-arrow {
  transform: scaleX(-1);
}

/* Use logical properties when possible */
.sidebar {
  padding-inline-start: 1rem;
  margin-inline-end: 2rem;
}
```

---

## Backend i18n

For API error messages and emails:

```go
// internal/i18n/i18n.go
package i18n

import (
    "embed"
    "encoding/json"
)

//go:embed locales/*.json
var localesFS embed.FS

type Translator struct {
    translations map[string]map[string]string
    defaultLang  string
}

func NewTranslator(defaultLang string) *Translator {
    t := &Translator{
        translations: make(map[string]map[string]string),
        defaultLang:  defaultLang,
    }
    t.loadTranslations()
    return t
}

func (t *Translator) T(lang, key string) string {
    if translations, ok := t.translations[lang]; ok {
        if value, ok := translations[key]; ok {
            return value
        }
    }
    // Fallback to default language
    if translations, ok := t.translations[t.defaultLang]; ok {
        if value, ok := translations[key]; ok {
            return value
        }
    }
    return key
}
```

---

## Testing

```tsx
// src/test/i18n-test-utils.tsx
import { I18nextProvider } from 'react-i18next';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Create test i18n instance
const testI18n = i18n.createInstance();
testI18n.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  resources: {
    en: {
      common: {
        save: 'Save',
        cancel: 'Cancel',
      },
    },
  },
});

export function renderWithI18n(component: React.ReactElement) {
  return render(
    <I18nextProvider i18n={testI18n}>{component}</I18nextProvider>
  );
}

// Usage in tests
test('renders save button', () => {
  renderWithI18n(<SaveButton />);
  expect(screen.getByText('Save')).toBeInTheDocument();
});
```

---

## Summary

Key points for i18n in Linkrift:

1. Use react-i18next for comprehensive translation support
2. Organize translations by namespace
3. Use interpolation and pluralization properly
4. Set up automated extraction and validation
5. Support RTL languages with proper CSS
6. Test translations as part of CI/CD

For translation contributions, see [CONTRIBUTING.md](../contributing/CONTRIBUTING.md).
