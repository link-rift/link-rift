# Linkrift Production Deployment Guide

**Last Updated: 2025-01-24**

This guide covers the complete production deployment process for Linkrift, a high-performance URL shortener designed for sub-millisecond redirect latency.

---

## Table of Contents

1. [Infrastructure Requirements](#infrastructure-requirements)
2. [Server Provisioning](#server-provisioning)
3. [Secrets Management](#secrets-management)
4. [SSL/TLS Setup with Cloudflare](#ssltls-setup-with-cloudflare)
5. [Database Migration in Production](#database-migration-in-production)
6. [Go Binary Build Process](#go-binary-build-process)
7. [Vite Production Build](#vite-production-build)
8. [Zero-Downtime Deployment](#zero-downtime-deployment)
9. [Rollback Procedures](#rollback-procedures)

---

## Infrastructure Requirements

### Minimum Production Setup

| Component | Specification | Purpose |
|-----------|--------------|---------|
| **API Server** | 2 vCPU, 4GB RAM | REST API, link management |
| **Redirect Server** | 2 vCPU, 2GB RAM | High-performance redirects |
| **Worker Server** | 2 vCPU, 4GB RAM | Background job processing |
| **PostgreSQL** | 4 vCPU, 8GB RAM, 100GB SSD | Primary data store |
| **Redis** | 2 vCPU, 4GB RAM | Caching, rate limiting |
| **ClickHouse** | 4 vCPU, 16GB RAM, 500GB SSD | Analytics storage |

### Recommended Production Setup (High Availability)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cloudflare CDN                           │
│                    (SSL Termination, DDoS)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Load Balancer (HAProxy)                    │
│                     (Health checks, routing)                    │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│  API Server   │       │  API Server   │       │  API Server   │
│   (Primary)   │       │  (Replica 1)  │       │  (Replica 2)  │
└───────────────┘       └───────────────┘       └───────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Redis Cluster (3 nodes)                    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│              PostgreSQL (Primary + 2 Read Replicas)             │
└─────────────────────────────────────────────────────────────────┘
```

### Cloud Provider Recommendations

#### AWS
- **Compute**: EC2 t3.medium (API), c6i.large (Redirect)
- **Database**: RDS PostgreSQL (db.r6g.large)
- **Cache**: ElastiCache Redis (cache.r6g.large)
- **Storage**: S3 for static assets

#### Google Cloud
- **Compute**: e2-standard-2 (API), c2-standard-4 (Redirect)
- **Database**: Cloud SQL PostgreSQL
- **Cache**: Memorystore Redis

#### DigitalOcean
- **Compute**: Premium Intel Droplets
- **Database**: Managed PostgreSQL
- **Cache**: Managed Redis

---

## Server Provisioning

### 1. Base Server Setup (Ubuntu 22.04 LTS)

```bash
#!/bin/bash
# provision-server.sh

set -euo pipefail

# Update system
apt-get update && apt-get upgrade -y

# Install essential packages
apt-get install -y \
    curl \
    wget \
    git \
    htop \
    vim \
    unzip \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    ufw \
    fail2ban

# Configure firewall
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable

# Configure fail2ban
systemctl enable fail2ban
systemctl start fail2ban

# Create linkrift user
useradd -m -s /bin/bash linkrift
usermod -aG sudo linkrift

# Create application directories
mkdir -p /opt/linkrift/{api,redirect,worker,web}
mkdir -p /var/log/linkrift
mkdir -p /etc/linkrift

chown -R linkrift:linkrift /opt/linkrift
chown -R linkrift:linkrift /var/log/linkrift
chown -R linkrift:linkrift /etc/linkrift
```

### 2. Install Go Runtime (for building)

```bash
#!/bin/bash
# install-go.sh

GO_VERSION="1.22.0"

wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
rm -rf /usr/local/go
tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
rm "go${GO_VERSION}.linux-amd64.tar.gz"

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
source /etc/profile.d/go.sh

# Verify installation
go version
```

### 3. Install Node.js (for frontend builds)

```bash
#!/bin/bash
# install-node.sh

curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs

# Install pnpm
npm install -g pnpm

# Verify installation
node --version
pnpm --version
```

### 4. Create Systemd Services

**API Service** (`/etc/systemd/system/linkrift-api.service`):

```ini
[Unit]
Description=Linkrift API Server
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=linkrift
Group=linkrift
WorkingDirectory=/opt/linkrift/api
ExecStart=/opt/linkrift/api/linkrift-api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=linkrift-api

# Environment
EnvironmentFile=/etc/linkrift/api.env

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/linkrift
PrivateTmp=true

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

**Redirect Service** (`/etc/systemd/system/linkrift-redirect.service`):

```ini
[Unit]
Description=Linkrift Redirect Server
After=network.target redis.service
Wants=redis.service

[Service]
Type=simple
User=linkrift
Group=linkrift
WorkingDirectory=/opt/linkrift/redirect
ExecStart=/opt/linkrift/redirect/linkrift-redirect
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=linkrift-redirect

# Environment
EnvironmentFile=/etc/linkrift/redirect.env

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true

# Performance tuning
LimitNOFILE=1048576
LimitNPROC=65535

# CPU affinity for performance
CPUAffinity=0 1

[Install]
WantedBy=multi-user.target
```

**Worker Service** (`/etc/systemd/system/linkrift-worker.service`):

```ini
[Unit]
Description=Linkrift Background Worker
After=network.target postgresql.service redis.service clickhouse.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=linkrift
Group=linkrift
WorkingDirectory=/opt/linkrift/worker
ExecStart=/opt/linkrift/worker/linkrift-worker
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=linkrift-worker

# Environment
EnvironmentFile=/etc/linkrift/worker.env

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/linkrift
PrivateTmp=true

# Resource limits
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 5. Enable and Start Services

```bash
#!/bin/bash
# enable-services.sh

systemctl daemon-reload

systemctl enable linkrift-api
systemctl enable linkrift-redirect
systemctl enable linkrift-worker

systemctl start linkrift-api
systemctl start linkrift-redirect
systemctl start linkrift-worker

# Check status
systemctl status linkrift-api
systemctl status linkrift-redirect
systemctl status linkrift-worker
```

---

## Secrets Management

### Environment Variables Structure

Create environment files with restricted permissions:

```bash
# Create env files
touch /etc/linkrift/{api,redirect,worker}.env
chmod 600 /etc/linkrift/*.env
chown linkrift:linkrift /etc/linkrift/*.env
```

**API Environment** (`/etc/linkrift/api.env`):

```bash
# Server
LINKRIFT_ENV=production
LINKRIFT_API_PORT=8080
LINKRIFT_API_HOST=0.0.0.0

# Database
LINKRIFT_DB_HOST=db.linkrift.internal
LINKRIFT_DB_PORT=5432
LINKRIFT_DB_NAME=linkrift
LINKRIFT_DB_USER=linkrift_api
LINKRIFT_DB_PASSWORD=${DB_PASSWORD}
LINKRIFT_DB_SSL_MODE=require
LINKRIFT_DB_MAX_CONNECTIONS=100
LINKRIFT_DB_MAX_IDLE_CONNECTIONS=25

# Redis
LINKRIFT_REDIS_URL=redis://:${REDIS_PASSWORD}@redis.linkrift.internal:6379/0
LINKRIFT_REDIS_POOL_SIZE=50

# JWT
LINKRIFT_JWT_SECRET=${JWT_SECRET}
LINKRIFT_JWT_EXPIRY=24h

# Rate Limiting
LINKRIFT_RATE_LIMIT_REQUESTS=100
LINKRIFT_RATE_LIMIT_WINDOW=1m

# Logging
LINKRIFT_LOG_LEVEL=info
LINKRIFT_LOG_FORMAT=json
```

**Redirect Environment** (`/etc/linkrift/redirect.env`):

```bash
# Server
LINKRIFT_ENV=production
LINKRIFT_REDIRECT_PORT=8081
LINKRIFT_REDIRECT_HOST=0.0.0.0

# Redis (primary data source for redirects)
LINKRIFT_REDIS_URL=redis://:${REDIS_PASSWORD}@redis.linkrift.internal:6379/0
LINKRIFT_REDIS_POOL_SIZE=200

# Database (fallback)
LINKRIFT_DB_HOST=db.linkrift.internal
LINKRIFT_DB_PORT=5432
LINKRIFT_DB_NAME=linkrift
LINKRIFT_DB_USER=linkrift_redirect
LINKRIFT_DB_PASSWORD=${DB_PASSWORD_READONLY}
LINKRIFT_DB_SSL_MODE=require

# Performance
LINKRIFT_REDIRECT_READ_TIMEOUT=1s
LINKRIFT_REDIRECT_WRITE_TIMEOUT=1s

# Logging
LINKRIFT_LOG_LEVEL=warn
LINKRIFT_LOG_FORMAT=json
```

**Worker Environment** (`/etc/linkrift/worker.env`):

```bash
# Server
LINKRIFT_ENV=production

# Database
LINKRIFT_DB_HOST=db.linkrift.internal
LINKRIFT_DB_PORT=5432
LINKRIFT_DB_NAME=linkrift
LINKRIFT_DB_USER=linkrift_worker
LINKRIFT_DB_PASSWORD=${DB_PASSWORD}
LINKRIFT_DB_SSL_MODE=require

# Redis
LINKRIFT_REDIS_URL=redis://:${REDIS_PASSWORD}@redis.linkrift.internal:6379/1

# ClickHouse
LINKRIFT_CLICKHOUSE_HOST=clickhouse.linkrift.internal
LINKRIFT_CLICKHOUSE_PORT=9000
LINKRIFT_CLICKHOUSE_DATABASE=linkrift_analytics
LINKRIFT_CLICKHOUSE_USER=linkrift
LINKRIFT_CLICKHOUSE_PASSWORD=${CLICKHOUSE_PASSWORD}

# Worker Settings
LINKRIFT_WORKER_CONCURRENCY=10
LINKRIFT_WORKER_BATCH_SIZE=1000
LINKRIFT_WORKER_FLUSH_INTERVAL=5s

# Logging
LINKRIFT_LOG_LEVEL=info
LINKRIFT_LOG_FORMAT=json
```

### Using HashiCorp Vault

For enterprise deployments, use Vault for secrets management:

```bash
#!/bin/bash
# vault-setup.sh

# Install Vault agent
curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list
apt-get update && apt-get install -y vault

# Configure Vault agent
cat > /etc/vault.d/agent.hcl <<EOF
vault {
  address = "https://vault.linkrift.internal:8200"
}

auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/vault.d/role-id"
      secret_id_file_path = "/etc/vault.d/secret-id"
      remove_secret_id_file_after_reading = false
    }
  }

  sink "file" {
    config = {
      path = "/tmp/vault-token"
    }
  }
}

template {
  source = "/etc/linkrift/api.env.tpl"
  destination = "/etc/linkrift/api.env"
  perms = "0600"
  command = "systemctl reload linkrift-api"
}
EOF

# Template file
cat > /etc/linkrift/api.env.tpl <<'EOF'
{{ with secret "secret/data/linkrift/production" }}
LINKRIFT_DB_PASSWORD={{ .Data.data.db_password }}
LINKRIFT_REDIS_PASSWORD={{ .Data.data.redis_password }}
LINKRIFT_JWT_SECRET={{ .Data.data.jwt_secret }}
{{ end }}
EOF
```

### AWS Secrets Manager Integration

```go
// internal/config/secrets_aws.go
package config

import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSSecrets struct {
    DBPassword         string `json:"db_password"`
    RedisPassword      string `json:"redis_password"`
    JWTSecret          string `json:"jwt_secret"`
    ClickHousePassword string `json:"clickhouse_password"`
}

func LoadAWSSecrets(ctx context.Context, secretName string) (*AWSSecrets, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, err
    }

    client := secretsmanager.NewFromConfig(cfg)

    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: &secretName,
    })
    if err != nil {
        return nil, err
    }

    var secrets AWSSecrets
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }

    return &secrets, nil
}
```

---

## SSL/TLS Setup with Cloudflare

### 1. Cloudflare Configuration

#### DNS Setup

```
Type    Name              Content                 Proxy Status
A       linkrift.io       <server-ip>            Proxied (orange)
A       api.linkrift.io   <api-server-ip>        Proxied (orange)
A       go.linkrift.io    <redirect-server-ip>   Proxied (orange)
CNAME   www               linkrift.io            Proxied (orange)
```

#### SSL/TLS Settings

1. **SSL/TLS Mode**: Full (strict)
2. **Edge Certificates**:
   - Always Use HTTPS: On
   - Automatic HTTPS Rewrites: On
   - Minimum TLS Version: TLS 1.2

3. **Origin Server Certificate**:

```bash
# Generate Origin CA certificate via Cloudflare dashboard
# Download and save to server

mkdir -p /etc/ssl/linkrift
# Save certificate as /etc/ssl/linkrift/origin.pem
# Save private key as /etc/ssl/linkrift/origin-key.pem

chmod 600 /etc/ssl/linkrift/*
chown root:root /etc/ssl/linkrift/*
```

### 2. HAProxy Configuration with Cloudflare

```haproxy
# /etc/haproxy/haproxy.cfg

global
    log /dev/log local0
    log /dev/log local1 notice
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

    # SSL configuration
    ssl-default-bind-ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384
    ssl-default-bind-ciphersuites TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256
    ssl-default-bind-options no-sslv3 no-tlsv10 no-tlsv11
    tune.ssl.default-dh-param 2048

defaults
    log global
    mode http
    option httplog
    option dontlognull
    option http-server-close
    option forwardfor except 127.0.0.0/8
    timeout connect 5s
    timeout client 30s
    timeout server 30s
    errorfile 400 /etc/haproxy/errors/400.http
    errorfile 403 /etc/haproxy/errors/403.http
    errorfile 408 /etc/haproxy/errors/408.http
    errorfile 500 /etc/haproxy/errors/500.http
    errorfile 502 /etc/haproxy/errors/502.http
    errorfile 503 /etc/haproxy/errors/503.http
    errorfile 504 /etc/haproxy/errors/504.http

# Cloudflare IP whitelist
acl cloudflare_ips src 173.245.48.0/20
acl cloudflare_ips src 103.21.244.0/22
acl cloudflare_ips src 103.22.200.0/22
acl cloudflare_ips src 103.31.4.0/22
acl cloudflare_ips src 141.101.64.0/18
acl cloudflare_ips src 108.162.192.0/18
acl cloudflare_ips src 190.93.240.0/20
acl cloudflare_ips src 188.114.96.0/20
acl cloudflare_ips src 197.234.240.0/22
acl cloudflare_ips src 198.41.128.0/17
acl cloudflare_ips src 162.158.0.0/15
acl cloudflare_ips src 104.16.0.0/13
acl cloudflare_ips src 104.24.0.0/14
acl cloudflare_ips src 172.64.0.0/13
acl cloudflare_ips src 131.0.72.0/22

# Frontend - HTTPS
frontend https_front
    bind *:443 ssl crt /etc/ssl/linkrift/combined.pem

    # Only accept Cloudflare traffic
    http-request deny unless cloudflare_ips

    # Get real client IP from Cloudflare
    http-request set-header X-Real-IP %[req.hdr(CF-Connecting-IP)]

    # Route based on host
    acl host_api hdr(host) -i api.linkrift.io
    acl host_redirect hdr(host) -i go.linkrift.io
    acl host_web hdr(host) -i linkrift.io
    acl host_web hdr(host) -i www.linkrift.io

    use_backend api_servers if host_api
    use_backend redirect_servers if host_redirect
    use_backend web_servers if host_web

    default_backend web_servers

# Backend - API
backend api_servers
    balance roundrobin
    option httpchk GET /health
    http-check expect status 200

    server api1 10.0.1.10:8080 check inter 5s fall 3 rise 2
    server api2 10.0.1.11:8080 check inter 5s fall 3 rise 2
    server api3 10.0.1.12:8080 check inter 5s fall 3 rise 2 backup

# Backend - Redirect (optimized for latency)
backend redirect_servers
    balance leastconn
    option httpchk GET /health
    http-check expect status 200

    server redirect1 10.0.2.10:8081 check inter 2s fall 2 rise 1
    server redirect2 10.0.2.11:8081 check inter 2s fall 2 rise 1
    server redirect3 10.0.2.12:8081 check inter 2s fall 2 rise 1

# Backend - Web
backend web_servers
    balance roundrobin
    option httpchk GET /
    http-check expect status 200

    server web1 10.0.3.10:80 check inter 10s fall 3 rise 2
    server web2 10.0.3.11:80 check inter 10s fall 3 rise 2

# Stats
listen stats
    bind 127.0.0.1:8404
    stats enable
    stats uri /stats
    stats refresh 10s
    stats auth admin:${HAPROXY_STATS_PASSWORD}
```

### 3. Cloudflare Worker for Edge Caching (Optional)

```javascript
// cloudflare-worker.js
// Deploy via Cloudflare Workers for edge-level redirect caching

const CACHE_TTL = 3600; // 1 hour

addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const url = new URL(request.url);

  // Only handle redirect domain
  if (url.hostname !== 'go.linkrift.io') {
    return fetch(request);
  }

  const shortCode = url.pathname.slice(1); // Remove leading /

  if (!shortCode || shortCode.includes('/')) {
    return fetch(request);
  }

  // Check edge cache
  const cache = caches.default;
  const cacheKey = new Request(url.toString(), request);
  let response = await cache.match(cacheKey);

  if (response) {
    // Cache hit - add header for debugging
    response = new Response(response.body, response);
    response.headers.set('X-Cache', 'HIT');
    return response;
  }

  // Cache miss - fetch from origin
  response = await fetch(request);

  // Only cache successful redirects
  if (response.status === 301 || response.status === 302) {
    const cacheResponse = new Response(response.body, response);
    cacheResponse.headers.set('Cache-Control', `public, max-age=${CACHE_TTL}`);
    cacheResponse.headers.set('X-Cache', 'MISS');

    // Store in cache
    event.waitUntil(cache.put(cacheKey, cacheResponse.clone()));

    return cacheResponse;
  }

  return response;
}
```

---

## Database Migration in Production

### Migration Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                    Migration Flow                                │
│                                                                  │
│  1. Create migration file                                        │
│  2. Test in staging                                              │
│  3. Take database snapshot                                       │
│  4. Run migration with lock timeout                              │
│  5. Verify migration                                             │
│  6. Monitor application                                          │
└─────────────────────────────────────────────────────────────────┘
```

### Using golang-migrate

```bash
# Install migrate CLI
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /usr/local/bin/

# Create migration
migrate create -ext sql -dir migrations -seq add_click_analytics
```

**Migration Files Structure**:

```
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_links.up.sql
├── 000002_create_links.down.sql
├── 000003_create_clicks.up.sql
├── 000003_create_clicks.down.sql
├── 000004_add_link_indexes.up.sql
├── 000004_add_link_indexes.down.sql
└── ...
```

**Example Migration** (`000002_create_links.up.sql`):

```sql
-- Create links table
CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    short_code VARCHAR(16) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    title VARCHAR(255),
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,
    password_hash VARCHAR(255),
    click_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_links_user_id ON links(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_links_short_code ON links(short_code);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_links_created_at ON links(created_at DESC);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_links_updated_at
    BEFORE UPDATE ON links
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Production Migration Script

```bash
#!/bin/bash
# migrate-production.sh

set -euo pipefail

# Configuration
DB_HOST="${LINKRIFT_DB_HOST}"
DB_PORT="${LINKRIFT_DB_PORT:-5432}"
DB_NAME="${LINKRIFT_DB_NAME}"
DB_USER="${LINKRIFT_DB_USER}"
DB_PASSWORD="${LINKRIFT_DB_PASSWORD}"
MIGRATIONS_PATH="/opt/linkrift/migrations"

# Build connection string
DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

echo "=== Linkrift Database Migration ==="
echo "Database: ${DB_NAME}@${DB_HOST}"
echo "Migrations: ${MIGRATIONS_PATH}"
echo ""

# Check current version
echo "Current migration version:"
migrate -path "${MIGRATIONS_PATH}" -database "${DB_URL}" version || echo "No migrations applied yet"
echo ""

# Create snapshot
echo "Creating database snapshot..."
SNAPSHOT_NAME="linkrift_pre_migration_$(date +%Y%m%d_%H%M%S)"
pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" \
    --format=custom --compress=9 \
    --file="/var/backups/linkrift/${SNAPSHOT_NAME}.dump"
echo "Snapshot created: ${SNAPSHOT_NAME}"
echo ""

# Set lock timeout
export PGOPTIONS="-c lock_timeout=30000 -c statement_timeout=300000"

# Run migrations
echo "Running migrations..."
migrate -path "${MIGRATIONS_PATH}" -database "${DB_URL}" up

# Verify
echo ""
echo "New migration version:"
migrate -path "${MIGRATIONS_PATH}" -database "${DB_URL}" version

echo ""
echo "=== Migration Complete ==="
```

### Zero-Downtime Migration Patterns

```sql
-- Pattern 1: Add column with default (PostgreSQL 11+, instant)
ALTER TABLE links ADD COLUMN utm_source VARCHAR(255);

-- Pattern 2: Add NOT NULL column safely
-- Step 1: Add nullable column
ALTER TABLE links ADD COLUMN workspace_id UUID;

-- Step 2: Backfill data (in batches)
UPDATE links SET workspace_id = (
    SELECT default_workspace_id FROM users WHERE users.id = links.user_id
)
WHERE workspace_id IS NULL
LIMIT 10000;

-- Step 3: Add NOT NULL constraint
ALTER TABLE links ALTER COLUMN workspace_id SET NOT NULL;

-- Pattern 3: Create index concurrently (non-blocking)
CREATE INDEX CONCURRENTLY idx_links_workspace_id ON links(workspace_id);

-- Pattern 4: Rename column safely
-- Step 1: Add new column
ALTER TABLE links ADD COLUMN destination_url TEXT;

-- Step 2: Create trigger to keep in sync
CREATE OR REPLACE FUNCTION sync_url_columns()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        IF NEW.destination_url IS NULL THEN
            NEW.destination_url = NEW.original_url;
        ELSIF NEW.original_url IS NULL OR NEW.original_url != NEW.destination_url THEN
            NEW.original_url = NEW.destination_url;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_links_url
    BEFORE INSERT OR UPDATE ON links
    FOR EACH ROW
    EXECUTE FUNCTION sync_url_columns();

-- Step 3: Backfill existing data
UPDATE links SET destination_url = original_url WHERE destination_url IS NULL;

-- Step 4: Deploy application using new column name

-- Step 5: Drop old column (later migration)
DROP TRIGGER sync_links_url ON links;
DROP FUNCTION sync_url_columns();
ALTER TABLE links DROP COLUMN original_url;
```

---

## Go Binary Build Process

### Build Configuration

**Makefile**:

```makefile
# Makefile

# Variables
APP_NAME := linkrift
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-s -w \
    -X main.Version=$(VERSION) \
    -X main.BuildTime=$(BUILD_TIME) \
    -X main.Commit=$(COMMIT) \
    -X main.Branch=$(BRANCH)"

# Directories
BUILD_DIR := ./build
CMD_DIR := ./cmd

# Targets
.PHONY: all build clean test lint

all: clean lint test build

# Build all services
build: build-api build-redirect build-worker

build-api:
	@echo "Building API server..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(BUILD_DIR)/linkrift-api $(CMD_DIR)/api

build-redirect:
	@echo "Building Redirect server..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(BUILD_DIR)/linkrift-redirect $(CMD_DIR)/redirect

build-worker:
	@echo "Building Worker..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(BUILD_DIR)/linkrift-worker $(CMD_DIR)/worker

# Build for multiple platforms
build-all-platforms:
	@echo "Building for all platforms..."
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) \
				-o $(BUILD_DIR)/linkrift-api-$$os-$$arch $(CMD_DIR)/api; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) \
				-o $(BUILD_DIR)/linkrift-redirect-$$os-$$arch $(CMD_DIR)/redirect; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) \
				-o $(BUILD_DIR)/linkrift-worker-$$os-$$arch $(CMD_DIR)/worker; \
		done \
	done

# Testing
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-integration:
	$(GOTEST) -v -tags=integration ./...

# Linting
lint:
	golangci-lint run ./...

# Clean
clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)

# Dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Branch: $(BRANCH)"
	@echo "Build Time: $(BUILD_TIME)"
```

### Build Script for CI/CD

```bash
#!/bin/bash
# build.sh

set -euo pipefail

echo "=== Linkrift Build Process ==="

# Environment
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# Version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse --short HEAD)

echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Build Time: ${BUILD_TIME}"
echo ""

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${COMMIT}"

# Create build directory
BUILD_DIR="./build"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Build each service
echo "Building API server..."
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/linkrift-api" ./cmd/api

echo "Building Redirect server..."
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/linkrift-redirect" ./cmd/redirect

echo "Building Worker..."
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/linkrift-worker" ./cmd/worker

# Generate checksums
echo ""
echo "Generating checksums..."
cd "${BUILD_DIR}"
sha256sum linkrift-* > checksums.sha256
cat checksums.sha256

echo ""
echo "=== Build Complete ==="
ls -lh "${BUILD_DIR}"/
```

### Version Embedding in Go

```go
// cmd/api/main.go
package main

import (
    "fmt"
    "os"
)

// Build-time variables
var (
    Version   = "dev"
    BuildTime = "unknown"
    Commit    = "unknown"
    Branch    = "unknown"
)

func main() {
    // Handle version flag
    if len(os.Args) > 1 && os.Args[1] == "version" {
        fmt.Printf("Linkrift API Server\n")
        fmt.Printf("  Version:    %s\n", Version)
        fmt.Printf("  Commit:     %s\n", Commit)
        fmt.Printf("  Branch:     %s\n", Branch)
        fmt.Printf("  Build Time: %s\n", BuildTime)
        os.Exit(0)
    }

    // Start server...
}
```

---

## Vite Production Build

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { compression } from 'vite-plugin-compression2';
import { visualizer } from 'rollup-plugin-visualizer';

export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    compression({
      algorithm: 'gzip',
      exclude: [/\.(br)$/, /\.(gz)$/],
    }),
    compression({
      algorithm: 'brotliCompress',
      exclude: [/\.(br)$/, /\.(gz)$/],
    }),
    mode === 'analyze' && visualizer({
      open: true,
      filename: 'bundle-analysis.html',
    }),
  ].filter(Boolean),

  build: {
    outDir: 'dist',
    sourcemap: mode === 'production' ? 'hidden' : true,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          charts: ['recharts'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
        },
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      },
    },
    chunkSizeWarningLimit: 500,
  },

  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
    __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
  },

  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
}));
```

### Frontend Build Script

```bash
#!/bin/bash
# build-frontend.sh

set -euo pipefail

echo "=== Linkrift Frontend Build ==="

cd /opt/linkrift/web

# Install dependencies
echo "Installing dependencies..."
pnpm install --frozen-lockfile

# Type checking
echo "Running type check..."
pnpm tsc --noEmit

# Linting
echo "Running linter..."
pnpm eslint src --ext .ts,.tsx --max-warnings 0

# Run tests
echo "Running tests..."
pnpm test --run --coverage

# Build for production
echo "Building for production..."
NODE_ENV=production pnpm build

# Verify build
echo ""
echo "Build output:"
ls -lh dist/

# Check bundle size
echo ""
echo "Bundle analysis:"
du -sh dist/
find dist -name "*.js" -exec du -h {} \; | sort -rh | head -10

echo ""
echo "=== Frontend Build Complete ==="
```

### Nginx Configuration for Frontend

```nginx
# /etc/nginx/sites-available/linkrift-web

server {
    listen 80;
    server_name linkrift.io www.linkrift.io;

    root /opt/linkrift/web/dist;
    index index.html;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;

    # Brotli compression (if module available)
    brotli on;
    brotli_comp_level 6;
    brotli_types text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' https://api.linkrift.io;" always;

    # Cache static assets
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Cache favicon and robots
    location ~* \.(ico|txt)$ {
        expires 1M;
        add_header Cache-Control "public";
    }

    # SPA routing - serve index.html for all routes
    location / {
        try_files $uri $uri/ /index.html;

        # Don't cache index.html
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    # Health check
    location /health {
        access_log off;
        return 200 "OK";
        add_header Content-Type text/plain;
    }
}
```

---

## Zero-Downtime Deployment

### Deployment Strategy Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                  Zero-Downtime Deployment Flow                    │
│                                                                   │
│  1. Build new version                                             │
│  2. Run database migrations                                       │
│  3. Deploy to canary server (10% traffic)                        │
│  4. Monitor metrics for 5 minutes                                 │
│  5. Gradually increase traffic (25%, 50%, 100%)                  │
│  6. Remove old version                                            │
└──────────────────────────────────────────────────────────────────┘
```

### Blue-Green Deployment Script

```bash
#!/bin/bash
# deploy.sh

set -euo pipefail

# Configuration
DEPLOY_USER="linkrift"
DEPLOY_HOST="${1:-}"
SERVICE="${2:-api}"
VERSION="${3:-latest}"

if [[ -z "${DEPLOY_HOST}" ]]; then
    echo "Usage: $0 <host> <service> [version]"
    exit 1
fi

echo "=== Linkrift Deployment ==="
echo "Host: ${DEPLOY_HOST}"
echo "Service: ${SERVICE}"
echo "Version: ${VERSION}"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Step 1: Upload new binary
log_info "Uploading new binary..."
scp "./build/linkrift-${SERVICE}" "${DEPLOY_USER}@${DEPLOY_HOST}:/opt/linkrift/${SERVICE}/linkrift-${SERVICE}.new"

# Step 2: Health check current service
log_info "Checking current service health..."
ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "curl -sf http://localhost:8080/health || true"

# Step 3: Atomic swap
log_info "Performing atomic binary swap..."
ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << 'EOF'
set -euo pipefail

SERVICE_DIR="/opt/linkrift/${SERVICE}"
BINARY="${SERVICE_DIR}/linkrift-${SERVICE}"

# Backup current binary
if [[ -f "${BINARY}" ]]; then
    cp "${BINARY}" "${BINARY}.backup"
fi

# Atomic move
mv "${BINARY}.new" "${BINARY}"
chmod +x "${BINARY}"
EOF

# Step 4: Graceful restart
log_info "Performing graceful restart..."
ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << 'EOF'
set -euo pipefail

SERVICE_NAME="linkrift-${SERVICE}"

# Send SIGUSR2 for graceful restart (if supported)
# Otherwise, use systemd reload
systemctl reload "${SERVICE_NAME}" 2>/dev/null || systemctl restart "${SERVICE_NAME}"
EOF

# Step 5: Wait for service to be ready
log_info "Waiting for service to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0

while [[ ${RETRY_COUNT} -lt ${MAX_RETRIES} ]]; do
    if ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "curl -sf http://localhost:8080/health" > /dev/null 2>&1; then
        log_info "Service is healthy!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    sleep 1
done

if [[ ${RETRY_COUNT} -ge ${MAX_RETRIES} ]]; then
    log_error "Service failed to become healthy!"
    log_warn "Initiating rollback..."
    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << 'EOF'
    BINARY="/opt/linkrift/${SERVICE}/linkrift-${SERVICE}"
    if [[ -f "${BINARY}.backup" ]]; then
        mv "${BINARY}.backup" "${BINARY}"
        systemctl restart "linkrift-${SERVICE}"
    fi
EOF
    exit 1
fi

# Step 6: Verify deployment
log_info "Verifying deployment..."
DEPLOYED_VERSION=$(ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "/opt/linkrift/${SERVICE}/linkrift-${SERVICE} version 2>/dev/null | grep Version | awk '{print \$2}'" || echo "unknown")
log_info "Deployed version: ${DEPLOYED_VERSION}"

# Step 7: Cleanup
log_info "Cleaning up..."
ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "rm -f /opt/linkrift/${SERVICE}/linkrift-${SERVICE}.backup"

echo ""
log_info "=== Deployment Complete ==="
```

### Rolling Deployment with Multiple Servers

```bash
#!/bin/bash
# rolling-deploy.sh

set -euo pipefail

# Server list
SERVERS=(
    "server1.linkrift.internal"
    "server2.linkrift.internal"
    "server3.linkrift.internal"
)

SERVICE="${1:-api}"
VERSION="${2:-latest}"
HEALTH_CHECK_URL="http://localhost:8080/health"
LB_API="http://haproxy.linkrift.internal:8404"

log_info() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1"; }
log_error() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1"; }

# Drain server from load balancer
drain_server() {
    local server=$1
    log_info "Draining ${server} from load balancer..."
    curl -sf -X POST "${LB_API}/servers/${server}/drain" || true
    sleep 10  # Wait for connections to drain
}

# Enable server in load balancer
enable_server() {
    local server=$1
    log_info "Enabling ${server} in load balancer..."
    curl -sf -X POST "${LB_API}/servers/${server}/enable" || true
}

# Deploy to single server
deploy_server() {
    local server=$1

    log_info "Deploying to ${server}..."

    # 1. Drain from LB
    drain_server "${server}"

    # 2. Deploy
    ./deploy.sh "${server}" "${SERVICE}" "${VERSION}"

    # 3. Health check
    local retries=0
    while [[ ${retries} -lt 30 ]]; do
        if ssh "linkrift@${server}" "curl -sf ${HEALTH_CHECK_URL}" > /dev/null 2>&1; then
            log_info "${server} is healthy"
            break
        fi
        retries=$((retries + 1))
        sleep 1
    done

    if [[ ${retries} -ge 30 ]]; then
        log_error "${server} failed health check!"
        return 1
    fi

    # 4. Re-enable in LB
    enable_server "${server}"

    # 5. Wait for traffic
    sleep 5

    log_info "${server} deployment complete"
}

# Main deployment loop
log_info "Starting rolling deployment of ${SERVICE} version ${VERSION}"
log_info "Servers: ${SERVERS[*]}"

for server in "${SERVERS[@]}"; do
    if ! deploy_server "${server}"; then
        log_error "Deployment failed at ${server}. Stopping rollout."
        exit 1
    fi
done

log_info "Rolling deployment complete!"
```

### Graceful Shutdown Handler (Go)

```go
// internal/server/graceful.go
package server

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"
)

type GracefulServer struct {
    server   *http.Server
    logger   *zap.Logger
    shutdown chan struct{}
}

func NewGracefulServer(addr string, handler http.Handler, logger *zap.Logger) *GracefulServer {
    return &GracefulServer{
        server: &http.Server{
            Addr:         addr,
            Handler:      handler,
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
        },
        logger:   logger,
        shutdown: make(chan struct{}),
    }
}

func (gs *GracefulServer) Start() error {
    // Channel to listen for errors from server
    errChan := make(chan error, 1)

    // Start server in goroutine
    go func() {
        gs.logger.Info("Starting server", zap.String("addr", gs.server.Addr))
        if err := gs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    // Listen for shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

    select {
    case err := <-errChan:
        return err
    case sig := <-sigChan:
        gs.logger.Info("Received signal", zap.String("signal", sig.String()))
        return gs.Shutdown()
    }
}

func (gs *GracefulServer) Shutdown() error {
    gs.logger.Info("Initiating graceful shutdown...")

    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Stop accepting new connections
    gs.server.SetKeepAlivesEnabled(false)

    // Wait for existing connections to finish
    if err := gs.server.Shutdown(ctx); err != nil {
        gs.logger.Error("Server shutdown error", zap.Error(err))
        return err
    }

    gs.logger.Info("Server gracefully stopped")
    close(gs.shutdown)
    return nil
}

func (gs *GracefulServer) Done() <-chan struct{} {
    return gs.shutdown
}
```

---

## Rollback Procedures

### Automated Rollback Script

```bash
#!/bin/bash
# rollback.sh

set -euo pipefail

DEPLOY_HOST="${1:-}"
SERVICE="${2:-api}"
TARGET_VERSION="${3:-previous}"

if [[ -z "${DEPLOY_HOST}" ]]; then
    echo "Usage: $0 <host> <service> [version|previous]"
    exit 1
fi

echo "=== Linkrift Rollback ==="
echo "Host: ${DEPLOY_HOST}"
echo "Service: ${SERVICE}"
echo "Target: ${TARGET_VERSION}"
echo ""

# Get available versions
ssh "linkrift@${DEPLOY_HOST}" << 'VERSIONS'
echo "Available versions:"
ls -la /opt/linkrift/releases/${SERVICE}/ | tail -10
VERSIONS

# Perform rollback
if [[ "${TARGET_VERSION}" == "previous" ]]; then
    ssh "linkrift@${DEPLOY_HOST}" << 'ROLLBACK'
    set -euo pipefail

    SERVICE_DIR="/opt/linkrift/${SERVICE}"
    RELEASES_DIR="/opt/linkrift/releases/${SERVICE}"

    # Get previous version
    CURRENT=$(readlink "${SERVICE_DIR}/current" | xargs basename)
    PREVIOUS=$(ls -t "${RELEASES_DIR}" | grep -v "${CURRENT}" | head -1)

    if [[ -z "${PREVIOUS}" ]]; then
        echo "No previous version found!"
        exit 1
    fi

    echo "Rolling back from ${CURRENT} to ${PREVIOUS}..."

    # Update symlink
    ln -sfn "${RELEASES_DIR}/${PREVIOUS}" "${SERVICE_DIR}/current"

    # Restart service
    systemctl restart "linkrift-${SERVICE}"

    echo "Rollback complete!"
ROLLBACK
else
    ssh "linkrift@${DEPLOY_HOST}" << ROLLBACK
    set -euo pipefail

    SERVICE_DIR="/opt/linkrift/${SERVICE}"
    RELEASES_DIR="/opt/linkrift/releases/${SERVICE}"
    TARGET="${TARGET_VERSION}"

    if [[ ! -d "${RELEASES_DIR}/${TARGET}" ]]; then
        echo "Version ${TARGET} not found!"
        exit 1
    fi

    echo "Rolling back to ${TARGET}..."

    # Update symlink
    ln -sfn "${RELEASES_DIR}/${TARGET}" "${SERVICE_DIR}/current"

    # Restart service
    systemctl restart "linkrift-${SERVICE}"

    echo "Rollback complete!"
ROLLBACK
fi

# Verify
echo ""
echo "Verifying rollback..."
ssh "linkrift@${DEPLOY_HOST}" "curl -sf http://localhost:8080/health && /opt/linkrift/${SERVICE}/current/linkrift-${SERVICE} version"
```

### Database Rollback

```bash
#!/bin/bash
# rollback-db.sh

set -euo pipefail

TARGET_VERSION="${1:-}"

if [[ -z "${TARGET_VERSION}" ]]; then
    echo "Usage: $0 <migration-version>"
    echo ""
    echo "Current version:"
    migrate -path ./migrations -database "${DATABASE_URL}" version
    exit 1
fi

echo "=== Database Rollback ==="
echo "Target Version: ${TARGET_VERSION}"
echo ""

# Show what will be rolled back
echo "Migrations to rollback:"
migrate -path ./migrations -database "${DATABASE_URL}" version
echo " -> ${TARGET_VERSION}"
echo ""

read -p "Are you sure you want to rollback? (yes/no): " CONFIRM
if [[ "${CONFIRM}" != "yes" ]]; then
    echo "Rollback cancelled."
    exit 0
fi

# Create backup first
BACKUP_FILE="/var/backups/linkrift/pre_rollback_$(date +%Y%m%d_%H%M%S).dump"
echo "Creating backup: ${BACKUP_FILE}"
pg_dump --format=custom --compress=9 --file="${BACKUP_FILE}" "${DATABASE_URL}"

# Rollback
echo "Rolling back migrations..."
migrate -path ./migrations -database "${DATABASE_URL}" goto "${TARGET_VERSION}"

echo ""
echo "New version:"
migrate -path ./migrations -database "${DATABASE_URL}" version

echo ""
echo "=== Rollback Complete ==="
```

### Emergency Rollback Runbook

```markdown
## Emergency Rollback Runbook

### Symptoms Requiring Rollback
- [ ] Error rate > 5%
- [ ] P99 latency > 100ms (redirect service)
- [ ] P99 latency > 500ms (API service)
- [ ] Memory leak detected
- [ ] Database connection exhaustion

### Immediate Actions

1. **Alert Team**
   ```bash
   # Post to Slack
   curl -X POST -H 'Content-type: application/json' \
     --data '{"text":"🚨 EMERGENCY ROLLBACK INITIATED - Linkrift Production"}' \
     $SLACK_WEBHOOK_URL
   ```

2. **Check Current State**
   ```bash
   # Service health
   for server in server{1..3}.linkrift.internal; do
     echo "=== $server ==="
     ssh linkrift@$server "systemctl status linkrift-api; curl -sf localhost:8080/health"
   done
   ```

3. **Execute Rollback**
   ```bash
   # Rollback all servers
   ./rollback.sh server1.linkrift.internal api previous
   ./rollback.sh server2.linkrift.internal api previous
   ./rollback.sh server3.linkrift.internal api previous
   ```

4. **Verify Recovery**
   ```bash
   # Check error rates
   curl -s "http://prometheus:9090/api/v1/query?query=rate(http_requests_total{status=~\"5..\"}[5m])"

   # Check latency
   curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.99,rate(http_request_duration_seconds_bucket[5m]))"
   ```

5. **Post-Incident**
   - [ ] Document timeline
   - [ ] Identify root cause
   - [ ] Create incident report
   - [ ] Schedule post-mortem
```

---

## Deployment Checklist

### Pre-Deployment
- [ ] All tests passing in CI
- [ ] Security scan completed
- [ ] Database migrations tested in staging
- [ ] Rollback procedure verified
- [ ] On-call engineer notified
- [ ] Deployment window confirmed

### During Deployment
- [ ] Database backup created
- [ ] Monitoring dashboards open
- [ ] Error rates baseline noted
- [ ] Canary deployed and verified
- [ ] Gradual rollout completed

### Post-Deployment
- [ ] All health checks passing
- [ ] Error rates within normal range
- [ ] No memory leaks detected
- [ ] Customer-facing features verified
- [ ] Deployment documented
- [ ] Old releases cleaned up

---

## Monitoring During Deployment

### Key Metrics to Watch

```promql
# Error rate
sum(rate(http_requests_total{status=~"5.."}[1m])) / sum(rate(http_requests_total[1m]))

# Redirect latency P99
histogram_quantile(0.99, sum(rate(redirect_duration_seconds_bucket[1m])) by (le))

# Active connections
sum(linkrift_active_connections)

# Database connections
sum(pg_stat_activity_count{datname="linkrift"})

# Redis memory
redis_memory_used_bytes

# CPU usage
avg(rate(process_cpu_seconds_total[1m])) by (service)
```

### Alerting Rules for Deployment

```yaml
# prometheus/alerts/deployment.yml
groups:
  - name: deployment
    rules:
      - alert: HighErrorRateDuringDeployment
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[2m]))
          / sum(rate(http_requests_total[2m])) > 0.05
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected during deployment
          description: Error rate is {{ $value | humanizePercentage }}

      - alert: LatencySpikeDuringDeployment
        expr: |
          histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[2m])) by (le)) > 0.5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: Latency spike detected during deployment
          description: P99 latency is {{ $value | humanizeDuration }}
```

---

*This deployment guide is maintained by the Linkrift Platform Team. For questions or updates, contact platform@linkrift.io*
