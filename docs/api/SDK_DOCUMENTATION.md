# Linkrift SDK Documentation

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [Official SDKs](#official-sdks)
- [Go SDK](#go-sdk)
  - [Installation](#go-installation)
  - [Quick Start](#go-quick-start)
  - [Configuration](#go-configuration)
  - [Links](#go-links)
  - [Domains](#go-domains)
  - [Analytics](#go-analytics)
  - [QR Codes](#go-qr-codes)
  - [Bio Pages](#go-bio-pages)
  - [Workspaces](#go-workspaces)
  - [Webhooks](#go-webhooks)
  - [Error Handling](#go-error-handling)
- [JavaScript/TypeScript SDK](#javascripttypescript-sdk)
  - [Installation](#js-installation)
  - [Quick Start](#js-quick-start)
  - [Configuration](#js-configuration)
  - [Links](#js-links)
  - [Domains](#js-domains)
  - [Analytics](#js-analytics)
  - [QR Codes](#js-qr-codes)
  - [Bio Pages](#js-bio-pages)
  - [Workspaces](#js-workspaces)
  - [Webhooks](#js-webhooks)
  - [Error Handling](#js-error-handling)
- [Python SDK](#python-sdk)
  - [Installation](#python-installation)
  - [Quick Start](#python-quick-start)
  - [Configuration](#python-configuration)
  - [Links](#python-links)
  - [Domains](#python-domains)
  - [Analytics](#python-analytics)
  - [QR Codes](#python-qr-codes)
  - [Bio Pages](#python-bio-pages)
  - [Workspaces](#python-workspaces)
  - [Webhooks](#python-webhooks)
  - [Error Handling](#python-error-handling)
- [Webhook Verification Helpers](#webhook-verification-helpers)
- [Common Patterns](#common-patterns)
- [Migration Guide](#migration-guide)

---

## Overview

Linkrift provides official SDKs for Go, JavaScript/TypeScript, and Python. These SDKs provide idiomatic interfaces for interacting with the Linkrift API, handling authentication, pagination, error handling, and webhook verification.

### SDK Features

- **Type Safety**: Full type definitions for all API objects
- **Automatic Pagination**: Built-in iterators for paginated endpoints
- **Retry Logic**: Automatic retries with exponential backoff
- **Rate Limit Handling**: Automatic rate limit detection and waiting
- **Webhook Verification**: Helper functions for secure webhook processing

---

## Official SDKs

| Language | Package | Repository |
|----------|---------|------------|
| Go | `github.com/link-rift/link-rift/sdk/go` | [GitHub](https://github.com/link-rift/link-rift) |
| JavaScript/TypeScript | `@linkrift/sdk` | [npm](https://www.npmjs.com/package/@linkrift/sdk) |
| Python | `linkrift` | [PyPI](https://pypi.org/project/linkrift/) |

---

## Go SDK

The official Go SDK is part of the main Linkrift repository.

### Go Installation

```bash
go get github.com/link-rift/link-rift/sdk/go
```

### Go Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    // Initialize client with API key
    client := linkrift.NewClient("lr_live_sk_your_api_key")

    // Create a short link
    link, err := client.Links.Create(context.Background(), &linkrift.CreateLinkParams{
        URL:  "https://example.com/very/long/url",
        Slug: "my-link",
        Tags: []string{"marketing"},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Created: %s\n", link.ShortURL)
}
```

### Go Configuration

```go
package main

import (
    "net/http"
    "time"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    // Basic configuration
    client := linkrift.NewClient("lr_live_sk_your_api_key")

    // With options
    client := linkrift.NewClient(
        "lr_live_sk_your_api_key",
        linkrift.WithBaseURL("https://api.linkrift.io"),
        linkrift.WithTimeout(30*time.Second),
        linkrift.WithRetries(3),
        linkrift.WithHTTPClient(&http.Client{
            Timeout: 30 * time.Second,
        }),
    )

    // Using PASETO token instead of API key
    client := linkrift.NewClientWithToken("v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1...")

    // Per-request context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    link, err := client.Links.Get(ctx, "lnk_1234567890abcdef")
}
```

### Go Links

```go
package main

import (
    "context"
    "fmt"
    "time"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Create a link
    link, err := client.Links.Create(ctx, &linkrift.CreateLinkParams{
        URL:         "https://example.com/page",
        Slug:        "custom-slug",
        Title:       "My Link",
        Description: "A description for this link",
        Tags:        []string{"marketing", "campaign"},
        ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 30 days
        Password:    "secret123",
        UTMSource:   "newsletter",
        UTMMedium:   "email",
        UTMCampaign: "january-2025",
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("Created: %s\n", link.ShortURL)

    // Get a link
    link, err = client.Links.Get(ctx, "lnk_1234567890abcdef")
    if err != nil {
        // Handle error
    }

    // Update a link
    link, err = client.Links.Update(ctx, "lnk_1234567890abcdef", &linkrift.UpdateLinkParams{
        Title: "Updated Title",
        Tags:  []string{"updated"},
    })

    // Delete a link
    err = client.Links.Delete(ctx, "lnk_1234567890abcdef")

    // List links with filtering
    links, err := client.Links.List(ctx, &linkrift.ListLinksParams{
        Limit:    20,
        Tag:      "marketing",
        DomainID: "dom_1234567890abcdef",
    })
    for _, link := range links.Data {
        fmt.Printf("- %s: %s\n", link.Slug, link.OriginalURL)
    }

    // Iterate through all links
    iter := client.Links.ListAll(ctx, nil)
    for iter.Next() {
        link := iter.Current()
        fmt.Printf("- %s\n", link.ShortURL)
    }
    if err := iter.Err(); err != nil {
        // Handle error
    }

    // Bulk create links
    results, err := client.Links.BulkCreate(ctx, []linkrift.CreateLinkParams{
        {URL: "https://example.com/page1"},
        {URL: "https://example.com/page2", Slug: "page-2"},
        {URL: "https://example.com/page3"},
    })
    fmt.Printf("Created: %d, Failed: %d\n", results.Summary.Created, results.Summary.Failed)
}
```

### Go Domains

```go
package main

import (
    "context"
    "fmt"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Add a domain
    domain, err := client.Domains.Create(ctx, &linkrift.CreateDomainParams{
        Hostname: "links.mycompany.com",
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("Domain added: %s\n", domain.Hostname)
    fmt.Printf("Verification record: %s -> %s\n",
        domain.Verification.Name,
        domain.Verification.Value)

    // List domains
    domains, err := client.Domains.List(ctx, nil)
    for _, d := range domains.Data {
        fmt.Printf("- %s (verified: %v)\n", d.Hostname, d.Verified)
    }

    // Verify domain
    domain, err = client.Domains.Verify(ctx, "dom_1234567890abcdef")
    if domain.Verified {
        fmt.Println("Domain verified!")
    }

    // Set as default
    err = client.Domains.SetDefault(ctx, "dom_1234567890abcdef")

    // Delete domain
    err = client.Domains.Delete(ctx, "dom_1234567890abcdef")
}
```

### Go Analytics

```go
package main

import (
    "context"
    "fmt"
    "time"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Get link analytics
    startDate := time.Now().Add(-30 * 24 * time.Hour)
    endDate := time.Now()

    analytics, err := client.Analytics.GetLink(ctx, "lnk_1234567890abcdef", &linkrift.AnalyticsParams{
        StartDate: startDate,
        EndDate:   endDate,
        Interval:  "day",
    })
    if err != nil {
        // Handle error
    }

    fmt.Printf("Total clicks: %d\n", analytics.Summary.TotalClicks)
    fmt.Printf("Unique visitors: %d\n", analytics.Summary.UniqueVisitors)

    // Top referrers
    for _, ref := range analytics.TopReferrers {
        fmt.Printf("- %s: %d clicks\n", ref.Referrer, ref.Clicks)
    }

    // Top countries
    for _, country := range analytics.TopCountries {
        fmt.Printf("- %s: %d clicks\n", country.Country, country.Clicks)
    }

    // Get account-wide analytics
    accountAnalytics, err := client.Analytics.GetAccount(ctx, &linkrift.AnalyticsParams{
        StartDate: startDate,
        EndDate:   endDate,
    })

    // Get real-time analytics
    realtime, err := client.Analytics.GetRealtime(ctx)
    fmt.Printf("Active visitors: %d\n", realtime.ActiveVisitors)
    fmt.Printf("Clicks last hour: %d\n", realtime.ClicksLastHour)
}
```

### Go QR Codes

```go
package main

import (
    "context"
    "fmt"
    "os"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Generate QR code
    qr, err := client.QRCodes.Create(ctx, &linkrift.CreateQRCodeParams{
        LinkID:          "lnk_1234567890abcdef",
        Format:          "png",
        Size:            512,
        ForegroundColor: "#000000",
        BackgroundColor: "#FFFFFF",
        ErrorCorrection: "M",
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("QR Code URL: %s\n", qr.URL)

    // Generate QR code with logo
    qr, err = client.QRCodes.Create(ctx, &linkrift.CreateQRCodeParams{
        LinkID:          "lnk_1234567890abcdef",
        Format:          "png",
        Size:            512,
        LogoURL:         "https://example.com/logo.png",
        ErrorCorrection: "H", // Higher error correction for logos
    })

    // Download QR code to file
    data, err := client.QRCodes.Download(ctx, qr.ID, &linkrift.DownloadQRCodeParams{
        Format: "svg",
    })
    if err != nil {
        // Handle error
    }

    err = os.WriteFile("qrcode.svg", data, 0644)
    if err != nil {
        // Handle error
    }

    // List QR codes
    qrcodes, err := client.QRCodes.List(ctx, &linkrift.ListQRCodesParams{
        LinkID: "lnk_1234567890abcdef",
    })
}
```

### Go Bio Pages

```go
package main

import (
    "context"
    "fmt"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Create bio page
    biopage, err := client.BioPages.Create(ctx, &linkrift.CreateBioPageParams{
        Username:  "johndoe",
        Title:     "John Doe",
        Bio:       "Developer & Content Creator",
        AvatarURL: "https://example.com/avatar.jpg",
        Theme:     "dark",
        Links: []linkrift.BioPageLink{
            {Title: "My Website", URL: "https://johndoe.com", Icon: "globe"},
            {Title: "Twitter", URL: "https://twitter.com/johndoe", Icon: "twitter"},
            {Title: "GitHub", URL: "https://github.com/johndoe", Icon: "github"},
        },
        SocialLinks: map[string]string{
            "twitter":  "johndoe",
            "github":   "johndoe",
            "linkedin": "johndoe",
        },
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("Bio page created: %s\n", biopage.URL)

    // Update bio page
    biopage, err = client.BioPages.Update(ctx, biopage.ID, &linkrift.UpdateBioPageParams{
        Bio:   "Updated bio text",
        Theme: "light",
    })

    // Add link to bio page
    err = client.BioPages.AddLink(ctx, biopage.ID, &linkrift.BioPageLink{
        Title:    "New Link",
        URL:      "https://example.com/new",
        Icon:     "link",
        Position: 0,
    })

    // Reorder links
    err = client.BioPages.ReorderLinks(ctx, biopage.ID, []string{
        "bl_link3",
        "bl_link1",
        "bl_link2",
    })

    // Delete bio page
    err = client.BioPages.Delete(ctx, biopage.ID)
}
```

### Go Workspaces

```go
package main

import (
    "context"
    "fmt"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Create workspace
    workspace, err := client.Workspaces.Create(ctx, &linkrift.CreateWorkspaceParams{
        Name: "Marketing Team",
        Slug: "marketing",
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("Workspace created: %s\n", workspace.ID)

    // List workspaces
    workspaces, err := client.Workspaces.List(ctx, nil)
    for _, ws := range workspaces.Data {
        fmt.Printf("- %s (%s)\n", ws.Name, ws.Role)
    }

    // Invite member
    invite, err := client.Workspaces.InviteMember(ctx, workspace.ID, &linkrift.InviteMemberParams{
        Email: "teammate@example.com",
        Role:  linkrift.RoleMember,
    })
    fmt.Printf("Invitation sent: %s\n", invite.ID)

    // List members
    members, err := client.Workspaces.ListMembers(ctx, workspace.ID, nil)
    for _, member := range members.Data {
        fmt.Printf("- %s (%s)\n", member.Name, member.Role)
    }

    // Update member role
    err = client.Workspaces.UpdateMemberRole(ctx, workspace.ID, "usr_member123", linkrift.RoleAdmin)

    // Remove member
    err = client.Workspaces.RemoveMember(ctx, workspace.ID, "usr_member123")

    // Switch workspace context for subsequent requests
    client.SetWorkspace(workspace.ID)

    // All subsequent calls will be in workspace context
    links, err := client.Links.List(ctx, nil)
}
```

### Go Webhooks

```go
package main

import (
    "context"
    "fmt"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    // Create webhook
    webhook, err := client.Webhooks.Create(ctx, &linkrift.CreateWebhookParams{
        URL: "https://example.com/webhooks/linkrift",
        Events: []string{
            linkrift.EventLinkClicked,
            linkrift.EventLinkCreated,
            linkrift.EventLinkUpdated,
        },
        Secret: "whsec_your_webhook_secret",
    })
    if err != nil {
        // Handle error
    }
    fmt.Printf("Webhook created: %s\n", webhook.ID)

    // List webhooks
    webhooks, err := client.Webhooks.List(ctx, nil)
    for _, wh := range webhooks.Data {
        fmt.Printf("- %s: %v\n", wh.URL, wh.Events)
    }

    // Update webhook
    webhook, err = client.Webhooks.Update(ctx, webhook.ID, &linkrift.UpdateWebhookParams{
        Events: []string{linkrift.EventLinkClicked},
        Active: true,
    })

    // Delete webhook
    err = client.Webhooks.Delete(ctx, webhook.ID)
}
```

### Go Error Handling

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func main() {
    client := linkrift.NewClient("lr_live_sk_your_api_key")
    ctx := context.Background()

    link, err := client.Links.Get(ctx, "lnk_nonexistent")
    if err != nil {
        // Check for specific error types
        var apiErr *linkrift.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API Error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
            fmt.Printf("Status: %d\n", apiErr.StatusCode)
            fmt.Printf("Request ID: %s\n", apiErr.RequestID)

            switch apiErr.Code {
            case linkrift.ErrCodeNotFound:
                fmt.Println("Link not found")
            case linkrift.ErrCodeRateLimitExceeded:
                fmt.Printf("Rate limited, retry after: %d seconds\n", apiErr.RetryAfter)
            case linkrift.ErrCodeInvalidParameter:
                fmt.Printf("Invalid parameter: %v\n", apiErr.Details)
            case linkrift.ErrCodeInsufficientPermissions:
                fmt.Println("Insufficient permissions")
            default:
                fmt.Println("Unknown error")
            }
            return
        }

        // Check for network errors
        var netErr *linkrift.NetworkError
        if errors.As(err, &netErr) {
            fmt.Printf("Network error: %s\n", netErr.Message)
            return
        }

        // Unknown error
        log.Fatal(err)
    }

    fmt.Printf("Link: %s\n", link.ShortURL)
}

// Error codes
const (
    ErrCodeInvalidRequest          = "invalid_request"
    ErrCodeInvalidJSON             = "invalid_json"
    ErrCodeMissingParameter        = "missing_parameter"
    ErrCodeInvalidParameter        = "invalid_parameter"
    ErrCodeAuthenticationRequired  = "authentication_required"
    ErrCodeInvalidToken            = "invalid_token"
    ErrCodeInvalidAPIKey           = "invalid_api_key"
    ErrCodeInsufficientPermissions = "insufficient_permissions"
    ErrCodeNotFound                = "resource_not_found"
    ErrCodeSlugAlreadyExists       = "slug_already_exists"
    ErrCodeDomainAlreadyExists     = "domain_already_exists"
    ErrCodeValidationError         = "validation_error"
    ErrCodeURLInvalid              = "url_invalid"
    ErrCodeURLBlocked              = "url_blocked"
    ErrCodeRateLimitExceeded       = "rate_limit_exceeded"
    ErrCodeInternalError           = "internal_error"
    ErrCodeServiceUnavailable      = "service_unavailable"
)
```

---

## JavaScript/TypeScript SDK

### JS Installation

```bash
# npm
npm install @linkrift/sdk

# yarn
yarn add @linkrift/sdk

# pnpm
pnpm add @linkrift/sdk
```

### JS Quick Start

```typescript
import { Linkrift } from '@linkrift/sdk';

// Initialize client
const client = new Linkrift('lr_live_sk_your_api_key');

// Create a short link
const link = await client.links.create({
  url: 'https://example.com/very/long/url',
  slug: 'my-link',
  tags: ['marketing'],
});

console.log(`Created: ${link.shortUrl}`);
```

### JS Configuration

```typescript
import { Linkrift, LinkriftConfig } from '@linkrift/sdk';

// Basic configuration
const client = new Linkrift('lr_live_sk_your_api_key');

// With options
const config: LinkriftConfig = {
  apiKey: 'lr_live_sk_your_api_key',
  baseUrl: 'https://api.linkrift.io',
  timeout: 30000,
  retries: 3,
  debug: false,
};
const client = new Linkrift(config);

// Using PASETO token
const client = new Linkrift({
  token: 'v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1...',
});

// With custom fetch implementation (e.g., for edge runtimes)
const client = new Linkrift({
  apiKey: 'lr_live_sk_your_api_key',
  fetch: customFetchImplementation,
});
```

### JS Links

```typescript
import { Linkrift } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Create a link
const link = await client.links.create({
  url: 'https://example.com/page',
  slug: 'custom-slug',
  title: 'My Link',
  description: 'A description for this link',
  tags: ['marketing', 'campaign'],
  expiresAt: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000), // 30 days
  password: 'secret123',
  utmSource: 'newsletter',
  utmMedium: 'email',
  utmCampaign: 'january-2025',
});
console.log(`Created: ${link.shortUrl}`);

// Get a link
const fetchedLink = await client.links.get('lnk_1234567890abcdef');

// Update a link
const updatedLink = await client.links.update('lnk_1234567890abcdef', {
  title: 'Updated Title',
  tags: ['updated'],
});

// Delete a link
await client.links.delete('lnk_1234567890abcdef');

// List links with filtering
const { data: links, pagination } = await client.links.list({
  limit: 20,
  tag: 'marketing',
  domainId: 'dom_1234567890abcdef',
});

for (const link of links) {
  console.log(`- ${link.slug}: ${link.originalUrl}`);
}

// Iterate through all links with async iterator
for await (const link of client.links.listAll()) {
  console.log(`- ${link.shortUrl}`);
}

// Bulk create links
const results = await client.links.bulkCreate([
  { url: 'https://example.com/page1' },
  { url: 'https://example.com/page2', slug: 'page-2' },
  { url: 'https://example.com/page3' },
]);
console.log(`Created: ${results.summary.created}, Failed: ${results.summary.failed}`);
```

### JS Domains

```typescript
import { Linkrift } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Add a domain
const domain = await client.domains.create({
  hostname: 'links.mycompany.com',
});
console.log(`Domain added: ${domain.hostname}`);
console.log(`Verification: ${domain.verification.name} -> ${domain.verification.value}`);

// List domains
const { data: domains } = await client.domains.list();
for (const d of domains) {
  console.log(`- ${d.hostname} (verified: ${d.verified})`);
}

// Verify domain
const verifiedDomain = await client.domains.verify('dom_1234567890abcdef');
if (verifiedDomain.verified) {
  console.log('Domain verified!');
}

// Set as default
await client.domains.setDefault('dom_1234567890abcdef');

// Delete domain
await client.domains.delete('dom_1234567890abcdef');
```

### JS Analytics

```typescript
import { Linkrift } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Get link analytics
const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
const endDate = new Date();

const analytics = await client.analytics.getLink('lnk_1234567890abcdef', {
  startDate,
  endDate,
  interval: 'day',
});

console.log(`Total clicks: ${analytics.summary.totalClicks}`);
console.log(`Unique visitors: ${analytics.summary.uniqueVisitors}`);

// Top referrers
for (const ref of analytics.topReferrers) {
  console.log(`- ${ref.referrer}: ${ref.clicks} clicks`);
}

// Get account-wide analytics
const accountAnalytics = await client.analytics.getAccount({
  startDate,
  endDate,
});

// Get real-time analytics
const realtime = await client.analytics.getRealtime();
console.log(`Active visitors: ${realtime.activeVisitors}`);
```

### JS QR Codes

```typescript
import { Linkrift } from '@linkrift/sdk';
import { writeFile } from 'fs/promises';

const client = new Linkrift('lr_live_sk_your_api_key');

// Generate QR code
const qr = await client.qrcodes.create({
  linkId: 'lnk_1234567890abcdef',
  format: 'png',
  size: 512,
  foregroundColor: '#000000',
  backgroundColor: '#FFFFFF',
  errorCorrection: 'M',
});
console.log(`QR Code URL: ${qr.url}`);

// Generate QR code with logo
const qrWithLogo = await client.qrcodes.create({
  linkId: 'lnk_1234567890abcdef',
  format: 'png',
  size: 512,
  logoUrl: 'https://example.com/logo.png',
  errorCorrection: 'H',
});

// Download QR code
const data = await client.qrcodes.download(qr.id, { format: 'svg' });
await writeFile('qrcode.svg', data);

// List QR codes
const { data: qrcodes } = await client.qrcodes.list({
  linkId: 'lnk_1234567890abcdef',
});
```

### JS Bio Pages

```typescript
import { Linkrift } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Create bio page
const biopage = await client.biopages.create({
  username: 'johndoe',
  title: 'John Doe',
  bio: 'Developer & Content Creator',
  avatarUrl: 'https://example.com/avatar.jpg',
  theme: 'dark',
  links: [
    { title: 'My Website', url: 'https://johndoe.com', icon: 'globe' },
    { title: 'Twitter', url: 'https://twitter.com/johndoe', icon: 'twitter' },
  ],
  socialLinks: {
    twitter: 'johndoe',
    github: 'johndoe',
  },
});
console.log(`Bio page created: ${biopage.url}`);

// Update bio page
const updated = await client.biopages.update(biopage.id, {
  bio: 'Updated bio text',
  theme: 'light',
});

// Add link
await client.biopages.addLink(biopage.id, {
  title: 'New Link',
  url: 'https://example.com/new',
  icon: 'link',
  position: 0,
});

// Reorder links
await client.biopages.reorderLinks(biopage.id, ['bl_link3', 'bl_link1', 'bl_link2']);

// Delete bio page
await client.biopages.delete(biopage.id);
```

### JS Workspaces

```typescript
import { Linkrift, WorkspaceRole } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Create workspace
const workspace = await client.workspaces.create({
  name: 'Marketing Team',
  slug: 'marketing',
});
console.log(`Workspace created: ${workspace.id}`);

// List workspaces
const { data: workspaces } = await client.workspaces.list();
for (const ws of workspaces) {
  console.log(`- ${ws.name} (${ws.role})`);
}

// Invite member
const invite = await client.workspaces.inviteMember(workspace.id, {
  email: 'teammate@example.com',
  role: WorkspaceRole.Member,
});

// List members
const { data: members } = await client.workspaces.listMembers(workspace.id);

// Update member role
await client.workspaces.updateMemberRole(
  workspace.id,
  'usr_member123',
  WorkspaceRole.Admin
);

// Remove member
await client.workspaces.removeMember(workspace.id, 'usr_member123');

// Set workspace context
client.setWorkspace(workspace.id);

// Subsequent calls use workspace context
const { data: links } = await client.links.list();
```

### JS Webhooks

```typescript
import { Linkrift, WebhookEvent } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

// Create webhook
const webhook = await client.webhooks.create({
  url: 'https://example.com/webhooks/linkrift',
  events: [
    WebhookEvent.LinkClicked,
    WebhookEvent.LinkCreated,
    WebhookEvent.LinkUpdated,
  ],
  secret: 'whsec_your_webhook_secret',
});
console.log(`Webhook created: ${webhook.id}`);

// List webhooks
const { data: webhooks } = await client.webhooks.list();

// Update webhook
await client.webhooks.update(webhook.id, {
  events: [WebhookEvent.LinkClicked],
  active: true,
});

// Delete webhook
await client.webhooks.delete(webhook.id);
```

### JS Error Handling

```typescript
import {
  Linkrift,
  LinkriftError,
  APIError,
  NetworkError,
  ErrorCode,
} from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

try {
  const link = await client.links.get('lnk_nonexistent');
} catch (error) {
  if (error instanceof APIError) {
    console.log(`API Error: ${error.message} (code: ${error.code})`);
    console.log(`Status: ${error.statusCode}`);
    console.log(`Request ID: ${error.requestId}`);

    switch (error.code) {
      case ErrorCode.NotFound:
        console.log('Link not found');
        break;
      case ErrorCode.RateLimitExceeded:
        console.log(`Rate limited, retry after: ${error.retryAfter} seconds`);
        break;
      case ErrorCode.InvalidParameter:
        console.log(`Invalid parameter: ${JSON.stringify(error.details)}`);
        break;
      case ErrorCode.InsufficientPermissions:
        console.log('Insufficient permissions');
        break;
      default:
        console.log('Unknown error');
    }
  } else if (error instanceof NetworkError) {
    console.log(`Network error: ${error.message}`);
  } else {
    throw error;
  }
}

// Error codes enum
enum ErrorCode {
  InvalidRequest = 'invalid_request',
  InvalidJSON = 'invalid_json',
  MissingParameter = 'missing_parameter',
  InvalidParameter = 'invalid_parameter',
  AuthenticationRequired = 'authentication_required',
  InvalidToken = 'invalid_token',
  InvalidAPIKey = 'invalid_api_key',
  InsufficientPermissions = 'insufficient_permissions',
  NotFound = 'resource_not_found',
  SlugAlreadyExists = 'slug_already_exists',
  DomainAlreadyExists = 'domain_already_exists',
  ValidationError = 'validation_error',
  URLInvalid = 'url_invalid',
  URLBlocked = 'url_blocked',
  RateLimitExceeded = 'rate_limit_exceeded',
  InternalError = 'internal_error',
  ServiceUnavailable = 'service_unavailable',
}
```

---

## Python SDK

### Python Installation

```bash
# pip
pip install linkrift

# poetry
poetry add linkrift

# pipenv
pipenv install linkrift
```

### Python Quick Start

```python
from linkrift import Linkrift

# Initialize client
client = Linkrift("lr_live_sk_your_api_key")

# Create a short link
link = client.links.create(
    url="https://example.com/very/long/url",
    slug="my-link",
    tags=["marketing"],
)

print(f"Created: {link.short_url}")
```

### Python Configuration

```python
from linkrift import Linkrift, LinkriftConfig
import httpx

# Basic configuration
client = Linkrift("lr_live_sk_your_api_key")

# With options
config = LinkriftConfig(
    api_key="lr_live_sk_your_api_key",
    base_url="https://api.linkrift.io",
    timeout=30.0,
    retries=3,
    debug=False,
)
client = Linkrift(config)

# Using PASETO token
client = Linkrift(token="v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1...")

# With custom httpx client
http_client = httpx.Client(timeout=30.0)
client = Linkrift(
    api_key="lr_live_sk_your_api_key",
    http_client=http_client,
)

# Async client
from linkrift import AsyncLinkrift

async_client = AsyncLinkrift("lr_live_sk_your_api_key")
link = await async_client.links.create(url="https://example.com")
```

### Python Links

```python
from linkrift import Linkrift
from datetime import datetime, timedelta

client = Linkrift("lr_live_sk_your_api_key")

# Create a link
link = client.links.create(
    url="https://example.com/page",
    slug="custom-slug",
    title="My Link",
    description="A description for this link",
    tags=["marketing", "campaign"],
    expires_at=datetime.now() + timedelta(days=30),
    password="secret123",
    utm_source="newsletter",
    utm_medium="email",
    utm_campaign="january-2025",
)
print(f"Created: {link.short_url}")

# Get a link
fetched_link = client.links.get("lnk_1234567890abcdef")

# Update a link
updated_link = client.links.update(
    "lnk_1234567890abcdef",
    title="Updated Title",
    tags=["updated"],
)

# Delete a link
client.links.delete("lnk_1234567890abcdef")

# List links with filtering
result = client.links.list(
    limit=20,
    tag="marketing",
    domain_id="dom_1234567890abcdef",
)

for link in result.data:
    print(f"- {link.slug}: {link.original_url}")

# Iterate through all links
for link in client.links.list_all():
    print(f"- {link.short_url}")

# Bulk create links
results = client.links.bulk_create([
    {"url": "https://example.com/page1"},
    {"url": "https://example.com/page2", "slug": "page-2"},
    {"url": "https://example.com/page3"},
])
print(f"Created: {results.summary.created}, Failed: {results.summary.failed}")
```

### Python Domains

```python
from linkrift import Linkrift

client = Linkrift("lr_live_sk_your_api_key")

# Add a domain
domain = client.domains.create(hostname="links.mycompany.com")
print(f"Domain added: {domain.hostname}")
print(f"Verification: {domain.verification.name} -> {domain.verification.value}")

# List domains
result = client.domains.list()
for d in result.data:
    print(f"- {d.hostname} (verified: {d.verified})")

# Verify domain
verified_domain = client.domains.verify("dom_1234567890abcdef")
if verified_domain.verified:
    print("Domain verified!")

# Set as default
client.domains.set_default("dom_1234567890abcdef")

# Delete domain
client.domains.delete("dom_1234567890abcdef")
```

### Python Analytics

```python
from linkrift import Linkrift
from datetime import datetime, timedelta

client = Linkrift("lr_live_sk_your_api_key")

# Get link analytics
start_date = datetime.now() - timedelta(days=30)
end_date = datetime.now()

analytics = client.analytics.get_link(
    "lnk_1234567890abcdef",
    start_date=start_date,
    end_date=end_date,
    interval="day",
)

print(f"Total clicks: {analytics.summary.total_clicks}")
print(f"Unique visitors: {analytics.summary.unique_visitors}")

# Top referrers
for ref in analytics.top_referrers:
    print(f"- {ref.referrer}: {ref.clicks} clicks")

# Top countries
for country in analytics.top_countries:
    print(f"- {country.country}: {country.clicks} clicks")

# Get account-wide analytics
account_analytics = client.analytics.get_account(
    start_date=start_date,
    end_date=end_date,
)

# Get real-time analytics
realtime = client.analytics.get_realtime()
print(f"Active visitors: {realtime.active_visitors}")
print(f"Clicks last hour: {realtime.clicks_last_hour}")
```

### Python QR Codes

```python
from linkrift import Linkrift

client = Linkrift("lr_live_sk_your_api_key")

# Generate QR code
qr = client.qrcodes.create(
    link_id="lnk_1234567890abcdef",
    format="png",
    size=512,
    foreground_color="#000000",
    background_color="#FFFFFF",
    error_correction="M",
)
print(f"QR Code URL: {qr.url}")

# Generate QR code with logo
qr_with_logo = client.qrcodes.create(
    link_id="lnk_1234567890abcdef",
    format="png",
    size=512,
    logo_url="https://example.com/logo.png",
    error_correction="H",
)

# Download QR code
data = client.qrcodes.download(qr.id, format="svg")
with open("qrcode.svg", "wb") as f:
    f.write(data)

# List QR codes
result = client.qrcodes.list(link_id="lnk_1234567890abcdef")
```

### Python Bio Pages

```python
from linkrift import Linkrift

client = Linkrift("lr_live_sk_your_api_key")

# Create bio page
biopage = client.biopages.create(
    username="johndoe",
    title="John Doe",
    bio="Developer & Content Creator",
    avatar_url="https://example.com/avatar.jpg",
    theme="dark",
    links=[
        {"title": "My Website", "url": "https://johndoe.com", "icon": "globe"},
        {"title": "Twitter", "url": "https://twitter.com/johndoe", "icon": "twitter"},
    ],
    social_links={
        "twitter": "johndoe",
        "github": "johndoe",
    },
)
print(f"Bio page created: {biopage.url}")

# Update bio page
updated = client.biopages.update(
    biopage.id,
    bio="Updated bio text",
    theme="light",
)

# Add link
client.biopages.add_link(
    biopage.id,
    title="New Link",
    url="https://example.com/new",
    icon="link",
    position=0,
)

# Reorder links
client.biopages.reorder_links(biopage.id, ["bl_link3", "bl_link1", "bl_link2"])

# Delete bio page
client.biopages.delete(biopage.id)
```

### Python Workspaces

```python
from linkrift import Linkrift, WorkspaceRole

client = Linkrift("lr_live_sk_your_api_key")

# Create workspace
workspace = client.workspaces.create(
    name="Marketing Team",
    slug="marketing",
)
print(f"Workspace created: {workspace.id}")

# List workspaces
result = client.workspaces.list()
for ws in result.data:
    print(f"- {ws.name} ({ws.role})")

# Invite member
invite = client.workspaces.invite_member(
    workspace.id,
    email="teammate@example.com",
    role=WorkspaceRole.MEMBER,
)

# List members
members = client.workspaces.list_members(workspace.id)

# Update member role
client.workspaces.update_member_role(
    workspace.id,
    "usr_member123",
    WorkspaceRole.ADMIN,
)

# Remove member
client.workspaces.remove_member(workspace.id, "usr_member123")

# Set workspace context
client.set_workspace(workspace.id)

# Subsequent calls use workspace context
result = client.links.list()
```

### Python Webhooks

```python
from linkrift import Linkrift, WebhookEvent

client = Linkrift("lr_live_sk_your_api_key")

# Create webhook
webhook = client.webhooks.create(
    url="https://example.com/webhooks/linkrift",
    events=[
        WebhookEvent.LINK_CLICKED,
        WebhookEvent.LINK_CREATED,
        WebhookEvent.LINK_UPDATED,
    ],
    secret="whsec_your_webhook_secret",
)
print(f"Webhook created: {webhook.id}")

# List webhooks
result = client.webhooks.list()

# Update webhook
client.webhooks.update(
    webhook.id,
    events=[WebhookEvent.LINK_CLICKED],
    active=True,
)

# Delete webhook
client.webhooks.delete(webhook.id)
```

### Python Error Handling

```python
from linkrift import (
    Linkrift,
    LinkriftError,
    APIError,
    NetworkError,
    ErrorCode,
)

client = Linkrift("lr_live_sk_your_api_key")

try:
    link = client.links.get("lnk_nonexistent")
except APIError as e:
    print(f"API Error: {e.message} (code: {e.code})")
    print(f"Status: {e.status_code}")
    print(f"Request ID: {e.request_id}")

    if e.code == ErrorCode.NOT_FOUND:
        print("Link not found")
    elif e.code == ErrorCode.RATE_LIMIT_EXCEEDED:
        print(f"Rate limited, retry after: {e.retry_after} seconds")
    elif e.code == ErrorCode.INVALID_PARAMETER:
        print(f"Invalid parameter: {e.details}")
    elif e.code == ErrorCode.INSUFFICIENT_PERMISSIONS:
        print("Insufficient permissions")
    else:
        print("Unknown error")

except NetworkError as e:
    print(f"Network error: {e.message}")

except LinkriftError as e:
    print(f"Linkrift error: {e}")


# Error codes enum
class ErrorCode:
    INVALID_REQUEST = "invalid_request"
    INVALID_JSON = "invalid_json"
    MISSING_PARAMETER = "missing_parameter"
    INVALID_PARAMETER = "invalid_parameter"
    AUTHENTICATION_REQUIRED = "authentication_required"
    INVALID_TOKEN = "invalid_token"
    INVALID_API_KEY = "invalid_api_key"
    INSUFFICIENT_PERMISSIONS = "insufficient_permissions"
    NOT_FOUND = "resource_not_found"
    SLUG_ALREADY_EXISTS = "slug_already_exists"
    DOMAIN_ALREADY_EXISTS = "domain_already_exists"
    VALIDATION_ERROR = "validation_error"
    URL_INVALID = "url_invalid"
    URL_BLOCKED = "url_blocked"
    RATE_LIMIT_EXCEEDED = "rate_limit_exceeded"
    INTERNAL_ERROR = "internal_error"
    SERVICE_UNAVAILABLE = "service_unavailable"
```

---

## Webhook Verification Helpers

All SDKs include helpers for verifying webhook signatures.

### Go Webhook Verification

```go
package main

import (
    "io"
    "net/http"

    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read body", http.StatusBadRequest)
        return
    }

    signature := r.Header.Get("X-Linkrift-Signature")
    secret := "whsec_your_webhook_secret"

    // Verify signature
    if !linkrift.VerifyWebhookSignature(payload, signature, secret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Parse the event
    event, err := linkrift.ParseWebhookEvent(payload)
    if err != nil {
        http.Error(w, "Failed to parse event", http.StatusBadRequest)
        return
    }

    // Handle the event
    switch event.Type {
    case linkrift.EventLinkClicked:
        // Handle link clicked
        data := event.Data.(*linkrift.LinkClickedData)
        fmt.Printf("Link %s clicked from %s\n", data.LinkID, data.Click.Country)
    case linkrift.EventLinkCreated:
        // Handle link created
        data := event.Data.(*linkrift.LinkCreatedData)
        fmt.Printf("Link created: %s\n", data.ShortURL)
    }

    w.WriteHeader(http.StatusOK)
}
```

### JavaScript/TypeScript Webhook Verification

```typescript
import { verifyWebhookSignature, parseWebhookEvent, WebhookEvent } from '@linkrift/sdk';
import express from 'express';

const app = express();

app.post('/webhooks/linkrift', express.raw({ type: 'application/json' }), (req, res) => {
  const payload = req.body;
  const signature = req.headers['x-linkrift-signature'] as string;
  const secret = 'whsec_your_webhook_secret';

  // Verify signature
  if (!verifyWebhookSignature(payload, signature, secret)) {
    return res.status(401).send('Invalid signature');
  }

  // Parse the event
  const event = parseWebhookEvent(payload);

  // Handle the event
  switch (event.type) {
    case WebhookEvent.LinkClicked:
      console.log(`Link ${event.data.linkId} clicked from ${event.data.click.country}`);
      break;
    case WebhookEvent.LinkCreated:
      console.log(`Link created: ${event.data.shortUrl}`);
      break;
  }

  res.status(200).send('OK');
});

// With the Webhook helper class
import { WebhookHandler } from '@linkrift/sdk';

const webhookHandler = new WebhookHandler('whsec_your_webhook_secret');

app.post('/webhooks/linkrift', express.raw({ type: 'application/json' }), (req, res) => {
  try {
    const event = webhookHandler.constructEvent(
      req.body,
      req.headers['x-linkrift-signature'] as string
    );

    // Handle event...

    res.status(200).send('OK');
  } catch (err) {
    res.status(400).send(`Webhook error: ${err.message}`);
  }
});
```

### Python Webhook Verification

```python
from flask import Flask, request
from linkrift import verify_webhook_signature, parse_webhook_event, WebhookEvent
from linkrift.webhooks import WebhookHandler

app = Flask(__name__)

@app.route('/webhooks/linkrift', methods=['POST'])
def webhook_handler():
    payload = request.data
    signature = request.headers.get('X-Linkrift-Signature')
    secret = 'whsec_your_webhook_secret'

    # Verify signature
    if not verify_webhook_signature(payload, signature, secret):
        return 'Invalid signature', 401

    # Parse the event
    event = parse_webhook_event(payload)

    # Handle the event
    if event.type == WebhookEvent.LINK_CLICKED:
        print(f"Link {event.data.link_id} clicked from {event.data.click.country}")
    elif event.type == WebhookEvent.LINK_CREATED:
        print(f"Link created: {event.data.short_url}")

    return 'OK', 200


# With the WebhookHandler class
webhook_handler = WebhookHandler('whsec_your_webhook_secret')

@app.route('/webhooks/linkrift', methods=['POST'])
def handle_webhook():
    try:
        event = webhook_handler.construct_event(
            request.data,
            request.headers.get('X-Linkrift-Signature')
        )

        # Handle event...

        return 'OK', 200
    except Exception as e:
        return f'Webhook error: {str(e)}', 400
```

---

## Common Patterns

### Retry with Exponential Backoff

All SDKs automatically implement retry logic with exponential backoff, but you can also implement custom retry logic:

```go
// Go
import (
    "time"
    linkrift "github.com/link-rift/link-rift/sdk/go"
)

func createLinkWithRetry(client *linkrift.Client, params *linkrift.CreateLinkParams) (*linkrift.Link, error) {
    var lastErr error
    for i := 0; i < 3; i++ {
        link, err := client.Links.Create(context.Background(), params)
        if err == nil {
            return link, nil
        }

        var apiErr *linkrift.APIError
        if errors.As(err, &apiErr) && apiErr.Code == linkrift.ErrCodeRateLimitExceeded {
            time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
            continue
        }

        lastErr = err
        time.Sleep(time.Duration(1<<i) * time.Second) // Exponential backoff
    }
    return nil, lastErr
}
```

### Batch Processing

```typescript
// TypeScript
import { Linkrift } from '@linkrift/sdk';

const client = new Linkrift('lr_live_sk_your_api_key');

async function processLinks(urls: string[]) {
  const batchSize = 100;
  const results = [];

  for (let i = 0; i < urls.length; i += batchSize) {
    const batch = urls.slice(i, i + batchSize).map(url => ({ url }));
    const result = await client.links.bulkCreate(batch);
    results.push(...result.data);

    // Respect rate limits
    if (i + batchSize < urls.length) {
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }

  return results;
}
```

### Context Management

```python
# Python - using context manager
from linkrift import Linkrift

with Linkrift("lr_live_sk_your_api_key") as client:
    link = client.links.create(url="https://example.com")
    print(link.short_url)
# Client is automatically closed


# Async context manager
from linkrift import AsyncLinkrift

async def main():
    async with AsyncLinkrift("lr_live_sk_your_api_key") as client:
        link = await client.links.create(url="https://example.com")
        print(link.short_url)
```

---

## Migration Guide

### Migrating from v0.x to v1.x

The v1.x release includes breaking changes. Here's how to migrate:

#### Method Naming Changes

```go
// Before (v0.x)
client.CreateLink(ctx, params)
client.GetLink(ctx, id)

// After (v1.x)
client.Links.Create(ctx, params)
client.Links.Get(ctx, id)
```

#### Error Handling Changes

```typescript
// Before (v0.x)
try {
  const link = await client.getLink('lnk_123');
} catch (error) {
  if (error.code === 'not_found') { ... }
}

// After (v1.x)
try {
  const link = await client.links.get('lnk_123');
} catch (error) {
  if (error instanceof APIError && error.code === ErrorCode.NotFound) { ... }
}
```

#### Configuration Changes

```python
# Before (v0.x)
client = Linkrift(api_key="...", timeout=30)

# After (v1.x)
config = LinkriftConfig(api_key="...", timeout=30.0)
client = Linkrift(config)
# Or simply:
client = Linkrift("lr_live_sk_your_api_key")
```

---

## Additional Resources

- **API Documentation**: [docs.linkrift.io/api](https://docs.linkrift.io/api)
- **GitHub Repository**: [github.com/link-rift/link-rift](https://github.com/link-rift/link-rift)
- **Changelog**: [docs.linkrift.io/changelog](https://docs.linkrift.io/changelog)
- **Support**: [support@linkrift.io](mailto:support@linkrift.io)

---

*This documentation is generated from the Linkrift SDK specifications. For the most up-to-date information, visit [docs.linkrift.io/sdk](https://docs.linkrift.io/sdk).*
