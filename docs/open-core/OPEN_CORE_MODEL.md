# Linkrift Open Core Business Model

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [Why AGPL-3.0 License](#why-agpl-30-license)
  - [Benefits of AGPL-3.0](#benefits-of-agpl-30)
  - [What This Means for Users](#what-this-means-for-users)
- [Community Edition vs Enterprise Edition](#community-edition-vs-enterprise-edition)
  - [Community Edition (CE)](#community-edition-ce)
  - [Enterprise Edition (EE)](#enterprise-edition-ee)
- [Feature Comparison by Tier](#feature-comparison-by-tier)
  - [Detailed Feature Table](#detailed-feature-table)
  - [Tier Descriptions](#tier-descriptions)
- [Self-Hosting Options](#self-hosting-options)
  - [Docker Deployment](#docker-deployment)
  - [Kubernetes Deployment](#kubernetes-deployment)
  - [Bare Metal Installation](#bare-metal-installation)
- [Revenue Model](#revenue-model)
  - [Revenue Streams](#revenue-streams)
  - [Pricing Philosophy](#pricing-philosophy)

---

## Overview

Linkrift follows an **open core business model**, where the core URL shortener functionality is freely available under an open-source license, while advanced enterprise features are offered through commercial licenses. This approach ensures:

1. **Community accessibility** - Anyone can use, modify, and self-host Linkrift
2. **Sustainable development** - Commercial offerings fund continued development
3. **Enterprise confidence** - Organizations get the features, support, and guarantees they need

---

## Why AGPL-3.0 License

Linkrift Community Edition is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. This license was chosen deliberately to balance openness with sustainability.

### Benefits of AGPL-3.0

| Benefit | Description |
|---------|-------------|
| **Network Copyleft** | If you modify Linkrift and provide it as a service, you must share your modifications |
| **Prevents Proprietary Forks** | Companies cannot take our code, modify it, and offer it as a closed-source SaaS |
| **Encourages Contribution** | Modifications flow back to the community |
| **True Open Source** | All freedoms of the GPL, extended to network use |

### What This Means for Users

**You CAN:**
- Use Linkrift for any purpose (personal, commercial, educational)
- Modify the source code for your own use
- Self-host Linkrift within your organization
- Contribute improvements back to the project

**You MUST (if distributing or offering as a service):**
- Keep the AGPL-3.0 license intact
- Make your modifications available under AGPL-3.0
- Provide access to the complete source code

**You CANNOT:**
- Remove or hide license notices
- Offer Linkrift as a closed-source service without a commercial license
- Sublicense under different terms

```text
Note: If AGPL-3.0 requirements don't work for your use case,
commercial licenses are available that remove these obligations.
```

---

## Community Edition vs Enterprise Edition

### Community Edition (CE)

The Community Edition includes all core URL shortening functionality:

```yaml
# Community Edition Features
core_features:
  - URL shortening with custom slugs
  - Basic link analytics (clicks, referrers)
  - REST API access
  - Single user or team (up to 5 members)
  - PostgreSQL/SQLite database support
  - Docker deployment support
  - Basic rate limiting
  - Webhook notifications
  - Link expiration
  - Password-protected links

limitations:
  - No SSO/SAML integration
  - No advanced analytics
  - No custom domains (limited to 1)
  - No audit logging
  - Community support only
```

### Enterprise Edition (EE)

The Enterprise Edition extends CE with advanced features:

```yaml
# Enterprise Edition Features (in addition to CE)
enterprise_features:
  - SAML/SSO authentication (Okta, Azure AD, Google Workspace)
  - Advanced analytics and reporting
  - Unlimited custom domains
  - Multi-region deployment support
  - Audit logging with compliance exports
  - Role-based access control (RBAC)
  - API rate limit customization
  - Priority support with SLA
  - Custom integrations
  - White-labeling options
  - HA clustering support
  - Dedicated success manager (Enterprise tier)
```

---

## Feature Comparison by Tier

### Detailed Feature Table

| Feature | Free | Pro | Business | Enterprise |
|---------|:----:|:---:|:--------:|:----------:|
| **Core Features** |
| URL Shortening | Unlimited | Unlimited | Unlimited | Unlimited |
| Custom Slugs | Yes | Yes | Yes | Yes |
| Link Expiration | Yes | Yes | Yes | Yes |
| Password Protection | Yes | Yes | Yes | Yes |
| QR Code Generation | Yes | Yes | Yes | Yes |
| REST API | Yes | Yes | Yes | Yes |
| **Analytics** |
| Click Tracking | 30 days | 1 year | 2 years | Unlimited |
| Geographic Data | Country only | City-level | City-level | City-level |
| Device Analytics | Basic | Detailed | Detailed | Detailed |
| Custom Reports | - | - | Yes | Yes |
| Export Data | CSV | CSV, JSON | All formats | All formats |
| Real-time Dashboard | - | Yes | Yes | Yes |
| **Team & Access** |
| Team Members | 1 | 5 | 25 | Unlimited |
| Workspaces | 1 | 3 | 10 | Unlimited |
| Role-Based Access | - | Basic | Advanced | Full RBAC |
| SSO/SAML | - | - | Yes | Yes |
| SCIM Provisioning | - | - | - | Yes |
| **Domains & Branding** |
| Custom Domains | 1 | 3 | 10 | Unlimited |
| SSL Certificates | Auto | Auto | Auto | Auto + Custom |
| White Labeling | - | - | Partial | Full |
| Custom Branding | - | Logo | Full | Full |
| **Infrastructure** |
| API Rate Limit | 100/min | 1,000/min | 10,000/min | Custom |
| Webhooks | 2 | 10 | 50 | Unlimited |
| Data Retention | 30 days | 1 year | 2 years | Custom |
| Uptime SLA | - | 99.5% | 99.9% | 99.99% |
| Multi-region | - | - | Yes | Yes |
| **Compliance & Security** |
| Audit Logs | - | - | 90 days | Unlimited |
| SOC 2 Type II | - | - | Yes | Yes |
| GDPR Tools | Basic | Full | Full | Full |
| HIPAA BAA | - | - | - | Yes |
| **Support** |
| Documentation | Yes | Yes | Yes | Yes |
| Community Forum | Yes | Yes | Yes | Yes |
| Email Support | - | Yes | Priority | Priority |
| Phone Support | - | - | - | Yes |
| Dedicated CSM | - | - | - | Yes |
| **Pricing** |
| Monthly (Cloud) | $0 | $29 | $99 | Custom |
| Annual (Cloud) | $0 | $290 | $990 | Custom |
| Self-Hosted | Free | $199/yr | $999/yr | Custom |

### Tier Descriptions

#### Free Tier
Perfect for individuals and small projects. Includes core URL shortening with basic analytics.

#### Pro Tier
Designed for growing teams and professional use. Adds extended analytics, more team members, and email support.

#### Business Tier
Built for organizations requiring advanced features. Includes SSO, advanced analytics, audit logs, and compliance features.

#### Enterprise Tier
For large organizations with complex requirements. Offers unlimited everything, full compliance, and dedicated support.

---

## Self-Hosting Options

Linkrift supports multiple self-hosting deployment methods:

### Docker Deployment

The simplest way to get started:

```bash
# Quick start with Docker Compose
git clone https://github.com/link-rift/link-rift.git
cd link-rift

# Configure environment
cp .env.example .env
vim .env

# Start services
docker-compose up -d
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  linkrift:
    image: ghcr.io/link-rift/link-rift:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://linkrift:password@db:5432/linkrift
      - REDIS_URL=redis://redis:6379
      - LINKRIFT_LICENSE_KEY=${LINKRIFT_LICENSE_KEY}
    depends_on:
      - db
      - redis

  db:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=linkrift
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=linkrift

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment

For production-grade deployments:

```bash
# Add Helm repository
helm repo add linkrift https://charts.linkrift.io
helm repo update

# Install with Helm
helm install linkrift linkrift/linkrift \
  --namespace linkrift \
  --create-namespace \
  --set license.key="${LINKRIFT_LICENSE_KEY}" \
  --set postgresql.enabled=true \
  --set redis.enabled=true
```

```yaml
# values.yaml - Example configuration
replicaCount: 3

image:
  repository: ghcr.io/link-rift/link-rift
  tag: latest

license:
  key: ""  # Set via --set or secret

postgresql:
  enabled: true
  auth:
    database: linkrift

redis:
  enabled: true
  architecture: standalone

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: links.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: linkrift-tls
      hosts:
        - links.example.com

resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Bare Metal Installation

For environments without containerization:

```bash
# Download latest release
curl -LO https://github.com/link-rift/link-rift/releases/latest/download/linkrift-linux-amd64.tar.gz
tar -xzf linkrift-linux-amd64.tar.gz
sudo mv linkrift /usr/local/bin/

# Create configuration directory
sudo mkdir -p /etc/linkrift
sudo cp config.example.yaml /etc/linkrift/config.yaml

# Create systemd service
sudo tee /etc/systemd/system/linkrift.service << EOF
[Unit]
Description=Linkrift URL Shortener
After=network.target postgresql.service

[Service]
Type=simple
User=linkrift
ExecStart=/usr/local/bin/linkrift serve --config /etc/linkrift/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl enable linkrift
sudo systemctl start linkrift
```

---

## Revenue Model

### Revenue Streams

Linkrift generates revenue through multiple channels:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Revenue Streams                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Cloud SaaS    │  │  Self-Hosted    │  │   Professional  │  │
│  │   Subscriptions │  │    Licenses     │  │    Services     │  │
│  │                 │  │                 │  │                 │  │
│  │  - Pro tier     │  │  - Annual keys  │  │  - Consulting   │  │
│  │  - Business     │  │  - Support      │  │  - Integration  │  │
│  │  - Enterprise   │  │  - Updates      │  │  - Training     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│           │                   │                    │             │
│           ▼                   ▼                    ▼             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                   Sustainable Development                    ││
│  │                                                              ││
│  │  - Core feature development                                  ││
│  │  - Security updates                                          ││
│  │  - Community support                                         ││
│  │  - Documentation                                             ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

| Revenue Stream | Description | Target Customer |
|----------------|-------------|-----------------|
| **Cloud SaaS** | Hosted service with tiered plans | Small to mid-size businesses |
| **Self-Hosted Licenses** | Annual licenses for on-premise | Enterprises, regulated industries |
| **Support Contracts** | Priority support and SLAs | All commercial customers |
| **Professional Services** | Custom development, integration | Enterprise customers |
| **Training & Certification** | Team training programs | Large organizations |

### Pricing Philosophy

Our pricing is guided by these principles:

1. **Free tier is genuinely useful** - Not a demo, but a complete product for small use cases
2. **Transparent pricing** - No hidden fees or surprise charges
3. **Pay for value** - Higher tiers unlock proportionally more value
4. **Self-hosting option** - Organizations can run on their infrastructure
5. **Fair source access** - Code is readable; commercial use requires license

```go
// Tier determination logic
func DetermineTier(usage Usage) Tier {
    switch {
    case usage.MonthlyClicks > 1_000_000:
        return TierEnterprise
    case usage.MonthlyClicks > 100_000:
        return TierBusiness
    case usage.MonthlyClicks > 10_000:
        return TierPro
    default:
        return TierFree
    }
}
```

---

## Related Documentation

- [License System](./LICENSE_SYSTEM.md) - Technical details of license verification
- [Repository Structure](./REPOSITORY_STRUCTURE.md) - Code organization and contribution guide
- [Self-Hosting Guide](../self-hosting/README.md) - Detailed deployment instructions
- [API Documentation](../api/README.md) - REST API reference

---

*Linkrift is committed to building a sustainable open-source business that benefits both the community and our commercial customers.*
