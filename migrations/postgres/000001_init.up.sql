BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. users
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
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

CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);

-- ============================================================================
-- 2. workspaces
-- ============================================================================
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_workspaces_slug ON workspaces(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);

-- ============================================================================
-- 3. workspace_members
-- ============================================================================
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

-- ============================================================================
-- 4. domains
-- ============================================================================
CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ssl_expires_at TIMESTAMPTZ,
    dns_records JSONB NOT NULL DEFAULT '[]',
    last_dns_check_at TIMESTAMPTZ,
    default_redirect_url TEXT,
    custom_404_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_domains_domain ON domains(domain) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_workspace ON domains(workspace_id);

-- ============================================================================
-- 5. links
-- ============================================================================
CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    domain_id UUID REFERENCES domains(id) ON DELETE SET NULL,
    url TEXT NOT NULL,
    short_code VARCHAR(50) NOT NULL,
    title VARCHAR(500),
    description TEXT,
    favicon_url VARCHAR(500),
    og_image_url VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    password_hash VARCHAR(255),
    expires_at TIMESTAMPTZ,
    max_clicks INTEGER,
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    utm_term VARCHAR(255),
    utm_content VARCHAR(255),
    total_clicks BIGINT NOT NULL DEFAULT 0,
    unique_clicks BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_links_short_code ON links(short_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_user ON links(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_workspace ON links(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_domain ON links(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_links_created_at ON links(created_at DESC);
CREATE INDEX idx_links_search ON links USING GIN (to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(description, '')));

-- ============================================================================
-- 6. link_rules
-- ============================================================================
CREATE TABLE link_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    rule_type VARCHAR(50) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    conditions JSONB NOT NULL DEFAULT '{}',
    destination_url TEXT NOT NULL,
    weight INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_link_rules_link ON link_rules(link_id);
CREATE INDEX idx_link_rules_active ON link_rules(link_id, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_link_rules_conditions ON link_rules USING GIN (conditions);

-- ============================================================================
-- 7. tags
-- ============================================================================
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, name)
);

CREATE INDEX idx_tags_workspace ON tags(workspace_id);

-- ============================================================================
-- 8. link_tags
-- ============================================================================
CREATE TABLE link_tags (
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (link_id, tag_id)
);

CREATE INDEX idx_link_tags_tag ON link_tags(tag_id);

-- ============================================================================
-- 9. clicks (partitioned by month)
-- ============================================================================
CREATE TABLE clicks (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    visitor_id VARCHAR(64),
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country_code CHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),
    device_type VARCHAR(20),
    browser VARCHAR(50),
    browser_version VARCHAR(20),
    os VARCHAR(50),
    os_version VARCHAR(20),
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    PRIMARY KEY (id, clicked_at)
) PARTITION BY RANGE (clicked_at);

CREATE TABLE clicks_2025_01 PARTITION OF clicks FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE clicks_2025_02 PARTITION OF clicks FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
CREATE TABLE clicks_2025_03 PARTITION OF clicks FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');
CREATE TABLE clicks_2025_04 PARTITION OF clicks FOR VALUES FROM ('2025-04-01') TO ('2025-05-01');
CREATE TABLE clicks_2025_05 PARTITION OF clicks FOR VALUES FROM ('2025-05-01') TO ('2025-06-01');
CREATE TABLE clicks_2025_06 PARTITION OF clicks FOR VALUES FROM ('2025-06-01') TO ('2025-07-01');
CREATE TABLE clicks_2025_07 PARTITION OF clicks FOR VALUES FROM ('2025-07-01') TO ('2025-08-01');
CREATE TABLE clicks_2025_08 PARTITION OF clicks FOR VALUES FROM ('2025-08-01') TO ('2025-09-01');
CREATE TABLE clicks_2025_09 PARTITION OF clicks FOR VALUES FROM ('2025-09-01') TO ('2025-10-01');
CREATE TABLE clicks_2025_10 PARTITION OF clicks FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
CREATE TABLE clicks_2025_11 PARTITION OF clicks FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE clicks_2025_12 PARTITION OF clicks FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE clicks_2026_01 PARTITION OF clicks FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE clicks_2026_02 PARTITION OF clicks FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
CREATE TABLE clicks_2026_03 PARTITION OF clicks FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE clicks_2026_04 PARTITION OF clicks FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE clicks_2026_05 PARTITION OF clicks FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE clicks_2026_06 PARTITION OF clicks FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE INDEX idx_clicks_link_id ON clicks(link_id, clicked_at DESC);
CREATE INDEX idx_clicks_visitor ON clicks(visitor_id, clicked_at DESC);

-- ============================================================================
-- 10. qr_codes
-- ============================================================================
CREATE TABLE qr_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    qr_type VARCHAR(20) NOT NULL DEFAULT 'dynamic',
    error_correction VARCHAR(1) NOT NULL DEFAULT 'M',
    foreground_color VARCHAR(7) NOT NULL DEFAULT '#000000',
    background_color VARCHAR(7) NOT NULL DEFAULT '#FFFFFF',
    logo_url VARCHAR(500),
    png_url VARCHAR(500),
    svg_url VARCHAR(500),
    scan_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_qr_codes_link ON qr_codes(link_id);

-- ============================================================================
-- 11. bio_pages
-- ============================================================================
CREATE TABLE bio_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url VARCHAR(500),
    theme_id UUID,
    custom_css TEXT,
    meta_title VARCHAR(100),
    meta_description VARCHAR(300),
    og_image_url VARCHAR(500),
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_bio_pages_slug ON bio_pages(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_bio_pages_workspace ON bio_pages(workspace_id);

-- ============================================================================
-- 12. bio_page_links
-- ============================================================================
CREATE TABLE bio_page_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bio_page_id UUID NOT NULL REFERENCES bio_pages(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    icon VARCHAR(50),
    position INTEGER NOT NULL DEFAULT 0,
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,
    visible_from TIMESTAMPTZ,
    visible_until TIMESTAMPTZ,
    click_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bio_page_links_page ON bio_page_links(bio_page_id, position);

-- ============================================================================
-- 13. api_keys
-- ============================================================================
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    last_used_at TIMESTAMPTZ,
    request_count BIGINT NOT NULL DEFAULT 0,
    rate_limit INTEGER,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_workspace ON api_keys(workspace_id);

-- ============================================================================
-- 14. webhooks
-- ============================================================================
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    failure_count INTEGER NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_workspace ON webhooks(workspace_id);

-- ============================================================================
-- 15. webhook_deliveries
-- ============================================================================
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    response_status INTEGER,
    response_body TEXT,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    last_attempt_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC);

-- ============================================================================
-- 16. sessions
-- ============================================================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    device_name VARCHAR(255),
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- ============================================================================
-- 17. password_resets
-- ============================================================================
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

-- ============================================================================
-- 18. audit_logs
-- ============================================================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id, created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- ============================================================================
-- 19. subscriptions
-- ============================================================================
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    usage_link_count INTEGER NOT NULL DEFAULT 0,
    usage_click_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_workspace ON subscriptions(workspace_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);

COMMIT;
