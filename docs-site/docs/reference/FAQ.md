# Frequently Asked Questions

> Last Updated: 2025-01-24

Answers to commonly asked questions about Linkrift URL shortener.

---

## Table of Contents

- [General Questions](#general-questions)
- [Technical Questions](#technical-questions)
- [Pricing Questions](#pricing-questions)
- [Feature Questions](#feature-questions)
- [Deployment Questions](#deployment-questions)
- [Contributing Questions](#contributing-questions)

---

## General Questions

### What is Linkrift?

Linkrift is a modern, high-performance URL shortening service built with Go (backend) and React with Vite (frontend). It provides enterprise-grade link management with features like custom domains, analytics, QR codes, and deep linking.

### Why choose Linkrift over Bitly or TinyURL?

| Feature | Linkrift | Bitly | TinyURL |
|---------|----------|-------|---------|
| Self-hosting | Yes | No | No |
| Open source | Yes | No | No |
| Custom domains | Yes | Paid | Limited |
| API access | Unlimited | Limited | Limited |
| Real-time analytics | Yes | Yes | Limited |
| White-label | Yes | Enterprise only | No |
| Pricing | Free/Self-hosted | Expensive | Free/Limited |

### Is Linkrift free?

Yes, Linkrift is open source and free to self-host. We also offer a managed cloud version with both free and paid tiers.

**Self-hosted:** Free forever, unlimited links
**Cloud Free:** 1,000 links/month, basic analytics
**Cloud Pro:** $15/month, unlimited links, advanced features
**Enterprise:** Custom pricing, white-label, dedicated support

### What are the system requirements for self-hosting?

**Minimum requirements:**
- 2 CPU cores
- 4 GB RAM
- 20 GB storage
- PostgreSQL 14+
- Redis 7+

**Recommended for production:**
- 4+ CPU cores
- 8+ GB RAM
- SSD storage
- Load balancer
- CDN for static assets

### How reliable is Linkrift?

The managed cloud service maintains 99.99% uptime SLA. The redirect service is designed for sub-50ms latency at the 99th percentile. Self-hosted reliability depends on your infrastructure.

---

## Technical Questions

### What technology stack does Linkrift use?

**Backend:**
- Go 1.22+ with standard library HTTP server
- PostgreSQL for primary data storage
- Redis for caching and rate limiting
- ClickHouse for analytics (optional)

**Frontend:**
- React 18 with TypeScript
- Vite for build tooling
- TanStack Query for data fetching
- Tailwind CSS for styling

### How does the short code generation work?

Linkrift generates cryptographically random short codes using Go's `crypto/rand` package. By default, codes are 6 characters using base62 encoding (a-z, A-Z, 0-9), providing over 56 billion possible combinations.

```go
// Simplified example
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateCode(length int) string {
    code := make([]byte, length)
    for i := range code {
        code[i] = charset[secureRandomInt(len(charset))]
    }
    return string(code)
}
```

### How fast are redirects?

Linkrift achieves extremely fast redirects through:

1. **Redis caching**: Hot links are served from memory
2. **Go's efficiency**: Minimal overhead in the redirect path
3. **Optimized routing**: Direct lookup without ORM overhead
4. **Connection pooling**: Reused database connections

**Typical latency:**
- Cached links: 5-15ms
- Database lookup: 20-50ms
- With geolocation: +10-20ms

### Can I use my own database?

Yes, Linkrift supports any PostgreSQL-compatible database:

- PostgreSQL 14+
- Amazon RDS PostgreSQL
- Google Cloud SQL
- Azure Database for PostgreSQL
- CockroachDB
- TimescaleDB

### How does analytics tracking work?

When a link is clicked:

1. Redirect service logs the click asynchronously
2. Click data is buffered in Redis
3. Background worker processes clicks in batches
4. Data is aggregated and stored in ClickHouse/PostgreSQL
5. No third-party trackers are used

**Data collected:**
- Timestamp
- Referrer
- User agent (parsed for device/browser/OS)
- Country/region (from IP, not stored)
- UTM parameters

### Is there an API rate limit?

**Cloud service:**
- Free: 60 requests/minute
- Pro: 1,000 requests/minute
- Enterprise: Custom limits

**Self-hosted:**
- Configurable via environment variables
- Default: 1,000 requests/minute per API key

### How do I handle high traffic?

For high-traffic deployments:

1. **Horizontal scaling**: Run multiple API instances behind a load balancer
2. **Redis cluster**: Use Redis Cluster for distributed caching
3. **Read replicas**: Add PostgreSQL read replicas for analytics queries
4. **CDN**: Use a CDN for the redirect service edge locations
5. **Async processing**: Click tracking is already asynchronous

See the [Performance Optimization](./PERFORMANCE_OPTIMIZATION.md) guide for details.

### Does Linkrift support webhooks?

Yes, Linkrift can send webhooks for various events:

```json
{
  "event": "link.clicked",
  "timestamp": "2025-01-24T10:30:00Z",
  "data": {
    "short_code": "abc123",
    "clicks": 100,
    "country": "US"
  }
}
```

Supported events:
- `link.created`
- `link.updated`
- `link.deleted`
- `link.clicked`
- `link.expired`

---

## Pricing Questions

### What's included in the free tier?

**Cloud Free tier includes:**
- 1,000 shortened links per month
- Basic click analytics (7-day retention)
- 1 custom domain
- API access (60 req/min)
- Community support

### How does billing work for the Pro plan?

The Pro plan ($15/month) is billed monthly. You can cancel anytime. Annual billing available at $144/year (20% discount).

**Pro tier includes:**
- Unlimited links
- Advanced analytics (1-year retention)
- 5 custom domains
- API access (1,000 req/min)
- QR code generation
- Link expiration
- Password protection
- Priority email support

### Is there a free trial for Enterprise?

Yes, we offer a 14-day free trial of the Enterprise tier with all features enabled. Contact sales@linkrift.io to get started.

### Can I get a refund?

Yes, we offer a 30-day money-back guarantee for paid plans. Contact support@linkrift.io for refund requests.

### Are there any hidden costs for self-hosting?

No hidden costs. Self-hosting is completely free. Your only costs are:
- Server hosting (VPS, cloud, on-premises)
- Domain registration (if using custom domains)
- SSL certificates (free with Let's Encrypt)

---

## Feature Questions

### Can I use my own domain for short links?

Yes! Custom domains are supported on all plans (1 domain on Free, 5 on Pro, unlimited on Enterprise).

Setup process:
1. Add your domain in the dashboard
2. Configure DNS (CNAME to linkrift.io or your self-hosted instance)
3. SSL is automatically provisioned via Let's Encrypt

### Do short links expire?

By default, links never expire. However, you can set expiration dates on Pro and Enterprise plans:

- Expire after a specific date
- Expire after X clicks
- Expire after X days from creation

### Can I edit a link after creation?

Yes, you can update:
- Destination URL
- Custom short code (if available)
- Expiration settings
- Password protection
- Metadata

Note: Changing the destination URL takes effect immediately for all future clicks.

### Does Linkrift support deep linking for mobile apps?

Yes, Linkrift supports:
- Universal Links (iOS)
- App Links (Android)
- Deferred deep linking
- Custom URL schemes

See the [Mobile SDK documentation](../integrations/MOBILE_SDK.md) for details.

### Can I generate QR codes for my links?

Yes, QR codes are available on Pro and Enterprise plans. Features:
- Customizable colors
- Logo embedding
- Multiple formats (PNG, SVG, PDF)
- Download and embed options

### Is there bulk link creation?

Yes, bulk operations are available:

**Via API:**
```bash
curl -X POST https://api.linkrift.io/v1/links/bulk \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"urls": ["https://example.com/1", "https://example.com/2"]}'
```

**Via Dashboard:**
Upload a CSV file with URLs to create multiple links at once.

### How do I track UTM parameters?

Linkrift has a built-in UTM builder:

1. Create a link
2. Click "Add UTM Parameters"
3. Fill in source, medium, campaign, etc.
4. Parameters are automatically appended to the destination URL

Analytics dashboard shows traffic breakdown by UTM parameters.

---

## Deployment Questions

### How do I deploy Linkrift with Docker?

```bash
# Quick start with Docker Compose
git clone https://github.com/linkrift/linkrift.git
cd linkrift
cp .env.example .env
# Edit .env with your settings
docker-compose up -d
```

See the [Deployment Guide](../operations/DEPLOYMENT.md) for detailed instructions.

### Can I run Linkrift on Kubernetes?

Yes, we provide Helm charts and Kubernetes manifests:

```bash
helm repo add linkrift https://charts.linkrift.io
helm install linkrift linkrift/linkrift \
  --set postgresql.enabled=true \
  --set redis.enabled=true
```

### How do I configure SSL/TLS?

**With Let's Encrypt (recommended):**
```yaml
# docker-compose.yml
services:
  traefik:
    image: traefik:v2.10
    command:
      - "--certificatesresolvers.le.acme.email=you@example.com"
      - "--certificatesresolvers.le.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.le.acme.httpchallenge.entrypoint=web"
```

**With custom certificates:**
```yaml
environment:
  - TLS_CERT_FILE=/certs/cert.pem
  - TLS_KEY_FILE=/certs/key.pem
```

### How do I backup my data?

**PostgreSQL backup:**
```bash
pg_dump -h localhost -U linkrift linkrift > backup.sql
```

**Automated backups (recommended):**
```bash
# Add to crontab
0 2 * * * pg_dump -h localhost -U linkrift linkrift | gzip > /backups/linkrift-$(date +\%Y\%m\%d).sql.gz
```

### How do I migrate from another service?

See the [Migration Guide](./MIGRATION_GUIDE.md) for step-by-step instructions on migrating from:
- Bitly
- TinyURL
- Rebrandly
- Other services

### Can I run Linkrift behind a reverse proxy?

Yes, Linkrift works with any reverse proxy:

**Nginx:**
```nginx
server {
    listen 80;
    server_name lnk.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Caddy:**
```
lnk.example.com {
    reverse_proxy localhost:8080
}
```

---

## Contributing Questions

### How can I contribute to Linkrift?

We welcome contributions! Here's how to get started:

1. Fork the repository on GitHub
2. Create a feature branch
3. Make your changes
4. Submit a pull request

See [CONTRIBUTING.md](../contributing/CONTRIBUTING.md) for detailed guidelines.

### What types of contributions are welcome?

- Bug fixes
- New features
- Documentation improvements
- Translations
- Test coverage
- Performance improvements
- Security fixes

### How do I report a bug?

Open an issue on GitHub with:
- Description of the bug
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, browser, version)
- Screenshots if applicable

### How do I request a feature?

Open a GitHub issue with:
- Clear description of the feature
- Use case / problem it solves
- Proposed implementation (optional)
- Mockups or examples (optional)

### Is there a roadmap I can check?

Yes! See the [Roadmap](./ROADMAP.md) for planned features and milestones.

### How do I get help with development?

- **GitHub Discussions**: Ask questions, share ideas
- **Discord**: Real-time chat with contributors
- **Documentation**: Comprehensive guides and API reference

---

## Still Have Questions?

If your question isn't answered here:

1. **Search existing issues** on GitHub
2. **Ask in Discussions** for general questions
3. **Open an issue** for bugs or feature requests
4. **Contact support** at support@linkrift.io for account-related questions

We're here to help!
