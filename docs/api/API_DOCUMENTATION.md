# Linkrift API Documentation

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [Base URL and Versioning](#base-url-and-versioning)
- [Authentication](#authentication)
  - [PASETO Tokens](#paseto-tokens)
  - [API Keys](#api-keys)
- [Request/Response Conventions](#requestresponse-conventions)
- [Pagination](#pagination)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication-endpoints)
  - [Links](#links)
  - [Domains](#domains)
  - [Analytics](#analytics)
  - [QR Codes](#qr-codes)
  - [Bio Pages](#bio-pages)
  - [Workspaces](#workspaces)
  - [Webhooks](#webhooks)

---

## Overview

The Linkrift API is a RESTful API that allows you to programmatically create and manage short links, custom domains, analytics, QR codes, bio pages, and more. All API endpoints return JSON responses and accept JSON request bodies where applicable.

### Key Features

- **URL Shortening**: Create, update, and manage short links
- **Custom Domains**: Use your own branded domains
- **Analytics**: Track clicks, geographic data, and device information
- **QR Codes**: Generate customizable QR codes for your links
- **Bio Pages**: Create link-in-bio landing pages
- **Workspaces**: Collaborate with team members
- **Webhooks**: Receive real-time notifications for events

---

## Base URL and Versioning

### Base URL

```
https://api.linkrift.io/v1
```

### API Versioning

The API uses URL-based versioning. The current version is `v1`. All endpoints are prefixed with the version number.

```
https://api.linkrift.io/v1/links
https://api.linkrift.io/v1/domains
https://api.linkrift.io/v1/analytics
```

When a new API version is released, the previous version will be supported for at least 12 months with deprecation notices.

---

## Authentication

Linkrift supports two authentication methods: PASETO tokens for user sessions and API keys for programmatic access.

### PASETO Tokens

PASETO (Platform-Agnostic Security Tokens) are used for user authentication after login. They are more secure than JWTs and provide built-in protection against common vulnerabilities.

#### Obtaining a Token

```bash
curl -X POST https://api.linkrift.io/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your-password"
  }'
```

**Response:**

```json
{
  "access_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJleHAiOiIyMDI1LTAxLTI1VDAwOjAwOjAwWiIsImlhdCI6IjIwMjUtMDEtMjRUMDA6MDA6MDBaIn0...",
  "refresh_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJ0eXBlIjoicmVmcmVzaCIsImV4cCI6IjIwMjUtMDItMjRUMDA6MDA6MDBaIn0...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

#### Using the Token

Include the token in the `Authorization` header:

```bash
curl https://api.linkrift.io/v1/links \
  -H "Authorization: Bearer v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAi..."
```

#### Token Refresh

```bash
curl -X POST https://api.linkrift.io/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJ0eXBlIjoicmVmcmVzaCIs..."
  }'
```

### API Keys

API keys are recommended for server-to-server communication and automated workflows. They do not expire but can be revoked at any time.

#### Creating an API Key

API keys can be created in the Linkrift dashboard under **Settings > API Keys** or via the API:

```bash
curl -X POST https://api.linkrift.io/v1/api-keys \
  -H "Authorization: Bearer {access_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Server",
    "scopes": ["links:read", "links:write", "analytics:read"]
  }'
```

**Response:**

```json
{
  "id": "ak_1234567890abcdef",
  "name": "Production Server",
  "key": "lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz",
  "scopes": ["links:read", "links:write", "analytics:read"],
  "created_at": "2025-01-24T00:00:00Z"
}
```

> **Important**: The full API key is only shown once. Store it securely.

#### Using API Keys

Include the API key in the `X-API-Key` header:

```bash
curl https://api.linkrift.io/v1/links \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### API Key Scopes

| Scope | Description |
|-------|-------------|
| `links:read` | Read link information |
| `links:write` | Create, update, delete links |
| `domains:read` | Read domain information |
| `domains:write` | Add, configure, delete domains |
| `analytics:read` | Access analytics data |
| `qrcodes:read` | Read QR code information |
| `qrcodes:write` | Generate and customize QR codes |
| `biopages:read` | Read bio page information |
| `biopages:write` | Create, update, delete bio pages |
| `workspaces:read` | Read workspace information |
| `workspaces:write` | Manage workspace settings and members |
| `webhooks:read` | Read webhook configurations |
| `webhooks:write` | Create, update, delete webhooks |

---

## Request/Response Conventions

### Content Type

All requests with a body must include the `Content-Type: application/json` header.

### Request IDs

Every API response includes an `X-Request-ID` header with a unique identifier for debugging purposes:

```
X-Request-ID: req_abc123def456
```

### Timestamps

All timestamps are returned in ISO 8601 format in UTC:

```json
{
  "created_at": "2025-01-24T12:00:00Z",
  "updated_at": "2025-01-24T14:30:00Z"
}
```

### Null Values

Null values are omitted from responses rather than being explicitly set to `null`.

---

## Pagination

The API uses cursor-based pagination for endpoints that return lists. This provides consistent results even when data is being added or removed.

### Pagination Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Number of items per page (default: 20, max: 100) |
| `cursor` | string | Cursor for the next page of results |

### Paginated Response Format

```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "next_cursor": "cursor_abc123def456",
    "total_count": 150
  }
}
```

### Example

**First page:**

```bash
curl "https://api.linkrift.io/v1/links?limit=20" \
  -H "X-API-Key: {api_key}"
```

**Next page:**

```bash
curl "https://api.linkrift.io/v1/links?limit=20&cursor=cursor_abc123def456" \
  -H "X-API-Key: {api_key}"
```

---

## Rate Limiting

Rate limits are applied per API key or user token. The limits vary by subscription tier.

### Rate Limit Tiers

| Tier | Requests/Minute | Requests/Day | Burst Limit |
|------|-----------------|--------------|-------------|
| Free | 60 | 1,000 | 10 |
| Starter | 300 | 10,000 | 50 |
| Pro | 1,000 | 100,000 | 200 |
| Business | 3,000 | 500,000 | 500 |
| Enterprise | Custom | Custom | Custom |

### Rate Limit Headers

Every response includes headers indicating your current rate limit status:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1706140800
X-RateLimit-Reset-After: 60
```

### Handling Rate Limits

When you exceed the rate limit, the API returns a `429 Too Many Requests` response:

```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded. Please retry after 60 seconds.",
    "retry_after": 60
  }
}
```

Implement exponential backoff when encountering rate limits.

---

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "error": {
    "code": "error_code",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional context about the error"
    },
    "request_id": "req_abc123def456"
  }
}
```

### HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| `200` | Success |
| `201` | Created |
| `204` | No Content (successful deletion) |
| `400` | Bad Request - Invalid parameters |
| `401` | Unauthorized - Missing or invalid authentication |
| `403` | Forbidden - Insufficient permissions |
| `404` | Not Found - Resource does not exist |
| `409` | Conflict - Resource already exists |
| `422` | Unprocessable Entity - Validation error |
| `429` | Too Many Requests - Rate limit exceeded |
| `500` | Internal Server Error |
| `503` | Service Unavailable |

### Error Codes Reference

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `invalid_request` | 400 | The request was malformed |
| `invalid_json` | 400 | Request body is not valid JSON |
| `missing_parameter` | 400 | Required parameter is missing |
| `invalid_parameter` | 400 | Parameter value is invalid |
| `authentication_required` | 401 | No authentication provided |
| `invalid_token` | 401 | Token is invalid or expired |
| `invalid_api_key` | 401 | API key is invalid or revoked |
| `insufficient_permissions` | 403 | Token/key lacks required scope |
| `resource_not_found` | 404 | Requested resource does not exist |
| `link_not_found` | 404 | Link with specified ID not found |
| `domain_not_found` | 404 | Domain with specified ID not found |
| `slug_already_exists` | 409 | Short link slug already in use |
| `domain_already_exists` | 409 | Domain already registered |
| `validation_error` | 422 | Request failed validation |
| `url_invalid` | 422 | Provided URL is not valid |
| `url_blocked` | 422 | URL is on the blocklist |
| `rate_limit_exceeded` | 429 | Too many requests |
| `internal_error` | 500 | Unexpected server error |
| `service_unavailable` | 503 | Service temporarily unavailable |

---

## API Endpoints

### Authentication Endpoints

#### Register User

```http
POST /v1/auth/register
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123!",
  "name": "John Doe"
}
```

**Response:** `201 Created`

```json
{
  "user": {
    "id": "usr_1234567890abcdef",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-24T00:00:00Z"
  },
  "access_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAi...",
  "refresh_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJ0eXBlIjoicmVmcmVzaCIs..."
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123!",
    "name": "John Doe"
  }'
```

#### Login

```http
POST /v1/auth/login
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123!"
}
```

**Response:** `200 OK`

```json
{
  "access_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAi...",
  "refresh_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJ0eXBlIjoicmVmcmVzaCIs...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123!"
  }'
```

#### Refresh Token

```http
POST /v1/auth/refresh
```

**Request Body:**

```json
{
  "refresh_token": "v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1Njc4OTAiLCJ0eXBlIjoicmVmcmVzaCIs..."
}
```

**Response:** `200 OK`

```json
{
  "access_token": "v4.public.eyJuZXdfdG9rZW4iOi...",
  "refresh_token": "v4.public.eyJuZXdfcmVmcmVzaCI6Li4u...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

#### Logout

```http
POST /v1/auth/logout
```

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response:** `204 No Content`

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/auth/logout \
  -H "Authorization: Bearer {access_token}"
```

#### Get Current User

```http
GET /v1/auth/me
```

**Response:** `200 OK`

```json
{
  "id": "usr_1234567890abcdef",
  "email": "user@example.com",
  "name": "John Doe",
  "avatar_url": "https://cdn.linkrift.io/avatars/usr_1234567890abcdef.jpg",
  "plan": "pro",
  "created_at": "2025-01-24T00:00:00Z"
}
```

---

### Links

#### Create Short Link

```http
POST /v1/links
```

**Request Body:**

```json
{
  "url": "https://example.com/very/long/url/that/needs/shortening",
  "slug": "my-custom-slug",
  "domain_id": "dom_1234567890abcdef",
  "title": "Example Page",
  "description": "A description for this link",
  "tags": ["marketing", "campaign-2025"],
  "expires_at": "2025-12-31T23:59:59Z",
  "password": "secretPassword",
  "ios_redirect": "https://apps.apple.com/app/example",
  "android_redirect": "https://play.google.com/store/apps/example",
  "utm_source": "newsletter",
  "utm_medium": "email",
  "utm_campaign": "january-2025"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | The destination URL |
| `slug` | string | No | Custom short code (auto-generated if not provided) |
| `domain_id` | string | No | Custom domain ID (uses default if not provided) |
| `title` | string | No | Link title for organization |
| `description` | string | No | Link description |
| `tags` | array | No | Tags for categorization |
| `expires_at` | string | No | Expiration datetime (ISO 8601) |
| `password` | string | No | Password protection |
| `ios_redirect` | string | No | iOS-specific redirect URL |
| `android_redirect` | string | No | Android-specific redirect URL |
| `utm_source` | string | No | UTM source parameter |
| `utm_medium` | string | No | UTM medium parameter |
| `utm_campaign` | string | No | UTM campaign parameter |

**Response:** `201 Created`

```json
{
  "id": "lnk_1234567890abcdef",
  "short_url": "https://lrift.co/my-custom-slug",
  "original_url": "https://example.com/very/long/url/that/needs/shortening",
  "slug": "my-custom-slug",
  "domain": {
    "id": "dom_1234567890abcdef",
    "hostname": "lrift.co"
  },
  "title": "Example Page",
  "description": "A description for this link",
  "tags": ["marketing", "campaign-2025"],
  "expires_at": "2025-12-31T23:59:59Z",
  "password_protected": true,
  "clicks": 0,
  "created_at": "2025-01-24T12:00:00Z",
  "updated_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/links \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/very/long/url",
    "slug": "my-link",
    "tags": ["marketing"]
  }'
```

#### List Links

```http
GET /v1/links
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Items per page (default: 20, max: 100) |
| `cursor` | string | Pagination cursor |
| `domain_id` | string | Filter by domain |
| `tag` | string | Filter by tag |
| `search` | string | Search in URL, slug, title |
| `created_after` | string | Filter by creation date |
| `created_before` | string | Filter by creation date |

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "lnk_1234567890abcdef",
      "short_url": "https://lrift.co/my-link",
      "original_url": "https://example.com/page",
      "slug": "my-link",
      "domain": {
        "id": "dom_1234567890abcdef",
        "hostname": "lrift.co"
      },
      "clicks": 1234,
      "created_at": "2025-01-24T12:00:00Z"
    }
  ],
  "pagination": {
    "has_more": true,
    "next_cursor": "cursor_abc123",
    "total_count": 150
  }
}
```

**curl Example:**

```bash
curl "https://api.linkrift.io/v1/links?limit=20&tag=marketing" \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Get Link

```http
GET /v1/links/{link_id}
```

**Response:** `200 OK`

```json
{
  "id": "lnk_1234567890abcdef",
  "short_url": "https://lrift.co/my-link",
  "original_url": "https://example.com/page",
  "slug": "my-link",
  "domain": {
    "id": "dom_1234567890abcdef",
    "hostname": "lrift.co"
  },
  "title": "Example Page",
  "description": "A description",
  "tags": ["marketing"],
  "expires_at": null,
  "password_protected": false,
  "ios_redirect": null,
  "android_redirect": null,
  "clicks": 1234,
  "created_at": "2025-01-24T12:00:00Z",
  "updated_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl https://api.linkrift.io/v1/links/lnk_1234567890abcdef \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Update Link

```http
PATCH /v1/links/{link_id}
```

**Request Body:**

```json
{
  "url": "https://example.com/new-destination",
  "title": "Updated Title",
  "tags": ["marketing", "updated"]
}
```

**Response:** `200 OK`

Returns the updated link object.

**curl Example:**

```bash
curl -X PATCH https://api.linkrift.io/v1/links/lnk_1234567890abcdef \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "tags": ["marketing", "updated"]
  }'
```

#### Delete Link

```http
DELETE /v1/links/{link_id}
```

**Response:** `204 No Content`

**curl Example:**

```bash
curl -X DELETE https://api.linkrift.io/v1/links/lnk_1234567890abcdef \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Bulk Create Links

```http
POST /v1/links/bulk
```

**Request Body:**

```json
{
  "links": [
    {
      "url": "https://example.com/page1",
      "tags": ["bulk"]
    },
    {
      "url": "https://example.com/page2",
      "slug": "custom-page2",
      "tags": ["bulk"]
    }
  ]
}
```

**Response:** `201 Created`

```json
{
  "data": [
    {
      "id": "lnk_aaa111",
      "short_url": "https://lrift.co/abc123",
      "original_url": "https://example.com/page1",
      "status": "created"
    },
    {
      "id": "lnk_bbb222",
      "short_url": "https://lrift.co/custom-page2",
      "original_url": "https://example.com/page2",
      "status": "created"
    }
  ],
  "summary": {
    "total": 2,
    "created": 2,
    "failed": 0
  }
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/links/bulk \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "links": [
      {"url": "https://example.com/page1"},
      {"url": "https://example.com/page2"}
    ]
  }'
```

---

### Domains

#### List Domains

```http
GET /v1/domains
```

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "dom_1234567890abcdef",
      "hostname": "links.mycompany.com",
      "verified": true,
      "default": true,
      "ssl_status": "active",
      "created_at": "2025-01-24T00:00:00Z"
    }
  ],
  "pagination": {
    "has_more": false,
    "next_cursor": null,
    "total_count": 1
  }
}
```

**curl Example:**

```bash
curl https://api.linkrift.io/v1/domains \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Add Domain

```http
POST /v1/domains
```

**Request Body:**

```json
{
  "hostname": "links.mycompany.com"
}
```

**Response:** `201 Created`

```json
{
  "id": "dom_1234567890abcdef",
  "hostname": "links.mycompany.com",
  "verified": false,
  "default": false,
  "ssl_status": "pending",
  "verification": {
    "type": "CNAME",
    "name": "_linkrift-verify.links.mycompany.com",
    "value": "verify.linkrift.io"
  },
  "dns_records": [
    {
      "type": "CNAME",
      "name": "links.mycompany.com",
      "value": "redirect.linkrift.io"
    }
  ],
  "created_at": "2025-01-24T00:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/domains \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "links.mycompany.com"
  }'
```

#### Verify Domain

```http
POST /v1/domains/{domain_id}/verify
```

**Response:** `200 OK`

```json
{
  "id": "dom_1234567890abcdef",
  "hostname": "links.mycompany.com",
  "verified": true,
  "ssl_status": "provisioning"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/domains/dom_1234567890abcdef/verify \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Set Default Domain

```http
POST /v1/domains/{domain_id}/default
```

**Response:** `200 OK`

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/domains/dom_1234567890abcdef/default \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Delete Domain

```http
DELETE /v1/domains/{domain_id}
```

**Response:** `204 No Content`

**curl Example:**

```bash
curl -X DELETE https://api.linkrift.io/v1/domains/dom_1234567890abcdef \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

---

### Analytics

#### Get Link Analytics

```http
GET /v1/analytics/links/{link_id}
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (ISO 8601) |
| `end_date` | string | End date (ISO 8601) |
| `interval` | string | Data interval: `hour`, `day`, `week`, `month` |

**Response:** `200 OK`

```json
{
  "link_id": "lnk_1234567890abcdef",
  "period": {
    "start": "2025-01-01T00:00:00Z",
    "end": "2025-01-24T23:59:59Z"
  },
  "summary": {
    "total_clicks": 12345,
    "unique_visitors": 8765,
    "average_daily_clicks": 514
  },
  "timeseries": [
    {
      "timestamp": "2025-01-01T00:00:00Z",
      "clicks": 450,
      "unique_visitors": 380
    },
    {
      "timestamp": "2025-01-02T00:00:00Z",
      "clicks": 520,
      "unique_visitors": 420
    }
  ],
  "top_referrers": [
    {"referrer": "google.com", "clicks": 3456},
    {"referrer": "twitter.com", "clicks": 2345},
    {"referrer": "direct", "clicks": 1890}
  ],
  "top_countries": [
    {"country": "US", "clicks": 5678},
    {"country": "GB", "clicks": 2345},
    {"country": "DE", "clicks": 1234}
  ],
  "devices": {
    "desktop": 6543,
    "mobile": 4567,
    "tablet": 1235
  },
  "browsers": [
    {"browser": "Chrome", "clicks": 7890},
    {"browser": "Safari", "clicks": 2345},
    {"browser": "Firefox", "clicks": 1110}
  ],
  "operating_systems": [
    {"os": "Windows", "clicks": 4567},
    {"os": "macOS", "clicks": 3456},
    {"os": "iOS", "clicks": 2345},
    {"os": "Android", "clicks": 1977}
  ]
}
```

**curl Example:**

```bash
curl "https://api.linkrift.io/v1/analytics/links/lnk_1234567890abcdef?start_date=2025-01-01&end_date=2025-01-24&interval=day" \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Get Account Analytics

```http
GET /v1/analytics/account
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (ISO 8601) |
| `end_date` | string | End date (ISO 8601) |

**Response:** `200 OK`

```json
{
  "period": {
    "start": "2025-01-01T00:00:00Z",
    "end": "2025-01-24T23:59:59Z"
  },
  "summary": {
    "total_links": 567,
    "total_clicks": 123456,
    "unique_visitors": 89012
  },
  "top_links": [
    {
      "id": "lnk_aaa111",
      "short_url": "https://lrift.co/top-link",
      "clicks": 12345
    }
  ]
}
```

**curl Example:**

```bash
curl "https://api.linkrift.io/v1/analytics/account?start_date=2025-01-01&end_date=2025-01-24" \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz"
```

#### Get Real-time Analytics

```http
GET /v1/analytics/realtime
```

**Response:** `200 OK`

```json
{
  "active_visitors": 42,
  "clicks_last_hour": 156,
  "clicks_last_24h": 3456,
  "recent_clicks": [
    {
      "link_id": "lnk_1234567890abcdef",
      "short_url": "https://lrift.co/my-link",
      "country": "US",
      "city": "San Francisco",
      "device": "mobile",
      "timestamp": "2025-01-24T14:30:00Z"
    }
  ]
}
```

---

### QR Codes

#### Generate QR Code

```http
POST /v1/qrcodes
```

**Request Body:**

```json
{
  "link_id": "lnk_1234567890abcdef",
  "format": "png",
  "size": 512,
  "foreground_color": "#000000",
  "background_color": "#FFFFFF",
  "logo_url": "https://example.com/logo.png",
  "error_correction": "M"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `link_id` | string | Link to generate QR code for |
| `format` | string | Output format: `png`, `svg`, `pdf` |
| `size` | integer | Size in pixels (default: 256, max: 2048) |
| `foreground_color` | string | Hex color for QR modules |
| `background_color` | string | Hex color for background |
| `logo_url` | string | URL of logo to embed in center |
| `error_correction` | string | Error correction level: `L`, `M`, `Q`, `H` |

**Response:** `201 Created`

```json
{
  "id": "qr_1234567890abcdef",
  "link_id": "lnk_1234567890abcdef",
  "url": "https://cdn.linkrift.io/qrcodes/qr_1234567890abcdef.png",
  "format": "png",
  "size": 512,
  "created_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/qrcodes \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "link_id": "lnk_1234567890abcdef",
    "format": "png",
    "size": 512,
    "foreground_color": "#1a1a1a"
  }'
```

#### Get QR Code

```http
GET /v1/qrcodes/{qrcode_id}
```

**Response:** `200 OK`

Returns the QR code metadata. Use the `url` field to download the image.

#### List QR Codes

```http
GET /v1/qrcodes
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `link_id` | string | Filter by link |
| `limit` | integer | Items per page |
| `cursor` | string | Pagination cursor |

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "qr_1234567890abcdef",
      "link_id": "lnk_1234567890abcdef",
      "url": "https://cdn.linkrift.io/qrcodes/qr_1234567890abcdef.png",
      "format": "png",
      "created_at": "2025-01-24T12:00:00Z"
    }
  ],
  "pagination": {
    "has_more": false,
    "next_cursor": null,
    "total_count": 1
  }
}
```

#### Download QR Code

```http
GET /v1/qrcodes/{qrcode_id}/download
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `format` | string | Override format: `png`, `svg`, `pdf` |
| `size` | integer | Override size |

**Response:** Binary image data with appropriate `Content-Type` header.

**curl Example:**

```bash
curl "https://api.linkrift.io/v1/qrcodes/qr_1234567890abcdef/download?format=svg" \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -o qrcode.svg
```

---

### Bio Pages

#### Create Bio Page

```http
POST /v1/biopages
```

**Request Body:**

```json
{
  "username": "johndoe",
  "title": "John Doe",
  "bio": "Developer & Content Creator",
  "avatar_url": "https://example.com/avatar.jpg",
  "theme": "dark",
  "custom_css": "",
  "links": [
    {
      "title": "My Website",
      "url": "https://johndoe.com",
      "icon": "globe"
    },
    {
      "title": "Twitter",
      "url": "https://twitter.com/johndoe",
      "icon": "twitter"
    }
  ],
  "social_links": {
    "twitter": "johndoe",
    "github": "johndoe",
    "linkedin": "johndoe"
  }
}
```

**Response:** `201 Created`

```json
{
  "id": "bio_1234567890abcdef",
  "username": "johndoe",
  "url": "https://lrift.co/bio/johndoe",
  "title": "John Doe",
  "bio": "Developer & Content Creator",
  "avatar_url": "https://example.com/avatar.jpg",
  "theme": "dark",
  "links": [
    {
      "id": "bl_aaa111",
      "title": "My Website",
      "url": "https://johndoe.com",
      "icon": "globe",
      "clicks": 0
    }
  ],
  "created_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/biopages \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "title": "John Doe",
    "bio": "Developer",
    "links": [
      {"title": "Website", "url": "https://johndoe.com"}
    ]
  }'
```

#### Get Bio Page

```http
GET /v1/biopages/{biopage_id}
```

**Response:** `200 OK`

Returns the full bio page object.

#### Update Bio Page

```http
PATCH /v1/biopages/{biopage_id}
```

**Request Body:**

```json
{
  "title": "John Doe - Updated",
  "bio": "Updated bio text"
}
```

**Response:** `200 OK`

#### Delete Bio Page

```http
DELETE /v1/biopages/{biopage_id}
```

**Response:** `204 No Content`

#### Add Link to Bio Page

```http
POST /v1/biopages/{biopage_id}/links
```

**Request Body:**

```json
{
  "title": "New Link",
  "url": "https://example.com/new",
  "icon": "link",
  "position": 0
}
```

**Response:** `201 Created`

#### Reorder Bio Page Links

```http
PUT /v1/biopages/{biopage_id}/links/reorder
```

**Request Body:**

```json
{
  "link_ids": ["bl_ccc333", "bl_aaa111", "bl_bbb222"]
}
```

**Response:** `200 OK`

---

### Workspaces

#### Create Workspace

```http
POST /v1/workspaces
```

**Request Body:**

```json
{
  "name": "Marketing Team",
  "slug": "marketing"
}
```

**Response:** `201 Created`

```json
{
  "id": "ws_1234567890abcdef",
  "name": "Marketing Team",
  "slug": "marketing",
  "owner_id": "usr_1234567890abcdef",
  "member_count": 1,
  "created_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/workspaces \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Marketing Team",
    "slug": "marketing"
  }'
```

#### List Workspaces

```http
GET /v1/workspaces
```

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "ws_1234567890abcdef",
      "name": "Marketing Team",
      "slug": "marketing",
      "role": "owner",
      "member_count": 5,
      "created_at": "2025-01-24T12:00:00Z"
    }
  ],
  "pagination": {
    "has_more": false,
    "next_cursor": null,
    "total_count": 1
  }
}
```

#### Get Workspace

```http
GET /v1/workspaces/{workspace_id}
```

**Response:** `200 OK`

```json
{
  "id": "ws_1234567890abcdef",
  "name": "Marketing Team",
  "slug": "marketing",
  "owner": {
    "id": "usr_1234567890abcdef",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "settings": {
    "default_domain_id": "dom_1234567890abcdef",
    "allow_member_domains": false
  },
  "member_count": 5,
  "link_count": 234,
  "created_at": "2025-01-24T12:00:00Z"
}
```

#### Invite Member

```http
POST /v1/workspaces/{workspace_id}/members
```

**Request Body:**

```json
{
  "email": "teammate@example.com",
  "role": "member"
}
```

| Role | Permissions |
|------|-------------|
| `owner` | Full access, can delete workspace |
| `admin` | Manage members, settings, full link access |
| `member` | Create, edit, delete own links |
| `viewer` | Read-only access |

**Response:** `201 Created`

```json
{
  "id": "inv_1234567890abcdef",
  "email": "teammate@example.com",
  "role": "member",
  "status": "pending",
  "expires_at": "2025-01-31T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/workspaces/ws_1234567890abcdef/members \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "teammate@example.com",
    "role": "member"
  }'
```

#### List Workspace Members

```http
GET /v1/workspaces/{workspace_id}/members
```

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "usr_1234567890abcdef",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "owner",
      "joined_at": "2025-01-24T12:00:00Z"
    },
    {
      "id": "usr_abcdef1234567890",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "role": "member",
      "joined_at": "2025-01-24T14:00:00Z"
    }
  ]
}
```

#### Update Member Role

```http
PATCH /v1/workspaces/{workspace_id}/members/{user_id}
```

**Request Body:**

```json
{
  "role": "admin"
}
```

**Response:** `200 OK`

#### Remove Member

```http
DELETE /v1/workspaces/{workspace_id}/members/{user_id}
```

**Response:** `204 No Content`

---

### Webhooks

#### Create Webhook

```http
POST /v1/webhooks
```

**Request Body:**

```json
{
  "url": "https://example.com/webhooks/linkrift",
  "events": ["link.clicked", "link.created", "link.updated"],
  "secret": "whsec_your_webhook_secret"
}
```

| Event | Description |
|-------|-------------|
| `link.created` | A new link was created |
| `link.updated` | A link was updated |
| `link.deleted` | A link was deleted |
| `link.clicked` | A link was clicked |
| `link.expired` | A link expired |
| `domain.verified` | A domain was verified |
| `domain.ssl_provisioned` | SSL certificate was provisioned |
| `biopage.created` | A bio page was created |
| `biopage.updated` | A bio page was updated |

**Response:** `201 Created`

```json
{
  "id": "wh_1234567890abcdef",
  "url": "https://example.com/webhooks/linkrift",
  "events": ["link.clicked", "link.created", "link.updated"],
  "secret": "whsec_your_webhook_secret",
  "active": true,
  "created_at": "2025-01-24T12:00:00Z"
}
```

**curl Example:**

```bash
curl -X POST https://api.linkrift.io/v1/webhooks \
  -H "X-API-Key: lr_live_sk_1234567890abcdefghijklmnopqrstuvwxyz" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/webhooks/linkrift",
    "events": ["link.clicked", "link.created"]
  }'
```

#### List Webhooks

```http
GET /v1/webhooks
```

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "wh_1234567890abcdef",
      "url": "https://example.com/webhooks/linkrift",
      "events": ["link.clicked", "link.created", "link.updated"],
      "active": true,
      "created_at": "2025-01-24T12:00:00Z"
    }
  ]
}
```

#### Get Webhook

```http
GET /v1/webhooks/{webhook_id}
```

**Response:** `200 OK`

#### Update Webhook

```http
PATCH /v1/webhooks/{webhook_id}
```

**Request Body:**

```json
{
  "events": ["link.clicked"],
  "active": true
}
```

**Response:** `200 OK`

#### Delete Webhook

```http
DELETE /v1/webhooks/{webhook_id}
```

**Response:** `204 No Content`

#### Webhook Payload Format

```json
{
  "id": "evt_1234567890abcdef",
  "type": "link.clicked",
  "created_at": "2025-01-24T14:30:00Z",
  "data": {
    "link_id": "lnk_1234567890abcdef",
    "short_url": "https://lrift.co/my-link",
    "click": {
      "country": "US",
      "city": "San Francisco",
      "device": "mobile",
      "browser": "Chrome",
      "os": "iOS",
      "referrer": "twitter.com",
      "timestamp": "2025-01-24T14:30:00Z"
    }
  }
}
```

#### Webhook Signature Verification

All webhook requests include a signature in the `X-Linkrift-Signature` header. Verify this signature to ensure the webhook is from Linkrift.

**Signature Format:**

```
X-Linkrift-Signature: t=1706104200,v1=5257a869e7ecebeda32affa62cdca3fa51cad7e77a0e56ff536d0ce8e108d8bd
```

**Verification Process:**

1. Extract the timestamp (`t`) and signature (`v1`) from the header
2. Construct the signed payload: `{timestamp}.{request_body}`
3. Compute HMAC-SHA256 using your webhook secret
4. Compare signatures using constant-time comparison

**Example (Go):**

```go
func VerifyWebhookSignature(payload []byte, header string, secret string) bool {
    parts := strings.Split(header, ",")
    var timestamp, signature string

    for _, part := range parts {
        kv := strings.SplitN(part, "=", 2)
        if len(kv) == 2 {
            switch kv[0] {
            case "t":
                timestamp = kv[1]
            case "v1":
                signature = kv[1]
            }
        }
    }

    signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(signedPayload))
    expectedSignature := hex.EncodeToString(mac.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## Additional Resources

- **API Status**: [status.linkrift.io](https://status.linkrift.io)
- **Changelog**: [docs.linkrift.io/changelog](https://docs.linkrift.io/changelog)
- **Support**: [support@linkrift.io](mailto:support@linkrift.io)
- **Community**: [community.linkrift.io](https://community.linkrift.io)

---

*This documentation is generated from the Linkrift API specification. For the most up-to-date information, visit [docs.linkrift.io/api](https://docs.linkrift.io/api).*
