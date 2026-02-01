# White-Label Solution

> Last Updated: 2025-01-24

Complete guide for deploying Linkrift as a white-label URL shortening solution for enterprise customers.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Multi-Tenant Architecture](#multi-tenant-architecture)
  - [Deployment Options](#deployment-options)
  - [Infrastructure Components](#infrastructure-components)
- [Customization Options](#customization-options)
  - [Branding](#branding)
  - [Domain Configuration](#domain-configuration)
  - [Feature Toggles](#feature-toggles)
  - [Custom Integrations](#custom-integrations)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Tenant Configuration](#tenant-configuration)
  - [Theme Configuration](#theme-configuration)
- [Enterprise Deployment](#enterprise-deployment)
  - [Self-Hosted Deployment](#self-hosted-deployment)
  - [Dedicated Cloud Instance](#dedicated-cloud-instance)
  - [Hybrid Deployment](#hybrid-deployment)
- [API Customization](#api-customization)
- [Security & Compliance](#security--compliance)
- [Monitoring & Analytics](#monitoring--analytics)
- [Support & SLA](#support--sla)

---

## Overview

Linkrift White-Label allows organizations to deploy a fully customized URL shortening service under their own brand. This solution is ideal for:

- **Enterprises** needing branded short links for marketing
- **SaaS platforms** adding URL shortening as a feature
- **Agencies** offering link management to clients
- **Resellers** building URL shortening businesses

**Key Benefits:**

| Benefit | Description |
|---------|-------------|
| Full Branding | Complete control over look and feel |
| Custom Domains | Use your own domains for short URLs |
| Data Ownership | Keep all data in your infrastructure |
| Feature Control | Enable/disable features per tenant |
| API Flexibility | White-label API endpoints |
| Compliance | Meet regulatory requirements |

---

## Architecture

### Multi-Tenant Architecture

Linkrift supports multiple tenancy models:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Load Balancer / CDN                          │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   Tenant A    │     │   Tenant B    │     │   Tenant C    │
│  (Isolated)   │     │  (Isolated)   │     │  (Shared)     │
├───────────────┤     ├───────────────┤     ├───────────────┤
│ • custom-a.co │     │ • links.b.io  │     │ • lnk.co/c    │
│ • Dedicated   │     │ • Dedicated   │     │ • Shared      │
│   Database    │     │   Database    │     │   Database    │
│ • Custom UI   │     │ • Custom UI   │     │ • Themed UI   │
└───────────────┘     └───────────────┘     └───────────────┘
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  PostgreSQL   │     │  PostgreSQL   │     │   Shared      │
│  (Dedicated)  │     │  (Dedicated)  │     │  PostgreSQL   │
└───────────────┘     └───────────────┘     └───────────────┘
```

**Tenancy Models:**

1. **Fully Isolated**: Separate infrastructure per tenant
2. **Database Isolated**: Shared compute, separate databases
3. **Schema Isolated**: Shared database, separate schemas
4. **Row-Level Isolation**: Shared everything with tenant_id filtering

### Deployment Options

| Option | Best For | Pros | Cons |
|--------|----------|------|------|
| **SaaS Multi-tenant** | Small/Medium tenants | Cost-effective, managed | Limited customization |
| **Dedicated Instance** | Large enterprises | Full isolation, customization | Higher cost |
| **Self-Hosted** | Regulated industries | Complete control | Operational overhead |
| **Hybrid** | Complex requirements | Flexibility | Complexity |

### Infrastructure Components

```yaml
# infrastructure/docker-compose.white-label.yml
version: '3.8'

services:
  # API Server (Go)
  api:
    image: linkrift/api:${VERSION}
    environment:
      - TENANT_MODE=multi
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
    deploy:
      replicas: 3
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.api.rule=PathPrefix(`/api`)"

  # Web Frontend
  web:
    image: linkrift/web:${VERSION}
    environment:
      - TENANT_CONFIG_URL=${TENANT_CONFIG_URL}
    deploy:
      replicas: 2

  # Redirect Service (High Performance)
  redirect:
    image: linkrift/redirect:${VERSION}
    environment:
      - REDIS_URL=${REDIS_URL}
      - CACHE_TTL=3600
    deploy:
      replicas: 5

  # Analytics Processor
  analytics:
    image: linkrift/analytics:${VERSION}
    environment:
      - CLICKHOUSE_URL=${CLICKHOUSE_URL}

  # Database
  postgres:
    image: postgres:16
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=${DB_PASSWORD}

  # Cache
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  # Analytics Database
  clickhouse:
    image: clickhouse/clickhouse-server:24
    volumes:
      - clickhouse_data:/var/lib/clickhouse

volumes:
  postgres_data:
  redis_data:
  clickhouse_data:
```

---

## Customization Options

### Branding

Full control over visual branding:

```typescript
// config/branding.ts
export interface BrandingConfig {
  // Identity
  companyName: string;
  productName: string;
  tagline?: string;

  // Logos
  logo: {
    light: string;    // URL or base64
    dark: string;
    favicon: string;
    appleTouchIcon: string;
  };

  // Colors
  colors: {
    primary: string;
    secondary: string;
    accent: string;
    background: string;
    surface: string;
    text: {
      primary: string;
      secondary: string;
      disabled: string;
    };
    error: string;
    warning: string;
    success: string;
    info: string;
  };

  // Typography
  typography: {
    fontFamily: string;
    headingFontFamily?: string;
    fontSize: {
      xs: string;
      sm: string;
      base: string;
      lg: string;
      xl: string;
      '2xl': string;
      '3xl': string;
    };
  };

  // Components
  components: {
    borderRadius: string;
    buttonStyle: 'rounded' | 'square' | 'pill';
    cardShadow: string;
  };

  // Content
  content: {
    supportEmail: string;
    privacyPolicyUrl: string;
    termsOfServiceUrl: string;
    documentationUrl?: string;
  };
}

// Example configuration
export const tenantBranding: BrandingConfig = {
  companyName: 'Acme Corp',
  productName: 'Acme Links',
  tagline: 'Shorten, Share, Succeed',

  logo: {
    light: '/assets/acme-logo-light.svg',
    dark: '/assets/acme-logo-dark.svg',
    favicon: '/assets/acme-favicon.ico',
    appleTouchIcon: '/assets/acme-apple-touch.png',
  },

  colors: {
    primary: '#2563eb',
    secondary: '#7c3aed',
    accent: '#f59e0b',
    background: '#ffffff',
    surface: '#f8fafc',
    text: {
      primary: '#0f172a',
      secondary: '#475569',
      disabled: '#94a3b8',
    },
    error: '#ef4444',
    warning: '#f59e0b',
    success: '#22c55e',
    info: '#3b82f6',
  },

  typography: {
    fontFamily: "'Inter', -apple-system, sans-serif",
    headingFontFamily: "'Poppins', sans-serif",
    fontSize: {
      xs: '0.75rem',
      sm: '0.875rem',
      base: '1rem',
      lg: '1.125rem',
      xl: '1.25rem',
      '2xl': '1.5rem',
      '3xl': '1.875rem',
    },
  },

  components: {
    borderRadius: '0.5rem',
    buttonStyle: 'rounded',
    cardShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },

  content: {
    supportEmail: 'support@acme.com',
    privacyPolicyUrl: 'https://acme.com/privacy',
    termsOfServiceUrl: 'https://acme.com/terms',
    documentationUrl: 'https://docs.acme.com/links',
  },
};
```

### Domain Configuration

```yaml
# config/domains.yml
tenant_id: acme-corp

# Primary short domain
primary_domain:
  domain: acme.link
  ssl:
    enabled: true
    provider: letsencrypt
  dns:
    type: CNAME
    target: redirect.linkrift.io

# Additional domains
custom_domains:
  - domain: go.acme.com
    ssl:
      enabled: true
      certificate: /certs/acme.com.pem
      key: /certs/acme.com.key
    routing:
      default_workspace: marketing

  - domain: links.acme.io
    ssl:
      enabled: true
      provider: letsencrypt
    routing:
      default_workspace: engineering

# Dashboard domain
dashboard_domain:
  domain: links-admin.acme.com
  ssl:
    enabled: true
    provider: cloudflare
```

### Feature Toggles

```go
// internal/config/features.go
package config

type FeatureFlags struct {
    // Core Features
    URLShortening     bool `json:"url_shortening"`
    CustomShortCodes  bool `json:"custom_short_codes"`
    BulkOperations    bool `json:"bulk_operations"`
    LinkExpiration    bool `json:"link_expiration"`
    PasswordProtected bool `json:"password_protected"`

    // Analytics
    BasicAnalytics    bool `json:"basic_analytics"`
    AdvancedAnalytics bool `json:"advanced_analytics"`
    RealTimeAnalytics bool `json:"realtime_analytics"`
    ExportAnalytics   bool `json:"export_analytics"`
    CustomEvents      bool `json:"custom_events"`

    // Customization
    CustomDomains     bool `json:"custom_domains"`
    QRCodes          bool `json:"qr_codes"`
    DeepLinking      bool `json:"deep_linking"`
    UTMBuilder       bool `json:"utm_builder"`
    LinkRotation     bool `json:"link_rotation"`

    // Team Features
    Workspaces       bool `json:"workspaces"`
    TeamMembers      bool `json:"team_members"`
    RoleBasedAccess  bool `json:"role_based_access"`
    AuditLog         bool `json:"audit_log"`

    // Integrations
    APIAccess        bool `json:"api_access"`
    Webhooks         bool `json:"webhooks"`
    Zapier           bool `json:"zapier"`
    SlackIntegration bool `json:"slack_integration"`
    SSOIntegration   bool `json:"sso_integration"`

    // Limits
    Limits struct {
        LinksPerMonth    int `json:"links_per_month"`    // -1 for unlimited
        ClicksPerMonth   int `json:"clicks_per_month"`
        CustomDomains    int `json:"custom_domains"`
        TeamMembers      int `json:"team_members"`
        Workspaces       int `json:"workspaces"`
        APIRequestsPerMin int `json:"api_requests_per_min"`
    } `json:"limits"`
}

// Preset configurations
var (
    StarterFeatures = FeatureFlags{
        URLShortening:    true,
        CustomShortCodes: false,
        BasicAnalytics:   true,
        Limits: struct {
            LinksPerMonth:     1000,
            ClicksPerMonth:    50000,
            CustomDomains:     0,
            TeamMembers:       1,
            Workspaces:        1,
            APIRequestsPerMin: 60,
        },
    }

    BusinessFeatures = FeatureFlags{
        URLShortening:     true,
        CustomShortCodes:  true,
        BulkOperations:    true,
        BasicAnalytics:    true,
        AdvancedAnalytics: true,
        CustomDomains:     true,
        QRCodes:          true,
        Workspaces:       true,
        TeamMembers:      true,
        APIAccess:        true,
        Webhooks:         true,
        Limits: struct {
            LinksPerMonth:     -1, // Unlimited
            ClicksPerMonth:    -1,
            CustomDomains:     5,
            TeamMembers:       25,
            Workspaces:        10,
            APIRequestsPerMin: 1000,
        },
    }

    EnterpriseFeatures = FeatureFlags{
        // All features enabled
        URLShortening:     true,
        CustomShortCodes:  true,
        BulkOperations:    true,
        LinkExpiration:    true,
        PasswordProtected: true,
        BasicAnalytics:    true,
        AdvancedAnalytics: true,
        RealTimeAnalytics: true,
        ExportAnalytics:   true,
        CustomEvents:      true,
        CustomDomains:     true,
        QRCodes:          true,
        DeepLinking:      true,
        UTMBuilder:       true,
        LinkRotation:     true,
        Workspaces:       true,
        TeamMembers:      true,
        RoleBasedAccess:  true,
        AuditLog:         true,
        APIAccess:        true,
        Webhooks:         true,
        Zapier:           true,
        SlackIntegration: true,
        SSOIntegration:   true,
        Limits: struct {
            LinksPerMonth:     -1,
            ClicksPerMonth:    -1,
            CustomDomains:     -1,
            TeamMembers:       -1,
            Workspaces:        -1,
            APIRequestsPerMin: 10000,
        },
    }
)
```

### Custom Integrations

```go
// internal/integrations/webhook.go
package integrations

type WebhookConfig struct {
    Enabled  bool              `json:"enabled"`
    URL      string            `json:"url"`
    Secret   string            `json:"secret"`
    Events   []string          `json:"events"`
    Headers  map[string]string `json:"headers"`
    RetryPolicy RetryPolicy    `json:"retry_policy"`
}

type RetryPolicy struct {
    MaxRetries     int           `json:"max_retries"`
    InitialBackoff time.Duration `json:"initial_backoff"`
    MaxBackoff     time.Duration `json:"max_backoff"`
}

// Webhook events
const (
    EventLinkCreated     = "link.created"
    EventLinkUpdated     = "link.updated"
    EventLinkDeleted     = "link.deleted"
    EventLinkClicked     = "link.clicked"
    EventLinkExpired     = "link.expired"
    EventDomainVerified  = "domain.verified"
    EventUserCreated     = "user.created"
    EventTeamMemberAdded = "team.member_added"
)

// Webhook payload
type WebhookPayload struct {
    ID        string                 `json:"id"`
    Event     string                 `json:"event"`
    Timestamp time.Time              `json:"timestamp"`
    TenantID  string                 `json:"tenant_id"`
    Data      map[string]interface{} `json:"data"`
}
```

---

## Configuration

### Environment Variables

```bash
# .env.production

# === Tenant Configuration ===
TENANT_MODE=multi                    # single, multi
TENANT_ID=acme-corp                  # For single-tenant mode
TENANT_CONFIG_URL=https://config.linkrift.io/tenants

# === Database ===
DATABASE_URL=postgres://user:pass@host:5432/linkrift
DATABASE_POOL_SIZE=20
DATABASE_SSL_MODE=require

# === Redis ===
REDIS_URL=redis://host:6379/0
REDIS_CLUSTER_MODE=true

# === Analytics ===
CLICKHOUSE_URL=clickhouse://host:9000/analytics
ANALYTICS_RETENTION_DAYS=365

# === Security ===
JWT_SECRET=your-256-bit-secret
ENCRYPTION_KEY=your-encryption-key
CORS_ORIGINS=https://admin.acme.com,https://acme.link

# === Features ===
FEATURE_FLAGS_SOURCE=database        # database, file, remote
FEATURE_FLAGS_URL=https://flags.linkrift.io

# === Integrations ===
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your-sendgrid-key

SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-secret

# === Observability ===
LOG_LEVEL=info
LOG_FORMAT=json
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel.linkrift.io
SENTRY_DSN=https://xxx@sentry.io/xxx

# === Rate Limiting ===
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=1000
```

### Tenant Configuration

```json
// config/tenants/acme-corp.json
{
  "tenant_id": "acme-corp",
  "name": "Acme Corporation",
  "status": "active",
  "created_at": "2024-01-15T00:00:00Z",

  "plan": {
    "type": "enterprise",
    "features": "enterprise",
    "custom_overrides": {
      "limits": {
        "api_requests_per_min": 50000
      }
    }
  },

  "branding": {
    "config_url": "https://assets.acme.com/linkrift/branding.json",
    "css_override_url": "https://assets.acme.com/linkrift/custom.css"
  },

  "domains": {
    "short_domains": ["acme.link", "go.acme.com"],
    "dashboard_domain": "links.acme.com",
    "api_domain": "api.links.acme.com"
  },

  "auth": {
    "sso": {
      "enabled": true,
      "provider": "okta",
      "config": {
        "issuer": "https://acme.okta.com",
        "client_id": "xxx",
        "scopes": ["openid", "profile", "email"]
      }
    },
    "allowed_email_domains": ["acme.com", "acme.io"]
  },

  "notifications": {
    "email": {
      "from_name": "Acme Links",
      "from_email": "links@acme.com",
      "reply_to": "support@acme.com"
    },
    "webhooks": [
      {
        "url": "https://hooks.acme.com/linkrift",
        "events": ["link.created", "link.clicked"],
        "secret": "webhook-secret"
      }
    ]
  },

  "data": {
    "retention_days": 730,
    "export_enabled": true,
    "gdpr_compliant": true
  },

  "contacts": {
    "admin": "admin@acme.com",
    "billing": "billing@acme.com",
    "technical": "devops@acme.com"
  }
}
```

### Theme Configuration

```css
/* Custom CSS override for white-label tenant */
/* config/themes/acme-corp/custom.css */

:root {
  /* Override primary colors */
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-500: #3b82f6;
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;

  /* Custom accent */
  --color-accent: #f59e0b;

  /* Typography */
  --font-family-sans: 'Inter', system-ui, sans-serif;
  --font-family-display: 'Poppins', var(--font-family-sans);

  /* Spacing */
  --border-radius-base: 8px;
  --border-radius-lg: 12px;

  /* Shadows */
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* Custom component styles */
.btn-primary {
  background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
  border-radius: var(--border-radius-base);
  font-weight: 600;
  text-transform: none;
  letter-spacing: 0;
}

.btn-primary:hover {
  background: linear-gradient(135deg, var(--color-primary-600), var(--color-primary-700));
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

/* Card customization */
.card {
  border-radius: var(--border-radius-lg);
  border: 1px solid var(--color-gray-200);
  box-shadow: var(--shadow-sm);
}

/* Header branding */
.app-header {
  background: white;
  border-bottom: 1px solid var(--color-gray-100);
}

.app-header .logo {
  height: 32px;
}

/* Footer customization */
.app-footer {
  background: var(--color-gray-50);
  border-top: 1px solid var(--color-gray-200);
}

/* Hide Linkrift branding */
.powered-by-linkrift {
  display: none;
}
```

---

## Enterprise Deployment

### Self-Hosted Deployment

```yaml
# kubernetes/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: linkrift-production

resources:
  - namespace.yaml
  - secrets.yaml
  - configmap.yaml
  - deployment-api.yaml
  - deployment-web.yaml
  - deployment-redirect.yaml
  - deployment-analytics.yaml
  - service.yaml
  - ingress.yaml
  - hpa.yaml
  - pdb.yaml

configMapGenerator:
  - name: linkrift-config
    files:
      - config/app.yaml
      - config/tenants.yaml

secretGenerator:
  - name: linkrift-secrets
    envs:
      - secrets.env

images:
  - name: linkrift/api
    newTag: v2.5.0
  - name: linkrift/web
    newTag: v2.5.0
  - name: linkrift/redirect
    newTag: v2.5.0
```

```yaml
# kubernetes/production/deployment-api.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkrift-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: linkrift-api
  template:
    metadata:
      labels:
        app: linkrift-api
    spec:
      containers:
        - name: api
          image: linkrift/api:v2.5.0
          ports:
            - containerPort: 8080
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: linkrift-secrets
                  key: DATABASE_URL
            - name: REDIS_URL
              valueFrom:
                secretKeyRef:
                  name: linkrift-secrets
                  key: REDIS_URL
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              cpu: 2000m
              memory: 2Gi
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

### Dedicated Cloud Instance

For large enterprises requiring complete isolation:

```hcl
# terraform/dedicated/main.tf

module "dedicated_instance" {
  source = "./modules/linkrift-dedicated"

  tenant_id   = "acme-corp"
  environment = "production"
  region      = "us-east-1"

  # Compute
  api_instance_type    = "c6i.2xlarge"
  api_min_instances    = 3
  api_max_instances    = 20

  redirect_instance_type = "c6i.xlarge"
  redirect_min_instances = 5
  redirect_max_instances = 50

  # Database
  db_instance_class    = "db.r6g.2xlarge"
  db_multi_az         = true
  db_storage_size     = 500
  db_backup_retention = 30

  # Cache
  redis_node_type     = "cache.r6g.xlarge"
  redis_num_nodes     = 3

  # Analytics
  clickhouse_instance_type = "r6i.2xlarge"
  clickhouse_storage_size  = 2000

  # Networking
  vpc_cidr = "10.100.0.0/16"
  enable_private_link = true

  # Security
  enable_waf           = true
  enable_shield        = true
  encryption_key_arn   = var.kms_key_arn

  # Monitoring
  enable_enhanced_monitoring = true
  alarm_email               = "ops@acme.com"

  tags = {
    Tenant      = "acme-corp"
    Environment = "production"
    CostCenter  = "engineering"
  }
}
```

### Hybrid Deployment

```yaml
# Hybrid deployment: On-prem redirect, cloud management
# config/hybrid-deployment.yaml

deployment:
  mode: hybrid

  # On-premises components (for latency-sensitive redirect)
  on_premises:
    location: acme-datacenter-1
    components:
      - redirect-service
      - redis-cache
    network:
      vpc_peering:
        enabled: true
        peer_vpc_id: vpc-0123456789

  # Cloud components (management and analytics)
  cloud:
    provider: aws
    region: us-east-1
    components:
      - api-service
      - web-dashboard
      - analytics-service
      - database

  # Data synchronization
  sync:
    links:
      direction: cloud_to_onprem
      frequency: realtime
      method: redis_replication
    analytics:
      direction: onprem_to_cloud
      frequency: 1m
      method: kafka
```

---

## API Customization

White-label API endpoints:

```go
// internal/api/whitelabel/router.go
package whitelabel

import (
    "github.com/go-chi/chi/v5"
)

func NewRouter(cfg *Config) chi.Router {
    r := chi.NewRouter()

    // Tenant-specific middleware
    r.Use(TenantContextMiddleware(cfg))
    r.Use(BrandingMiddleware(cfg))
    r.Use(FeatureFlagsMiddleware(cfg))

    // Public API (tenant's custom domain)
    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/links", CreateLink)
        r.Get("/links/{code}", GetLink)
        r.Delete("/links/{code}", DeleteLink)
        r.Get("/links/{code}/analytics", GetAnalytics)
    })

    // Admin API
    r.Route("/admin/api/v1", func(r chi.Router) {
        r.Use(AdminAuthMiddleware)
        r.Get("/settings", GetTenantSettings)
        r.Put("/settings", UpdateTenantSettings)
        r.Get("/domains", ListDomains)
        r.Post("/domains", AddDomain)
        r.Get("/users", ListUsers)
        r.Post("/users", InviteUser)
    })

    return r
}

// Custom API response wrapper
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *APIMeta    `json:"meta,omitempty"`
}

type APIMeta struct {
    RequestID string `json:"request_id"`
    Timestamp string `json:"timestamp"`
    // Tenant can customize additional meta fields
    Custom    map[string]interface{} `json:"custom,omitempty"`
}
```

---

## Security & Compliance

```yaml
# config/security.yaml
security:
  # Data encryption
  encryption:
    at_rest:
      enabled: true
      algorithm: AES-256-GCM
      key_management: aws-kms
    in_transit:
      tls_version: "1.3"
      cipher_suites:
        - TLS_AES_256_GCM_SHA384
        - TLS_CHACHA20_POLY1305_SHA256

  # Authentication
  authentication:
    session_timeout: 3600
    mfa_required: true
    password_policy:
      min_length: 12
      require_uppercase: true
      require_lowercase: true
      require_numbers: true
      require_special: true
      max_age_days: 90

  # Authorization
  authorization:
    rbac_enabled: true
    default_role: viewer
    roles:
      - name: admin
        permissions: ["*"]
      - name: editor
        permissions: ["links:*", "analytics:read"]
      - name: viewer
        permissions: ["links:read", "analytics:read"]

  # Audit logging
  audit:
    enabled: true
    retention_days: 365
    events:
      - user.login
      - user.logout
      - link.create
      - link.delete
      - settings.update
      - domain.add
      - domain.remove

  # Compliance
  compliance:
    gdpr:
      enabled: true
      data_retention_days: 730
      right_to_erasure: true
      data_portability: true
    soc2:
      enabled: true
    hipaa:
      enabled: false  # Enable for healthcare tenants
```

---

## Monitoring & Analytics

```yaml
# config/monitoring.yaml
monitoring:
  # Metrics
  metrics:
    provider: prometheus
    endpoints:
      - path: /metrics
        port: 9090

  # Logging
  logging:
    provider: elasticsearch
    format: json
    level: info
    retention_days: 30

  # Tracing
  tracing:
    provider: jaeger
    sample_rate: 0.1
    endpoint: https://tracing.linkrift.io

  # Alerting
  alerting:
    provider: pagerduty
    escalation_policy: default
    rules:
      - name: high_error_rate
        condition: error_rate > 0.01
        severity: critical
      - name: high_latency
        condition: p99_latency > 500ms
        severity: warning
      - name: redirect_failures
        condition: redirect_success_rate < 0.999
        severity: critical

  # Dashboards
  dashboards:
    - name: tenant-overview
      provider: grafana
      template: tenant-overview-v2
    - name: redirect-performance
      provider: grafana
      template: redirect-performance-v2
```

---

## Support & SLA

| Tier | Response Time | Availability | Support Channels |
|------|---------------|--------------|------------------|
| Standard | 24 hours | 99.9% | Email, Docs |
| Business | 4 hours | 99.95% | Email, Chat, Phone |
| Enterprise | 1 hour | 99.99% | Dedicated TAM, 24/7 Phone |

**Enterprise SLA guarantees:**
- 99.99% redirect uptime
- < 50ms P99 redirect latency
- 24/7 on-call support
- Dedicated technical account manager
- Quarterly business reviews

---

For white-label inquiries, contact enterprise@linkrift.io or visit [linkrift.io/enterprise](https://linkrift.io/enterprise).
