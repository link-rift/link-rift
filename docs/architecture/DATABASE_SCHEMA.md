# Database Schema

> Last Updated: 2025-01-24

Complete database schema documentation for PostgreSQL, ClickHouse, and Redis.

## Table of Contents

- [PostgreSQL Schema](#postgresql-schema)
- [ClickHouse Schema](#clickhouse-schema)
- [Redis Key Patterns](#redis-key-patterns)
- [Migrations](#migrations)
- [Indexes](#indexes)
- [Query Examples](#query-examples)

---

## PostgreSQL Schema

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ENTITY RELATIONSHIPS                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────┐     ┌──────────────────┐     ┌──────────────┐                │
│  │  users   │────<│workspace_members │>────│  workspaces  │                │
│  └────┬─────┘     └──────────────────┘     └──────┬───────┘                │
│       │                                          │                         │
│       │           ┌──────────────────┐           │                         │
│       └──────────>│      links       │<──────────┘                         │
│                   └────────┬─────────┘                                     │
│                            │                                               │
│           ┌────────────────┼────────────────┐                              │
│           │                │                │                              │
│           ▼                ▼                ▼                              │
│    ┌──────────┐     ┌──────────┐     ┌──────────┐                         │
│    │link_rules│     │link_tags │     │ clicks   │                         │
│    └──────────┘     └──────────┘     └──────────┘                         │
│                                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌──────────────┐                   │
│  │ domains  │     │   qr_codes   │     │  bio_pages   │                   │
│  └──────────┘     └──────────────┘     └──────────────┘                   │
│                                                                             │
│  ┌──────────┐     ┌──────────────┐     ┌──────────────┐                   │
│  │ api_keys │     │   webhooks   │     │subscriptions │                   │
│  └──────────┘     └──────────────┘     └──────────────┘                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Core Tables

#### users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    email_verified_at TIMESTAMPTZ,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### workspaces

```sql
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_slug ON workspaces(slug) WHERE deleted_at IS NULL;
```

#### workspace_members

```sql
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
```

#### links

```sql
CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    domain_id UUID REFERENCES domains(id) ON DELETE SET NULL,

    -- URL data
    url TEXT NOT NULL,
    short_code VARCHAR(50) NOT NULL,
    title VARCHAR(500),
    description TEXT,
    favicon_url VARCHAR(500),
    og_image_url VARCHAR(500),

    -- Settings
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    password_hash VARCHAR(255),
    expires_at TIMESTAMPTZ,
    max_clicks INTEGER,

    -- Metadata
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    utm_term VARCHAR(255),
    utm_content VARCHAR(255),

    -- Counters (denormalized for performance)
    total_clicks BIGINT NOT NULL DEFAULT 0,
    unique_clicks BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(short_code) WHERE deleted_at IS NULL
);

CREATE INDEX idx_links_user ON links(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_workspace ON links(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_short_code ON links(short_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_domain ON links(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_created_at ON links(created_at DESC);
```

#### link_rules

```sql
CREATE TABLE link_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,

    -- Rule configuration
    rule_type VARCHAR(50) NOT NULL,  -- device, geo, time, ab_test, referrer
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Conditions (JSONB for flexibility)
    conditions JSONB NOT NULL DEFAULT '{}',

    -- Destination
    destination_url TEXT NOT NULL,

    -- A/B testing
    weight INTEGER,  -- For A/B test distribution

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_link_rules_link ON link_rules(link_id);
CREATE INDEX idx_link_rules_active ON link_rules(link_id, is_active) WHERE is_active = TRUE;
```

#### domains

```sql
CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    domain VARCHAR(255) NOT NULL UNIQUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,

    -- SSL
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ssl_expires_at TIMESTAMPTZ,

    -- DNS
    dns_records JSONB NOT NULL DEFAULT '[]',
    last_dns_check_at TIMESTAMPTZ,

    -- Settings
    default_redirect_url TEXT,
    custom_404_url TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_domains_workspace ON domains(workspace_id);
CREATE INDEX idx_domains_domain ON domains(domain) WHERE deleted_at IS NULL;
```

#### clicks (Partitioned)

```sql
-- Partitioned by month for efficient querying and data retention
CREATE TABLE clicks (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL,

    -- Timestamp (partition key)
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Visitor info
    visitor_id VARCHAR(64),  -- Fingerprint hash
    ip_address INET,
    user_agent TEXT,
    referer TEXT,

    -- Geo data
    country_code CHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),

    -- Device data
    device_type VARCHAR(20),  -- desktop, mobile, tablet
    browser VARCHAR(50),
    browser_version VARCHAR(20),
    os VARCHAR(50),
    os_version VARCHAR(20),

    -- Bot detection
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,

    -- UTM captured
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),

    PRIMARY KEY (id, clicked_at)
) PARTITION BY RANGE (clicked_at);

-- Create monthly partitions
CREATE TABLE clicks_2025_01 PARTITION OF clicks
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE clicks_2025_02 PARTITION OF clicks
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Indexes on partitioned table
CREATE INDEX idx_clicks_link_id ON clicks(link_id, clicked_at DESC);
CREATE INDEX idx_clicks_visitor ON clicks(visitor_id, clicked_at DESC);
```

### Supporting Tables

#### qr_codes

```sql
CREATE TABLE qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,

    -- QR Configuration
    qr_type VARCHAR(20) NOT NULL DEFAULT 'dynamic',  -- dynamic, static
    error_correction VARCHAR(1) NOT NULL DEFAULT 'M',  -- L, M, Q, H

    -- Styling
    foreground_color VARCHAR(7) NOT NULL DEFAULT '#000000',
    background_color VARCHAR(7) NOT NULL DEFAULT '#FFFFFF',
    logo_url VARCHAR(500),

    -- Generated files
    png_url VARCHAR(500),
    svg_url VARCHAR(500),

    -- Counters
    scan_count BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_qr_codes_link ON qr_codes(link_id);
```

#### bio_pages

```sql
CREATE TABLE bio_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Page info
    slug VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url VARCHAR(500),

    -- Styling
    theme_id UUID REFERENCES bio_page_themes(id),
    custom_css TEXT,

    -- SEO
    meta_title VARCHAR(100),
    meta_description VARCHAR(300),
    og_image_url VARCHAR(500),

    -- Settings
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_bio_pages_workspace ON bio_pages(workspace_id);
CREATE INDEX idx_bio_pages_slug ON bio_pages(slug) WHERE deleted_at IS NULL;
```

#### bio_page_links

```sql
CREATE TABLE bio_page_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bio_page_id UUID NOT NULL REFERENCES bio_pages(id) ON DELETE CASCADE,

    -- Content
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    icon VARCHAR(50),

    -- Display
    position INTEGER NOT NULL DEFAULT 0,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,

    -- Scheduling
    visible_from TIMESTAMPTZ,
    visible_until TIMESTAMPTZ,

    -- Analytics
    click_count BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bio_page_links_page ON bio_page_links(bio_page_id, position);
```

#### api_keys

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,  -- Only store hash
    key_prefix VARCHAR(12) NOT NULL,  -- For display: "lr_abc123..."

    -- Permissions
    scopes TEXT[] NOT NULL DEFAULT '{}',

    -- Usage
    last_used_at TIMESTAMPTZ,
    request_count BIGINT NOT NULL DEFAULT 0,

    -- Limits
    rate_limit INTEGER,  -- Override default
    expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
```

#### webhooks

```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,  -- For HMAC signing

    -- Events
    events TEXT[] NOT NULL DEFAULT '{}',  -- link.created, link.clicked, etc.

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    failure_count INTEGER NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_workspace ON webhooks(workspace_id);
```

#### subscriptions

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Stripe
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,

    -- Plan
    plan VARCHAR(50) NOT NULL,  -- pro, business, enterprise
    status VARCHAR(50) NOT NULL,  -- active, past_due, canceled

    -- Billing
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,

    -- Usage (for metered billing)
    usage_link_count INTEGER NOT NULL DEFAULT 0,
    usage_click_count BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_workspace ON subscriptions(workspace_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
```

#### audit_logs

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Action
    action VARCHAR(100) NOT NULL,  -- link.created, user.invited, etc.
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,

    -- Details
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,

    -- Context
    ip_address INET,
    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id, created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
```

### Authentication Tables

#### sessions

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,

    -- Device info
    ip_address INET,
    user_agent TEXT,
    device_name VARCHAR(255),

    -- Status
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

#### password_resets

```sql
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token_hash VARCHAR(255) NOT NULL UNIQUE,

    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_password_resets_token ON password_resets(token_hash);
CREATE INDEX idx_password_resets_user ON password_resets(user_id);
```

---

## ClickHouse Schema

### clicks_analytics

```sql
CREATE TABLE clicks_analytics (
    -- Identifiers
    link_id UUID,
    workspace_id UUID,

    -- Time
    clicked_at DateTime,
    date Date MATERIALIZED toDate(clicked_at),
    hour UInt8 MATERIALIZED toHour(clicked_at),

    -- Visitor
    visitor_id String,
    is_unique UInt8,

    -- Geo
    country_code LowCardinality(String),
    region String,
    city String,

    -- Device
    device_type LowCardinality(String),
    browser LowCardinality(String),
    os LowCardinality(String),

    -- Traffic source
    referer_domain String,
    utm_source String,
    utm_medium String,
    utm_campaign String,

    -- Bot
    is_bot UInt8
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (workspace_id, link_id, clicked_at)
TTL clicked_at + INTERVAL 2 YEAR;
```

### click_aggregates_daily

```sql
CREATE TABLE click_aggregates_daily (
    date Date,
    link_id UUID,
    workspace_id UUID,

    -- Counts
    total_clicks UInt64,
    unique_clicks UInt64,

    -- By dimension
    clicks_by_country Map(String, UInt64),
    clicks_by_device Map(String, UInt64),
    clicks_by_browser Map(String, UInt64),
    clicks_by_os Map(String, UInt64),
    clicks_by_referer Map(String, UInt64),

    -- Hourly distribution
    clicks_by_hour Array(UInt64)
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (workspace_id, link_id, date);
```

### Materialized Views

```sql
-- Aggregate clicks to daily table
CREATE MATERIALIZED VIEW clicks_to_daily_mv
TO click_aggregates_daily
AS SELECT
    toDate(clicked_at) AS date,
    link_id,
    workspace_id,
    count() AS total_clicks,
    uniqExact(visitor_id) AS unique_clicks,
    sumMap(map(country_code, 1)) AS clicks_by_country,
    sumMap(map(device_type, 1)) AS clicks_by_device,
    sumMap(map(browser, 1)) AS clicks_by_browser,
    sumMap(map(os, 1)) AS clicks_by_os,
    sumMap(map(referer_domain, 1)) AS clicks_by_referer,
    groupArray(toHour(clicked_at)) AS clicks_by_hour
FROM clicks_analytics
WHERE is_bot = 0
GROUP BY date, link_id, workspace_id;
```

---

## Redis Key Patterns

### Link Cache

```
link:{short_code}
├── url: "https://example.com"
├── rules: [{...}]
├── is_active: true
└── TTL: 300s

link:rules:{short_code}
├── rules: [{type: "device", ...}]
└── TTL: 300s
```

### Sessions

```
session:{refresh_token}
├── user_id: "uuid"
├── expires_at: timestamp
└── TTL: 7 days

user:sessions:{user_id}
├── [session_ids]
└── TTL: none
```

### Rate Limiting

```
rate:{ip}:{endpoint}
├── count: 42
├── window_start: timestamp
└── TTL: 60s

rate:api:{api_key_prefix}
├── count: 150
└── TTL: 60s
```

### Real-time Counters

```
clicks:realtime:{link_id}
├── count: 1523
└── TTL: 86400s

clicks:realtime:total:{workspace_id}
├── count: 50234
└── TTL: 86400s
```

### Job Queue (Asynq)

```
asynq:{queue}:pending    # List of pending jobs
asynq:{queue}:active     # Currently processing
asynq:{queue}:scheduled  # Future scheduled jobs
asynq:{queue}:retry      # Jobs to retry
asynq:{queue}:dead       # Failed jobs
```

---

## Migrations

### Creating Migrations

```bash
# Create new migration
make migrate-create name=add_tags_table

# This creates:
# migrations/postgres/000005_add_tags_table.up.sql
# migrations/postgres/000005_add_tags_table.down.sql
```

### Migration Best Practices

```sql
-- Up migration
-- migrations/postgres/000005_add_tags_table.up.sql

BEGIN;

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, name)
);

CREATE TABLE link_tags (
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (link_id, tag_id)
);

CREATE INDEX idx_tags_workspace ON tags(workspace_id);
CREATE INDEX idx_link_tags_tag ON link_tags(tag_id);

COMMIT;
```

```sql
-- Down migration
-- migrations/postgres/000005_add_tags_table.down.sql

BEGIN;

DROP TABLE IF EXISTS link_tags;
DROP TABLE IF EXISTS tags;

COMMIT;
```

### Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check status
make migrate-status

# Force version (use carefully)
migrate -path migrations/postgres -database "$DATABASE_URL" force 5
```

---

## Indexes

### Index Strategy

| Type | Use Case | Example |
|------|----------|---------|
| B-tree | Equality, range queries | `idx_links_created_at` |
| Hash | Equality only | `idx_api_keys_prefix` |
| GIN | JSONB, arrays | `idx_link_rules_conditions` |
| Partial | Filtered queries | `WHERE deleted_at IS NULL` |

### Key Indexes

```sql
-- Frequently queried
CREATE INDEX idx_links_short_code ON links(short_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_workspace_created ON links(workspace_id, created_at DESC) WHERE deleted_at IS NULL;

-- Foreign keys (for JOIN performance)
CREATE INDEX idx_links_user ON links(user_id);
CREATE INDEX idx_links_domain ON links(domain_id);

-- JSONB queries
CREATE INDEX idx_link_rules_conditions ON link_rules USING GIN (conditions);

-- Full-text search
CREATE INDEX idx_links_search ON links USING GIN (to_tsvector('english', title || ' ' || COALESCE(description, '')));
```

---

## Query Examples

### Get Link with Rules

```sql
-- name: GetLinkWithRules :one
SELECT
    l.*,
    COALESCE(
        json_agg(r.*) FILTER (WHERE r.id IS NOT NULL),
        '[]'
    ) AS rules
FROM links l
LEFT JOIN link_rules r ON r.link_id = l.id AND r.is_active = TRUE
WHERE l.short_code = $1 AND l.deleted_at IS NULL
GROUP BY l.id;
```

### List Links with Pagination

```sql
-- name: ListLinksForWorkspace :many
SELECT
    l.*,
    COUNT(*) OVER() AS total_count
FROM links l
WHERE l.workspace_id = $1
    AND l.deleted_at IS NULL
    AND ($2::text IS NULL OR l.title ILIKE '%' || $2 || '%')
ORDER BY l.created_at DESC
LIMIT $3 OFFSET $4;
```

### Analytics Summary

```sql
-- name: GetLinkAnalyticsSummary :one
SELECT
    COUNT(*) AS total_clicks,
    COUNT(DISTINCT visitor_id) AS unique_clicks,
    COUNT(*) FILTER (WHERE device_type = 'mobile') AS mobile_clicks,
    COUNT(*) FILTER (WHERE device_type = 'desktop') AS desktop_clicks,
    jsonb_object_agg(country_code, cnt) AS clicks_by_country
FROM (
    SELECT
        visitor_id,
        device_type,
        country_code,
        COUNT(*) AS cnt
    FROM clicks
    WHERE link_id = $1
        AND clicked_at >= $2
        AND clicked_at <= $3
    GROUP BY visitor_id, device_type, country_code
) sub;
```

---

## Related Documentation

- [Architecture](ARCHITECTURE.md) — System design
- [Go Patterns](GO_PATTERNS.md) — Repository patterns
- [Analytics Pipeline](../features/ANALYTICS_PIPELINE.md) — Click processing
- [Scaling Guide](../deployment/SCALING_GUIDE.md) — Database scaling
