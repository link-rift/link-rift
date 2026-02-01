# Migration Guide

> Last Updated: 2025-01-24

Step-by-step guide for migrating to Linkrift from other URL shortening services including Bitly, TinyURL, Rebrandly, and others.

---

## Table of Contents

- [Overview](#overview)
- [Pre-Migration Checklist](#pre-migration-checklist)
- [Migrating from Bitly](#migrating-from-bitly)
- [Migrating from TinyURL](#migrating-from-tinyurl)
- [Migrating from Rebrandly](#migrating-from-rebrandly)
- [Migrating from Other Services](#migrating-from-other-services)
- [Data Import/Export](#data-importexport)
  - [Import Formats](#import-formats)
  - [Export Options](#export-options)
  - [API Migration](#api-migration)
- [Domain Migration](#domain-migration)
- [Redirect Handling](#redirect-handling)
- [Preserving Analytics](#preserving-analytics)
- [Post-Migration Verification](#post-migration-verification)
- [Rollback Plan](#rollback-plan)

---

## Overview

Migrating to Linkrift involves:

1. **Exporting existing data** from your current service
2. **Importing links** into Linkrift
3. **Configuring custom domains** (if applicable)
4. **Setting up redirects** from old short URLs
5. **Updating integrations** to use Linkrift API
6. **Verifying everything works** correctly

**Migration Support:**
- Self-service migration tools in dashboard
- API for bulk imports
- Professional migration assistance (Enterprise plans)

---

## Pre-Migration Checklist

Before starting your migration:

### 1. Inventory Your Links

- [ ] Total number of links to migrate
- [ ] Links with custom short codes
- [ ] Links with tags/labels
- [ ] Links with custom domains
- [ ] Links with expiration dates

### 2. Document Current Setup

- [ ] List all custom domains
- [ ] Note DNS configuration
- [ ] Record SSL certificate details
- [ ] Document API integrations
- [ ] List team members and permissions

### 3. Plan Downtime (if any)

- [ ] Schedule migration window
- [ ] Notify stakeholders
- [ ] Prepare rollback plan
- [ ] Set up monitoring

### 4. Prepare Linkrift Account

- [ ] Create Linkrift account
- [ ] Upgrade to appropriate plan
- [ ] Add custom domains
- [ ] Invite team members

---

## Migrating from Bitly

### Step 1: Export from Bitly

**Via Bitly Dashboard:**
1. Log in to Bitly
2. Go to Settings > Data Export
3. Select date range (or "All time")
4. Choose export format (CSV recommended)
5. Click "Export" and download the file

**Via Bitly API:**

```bash
# Get all bitlinks for a group
curl -X GET "https://api-ssl.bitly.com/v4/groups/{group_guid}/bitlinks" \
  -H "Authorization: Bearer YOUR_BITLY_TOKEN" \
  > bitly_export.json

# Paginate through all results
#!/bin/bash
PAGE=1
while true; do
  RESPONSE=$(curl -s -X GET \
    "https://api-ssl.bitly.com/v4/groups/{group_guid}/bitlinks?page=$PAGE&size=100" \
    -H "Authorization: Bearer YOUR_BITLY_TOKEN")

  echo "$RESPONSE" >> bitly_all_links.json

  NEXT=$(echo "$RESPONSE" | jq -r '.pagination.next')
  if [ "$NEXT" == "null" ]; then
    break
  fi
  ((PAGE++))
done
```

### Step 2: Transform Data

Convert Bitly format to Linkrift format:

```javascript
// scripts/transform-bitly.js
const fs = require('fs');
const csv = require('csv-parse/sync');

const bitlyData = fs.readFileSync('bitly_export.csv', 'utf-8');
const records = csv.parse(bitlyData, { columns: true });

const linkriftLinks = records.map(record => ({
  original_url: record['Long URL'],
  custom_code: extractCode(record['Bitlink']),
  title: record['Title'] || null,
  created_at: record['Created'],
  tags: record['Tags'] ? record['Tags'].split(',').map(t => t.trim()) : [],
  // Preserve analytics reference
  metadata: {
    bitly_id: record['Bitlink'],
    migrated_from: 'bitly',
    migration_date: new Date().toISOString(),
  },
}));

function extractCode(bitlink) {
  // Extract code from bit.ly/abc123
  const match = bitlink.match(/bit\.ly\/(.+)/);
  return match ? match[1] : null;
}

fs.writeFileSync(
  'linkrift_import.json',
  JSON.stringify(linkriftLinks, null, 2)
);

console.log(`Transformed ${linkriftLinks.length} links`);
```

### Step 3: Import to Linkrift

**Via Dashboard:**
1. Go to Settings > Import
2. Upload `linkrift_import.json`
3. Review import preview
4. Click "Import Links"

**Via API:**

```bash
curl -X POST "https://api.linkrift.io/v1/links/import" \
  -H "Authorization: Bearer YOUR_LINKRIFT_TOKEN" \
  -H "Content-Type: application/json" \
  -d @linkrift_import.json
```

### Step 4: Set Up Redirects

If you're not keeping bit.ly domain, set up redirects:

```nginx
# nginx.conf - Redirect old bit.ly links to Linkrift
server {
    server_name bit.ly;

    location / {
        return 301 https://lnkr.ft$request_uri;
    }
}
```

Or use Bitly's branded short domain to point to Linkrift (if you own it).

---

## Migrating from TinyURL

### Step 1: Export from TinyURL

TinyURL doesn't have a native export feature. Options:

**Option A: Manual Export**
- Log in to TinyURL Pro
- Go to "My TinyURLs"
- Use browser developer tools to extract data
- Or manually copy link data

**Option B: API Export (Pro users)**

```python
# scripts/export_tinyurl.py
import requests
import json

API_KEY = 'your_tinyurl_api_key'
BASE_URL = 'https://api.tinyurl.com'

def get_all_links():
    links = []
    page = 1

    while True:
        response = requests.get(
            f'{BASE_URL}/alias/list',
            headers={'Authorization': f'Bearer {API_KEY}'},
            params={'page': page, 'per_page': 100}
        )
        data = response.json()

        if not data['data']:
            break

        links.extend(data['data'])
        page += 1

    return links

links = get_all_links()
with open('tinyurl_export.json', 'w') as f:
    json.dump(links, f, indent=2)

print(f'Exported {len(links)} links')
```

### Step 2: Transform and Import

```python
# scripts/transform_tinyurl.py
import json

with open('tinyurl_export.json') as f:
    tinyurl_links = json.load(f)

linkrift_links = []
for link in tinyurl_links:
    linkrift_links.append({
        'original_url': link['url'],
        'custom_code': link['alias'],
        'created_at': link['created_at'],
        'metadata': {
            'tinyurl_id': link['id'],
            'migrated_from': 'tinyurl'
        }
    })

with open('linkrift_import.json', 'w') as f:
    json.dump(linkrift_links, f, indent=2)
```

---

## Migrating from Rebrandly

### Step 1: Export from Rebrandly

**Via Dashboard:**
1. Go to Links page
2. Click "Export" button
3. Select format (CSV)
4. Download file

**Via API:**

```python
# scripts/export_rebrandly.py
import requests
import json

API_KEY = 'your_rebrandly_api_key'
WORKSPACE_ID = 'your_workspace_id'

def get_all_links():
    links = []
    last_id = None

    while True:
        params = {'limit': 25}
        if last_id:
            params['last'] = last_id

        response = requests.get(
            'https://api.rebrandly.com/v1/links',
            headers={
                'apikey': API_KEY,
                'workspace': WORKSPACE_ID
            },
            params=params
        )
        data = response.json()

        if not data:
            break

        links.extend(data)
        last_id = data[-1]['id']

    return links

# Also export domains
def get_domains():
    response = requests.get(
        'https://api.rebrandly.com/v1/domains',
        headers={
            'apikey': API_KEY,
            'workspace': WORKSPACE_ID
        }
    )
    return response.json()

links = get_all_links()
domains = get_domains()

with open('rebrandly_export.json', 'w') as f:
    json.dump({'links': links, 'domains': domains}, f, indent=2)
```

### Step 2: Transform Data

```python
# scripts/transform_rebrandly.py
import json

with open('rebrandly_export.json') as f:
    data = json.load(f)

linkrift_links = []
for link in data['links']:
    linkrift_links.append({
        'original_url': link['destination'],
        'custom_code': link['slashtag'],
        'domain': link['domain']['fullName'],
        'title': link.get('title'),
        'created_at': link['createdAt'],
        'tags': [tag['name'] for tag in link.get('tags', [])],
        'metadata': {
            'rebrandly_id': link['id'],
            'migrated_from': 'rebrandly'
        }
    })

with open('linkrift_import.json', 'w') as f:
    json.dump(linkrift_links, f, indent=2)

# Document domains to set up
print("Domains to configure in Linkrift:")
for domain in data['domains']:
    print(f"  - {domain['fullName']}")
```

---

## Migrating from Other Services

### Generic Migration Template

For services not specifically covered:

```python
# scripts/generic_migration.py
import json
import csv
from datetime import datetime

def migrate_from_csv(input_file, url_column, code_column, date_column=None):
    """Generic CSV migration"""
    links = []

    with open(input_file) as f:
        reader = csv.DictReader(f)
        for row in reader:
            link = {
                'original_url': row[url_column],
                'custom_code': row.get(code_column),
                'metadata': {
                    'migrated_from': 'csv_import',
                    'original_row': dict(row)
                }
            }
            if date_column and row.get(date_column):
                link['created_at'] = row[date_column]

            links.append(link)

    return links

def migrate_from_json(input_file, url_path, code_path):
    """Generic JSON migration"""
    with open(input_file) as f:
        data = json.load(f)

    links = []
    items = data if isinstance(data, list) else data.get('links', data.get('data', []))

    for item in items:
        link = {
            'original_url': get_nested(item, url_path),
            'custom_code': get_nested(item, code_path),
            'metadata': {
                'migrated_from': 'json_import',
                'original_data': item
            }
        }
        links.append(link)

    return links

def get_nested(obj, path):
    """Get nested value from dict using dot notation"""
    keys = path.split('.')
    for key in keys:
        obj = obj.get(key)
        if obj is None:
            return None
    return obj

# Usage
links = migrate_from_csv(
    'export.csv',
    url_column='destination_url',
    code_column='short_code',
    date_column='created_date'
)

with open('linkrift_import.json', 'w') as f:
    json.dump(links, f, indent=2)
```

---

## Data Import/Export

### Import Formats

Linkrift accepts the following import formats:

**JSON Format:**

```json
{
  "links": [
    {
      "original_url": "https://example.com/page",
      "custom_code": "abc123",
      "title": "Example Page",
      "created_at": "2025-01-15T10:00:00Z",
      "expires_at": null,
      "tags": ["marketing", "campaign-2025"],
      "metadata": {
        "source": "migration",
        "original_id": "xyz789"
      }
    }
  ]
}
```

**CSV Format:**

```csv
original_url,custom_code,title,created_at,tags
https://example.com/page,abc123,Example Page,2025-01-15T10:00:00Z,"marketing,campaign"
https://example.com/another,def456,Another Page,2025-01-16T10:00:00Z,
```

### Export Options

**Export via Dashboard:**
1. Go to Settings > Export
2. Select date range and filters
3. Choose format (JSON, CSV)
4. Include analytics data (optional)
5. Download

**Export via API:**

```bash
# Export all links
curl -X GET "https://api.linkrift.io/v1/links/export?format=json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o linkrift_export.json

# Export with filters
curl -X GET "https://api.linkrift.io/v1/links/export?format=csv&created_after=2025-01-01&tags=marketing" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o linkrift_export.csv

# Export with analytics
curl -X GET "https://api.linkrift.io/v1/links/export?include_analytics=true" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o linkrift_full_export.json
```

### API Migration

Update your applications to use Linkrift API:

**Before (Bitly API):**

```python
import requests

def create_short_link(url):
    response = requests.post(
        'https://api-ssl.bitly.com/v4/shorten',
        headers={'Authorization': f'Bearer {BITLY_TOKEN}'},
        json={'long_url': url}
    )
    return response.json()['link']
```

**After (Linkrift API):**

```python
import requests

def create_short_link(url):
    response = requests.post(
        'https://api.linkrift.io/v1/links',
        headers={'Authorization': f'Bearer {LINKRIFT_TOKEN}'},
        json={'url': url}
    )
    return response.json()['short_url']
```

**API Mapping:**

| Operation | Bitly | Linkrift |
|-----------|-------|----------|
| Create link | POST /v4/shorten | POST /v1/links |
| Get link | GET /v4/bitlinks/{id} | GET /v1/links/{code} |
| Update link | PATCH /v4/bitlinks/{id} | PUT /v1/links/{code} |
| Delete link | DELETE /v4/bitlinks/{id} | DELETE /v1/links/{code} |
| Get clicks | GET /v4/bitlinks/{id}/clicks | GET /v1/links/{code}/analytics |

---

## Domain Migration

### Step 1: Add Domain to Linkrift

```bash
curl -X POST "https://api.linkrift.io/v1/domains" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"domain": "links.yourdomain.com"}'
```

### Step 2: Verify DNS

Add DNS records as instructed:

```
# CNAME record
links.yourdomain.com  CNAME  redirect.linkrift.io

# Or A record (for apex domains)
yourdomain.com  A  203.0.113.50
```

### Step 3: SSL Certificate

Linkrift automatically provisions SSL via Let's Encrypt. Verify:

```bash
curl -I https://links.yourdomain.com/test123
```

### Step 4: Update Old Service

Point your domain away from the old service to prevent conflicts.

---

## Redirect Handling

### Option 1: DNS Redirect (Recommended)

Point old domain to Linkrift. All short codes will work if imported.

### Option 2: HTTP Redirect

If you can't transfer the domain, set up redirects:

```nginx
# For links that couldn't be migrated
server {
    server_name old-shortener.com;

    # Specific redirects
    location /abc123 {
        return 301 https://lnkr.ft/abc123;
    }

    # Catch-all to landing page
    location / {
        return 301 https://linkrift.io/migrated;
    }
}
```

### Option 3: Parallel Running

Run both services temporarily:

1. Import links to Linkrift
2. Keep old service running
3. Gradually update references
4. Monitor traffic on both
5. Shut down old service after transition period

---

## Preserving Analytics

### Export Historical Analytics

```python
# scripts/export_analytics.py
import requests
import json
from datetime import datetime, timedelta

def export_bitly_analytics(bitlink, access_token):
    """Export click data from Bitly"""
    response = requests.get(
        f'https://api-ssl.bitly.com/v4/bitlinks/{bitlink}/clicks',
        headers={'Authorization': f'Bearer {access_token}'},
        params={'unit': 'day', 'units': -1}
    )
    return response.json()

# Export all link analytics
with open('bitly_export.json') as f:
    links = json.load(f)

analytics_data = []
for link in links:
    clicks = export_bitly_analytics(link['id'], BITLY_TOKEN)
    analytics_data.append({
        'short_code': link['custom_code'],
        'historical_clicks': clicks
    })

with open('historical_analytics.json', 'w') as f:
    json.dump(analytics_data, f, indent=2)
```

### Import Historical Data

```bash
curl -X POST "https://api.linkrift.io/v1/analytics/import" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d @historical_analytics.json
```

---

## Post-Migration Verification

### Verification Checklist

```bash
#!/bin/bash
# scripts/verify_migration.sh

echo "Verifying migration..."

# Test sample links
LINKS=("abc123" "def456" "xyz789")

for code in "${LINKS[@]}"; do
    echo "Testing $code..."

    # Check redirect works
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -L "https://lnkr.ft/$code")

    if [ "$STATUS" = "200" ]; then
        echo "  ✓ Redirect works"
    else
        echo "  ✗ Redirect failed (status: $STATUS)"
    fi

    # Check API access
    API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $LINKRIFT_TOKEN" \
        "https://api.linkrift.io/v1/links/$code")

    if [ "$API_STATUS" = "200" ]; then
        echo "  ✓ API access works"
    else
        echo "  ✗ API access failed (status: $API_STATUS)"
    fi
done

echo "Verification complete!"
```

### Monitor for Issues

```python
# scripts/monitor_migration.py
import requests
import time

def check_link(short_url, expected_destination):
    try:
        response = requests.head(short_url, allow_redirects=False, timeout=5)
        location = response.headers.get('Location', '')

        if expected_destination in location:
            return True, None
        else:
            return False, f"Unexpected redirect: {location}"
    except Exception as e:
        return False, str(e)

# Monitor critical links
critical_links = [
    ('https://lnkr.ft/abc123', 'example.com/landing'),
    ('https://lnkr.ft/campaign', 'example.com/promo'),
]

while True:
    for short_url, expected in critical_links:
        success, error = check_link(short_url, expected)
        if not success:
            print(f"ALERT: {short_url} - {error}")
            # Send notification

    time.sleep(60)  # Check every minute
```

---

## Rollback Plan

If issues occur, have a rollback plan ready:

### Immediate Rollback

1. **DNS Rollback**: Point domain back to old service
   ```bash
   # Restore old DNS records
   links.domain.com  CNAME  old-service.com
   ```

2. **Keep Old Service Active**: Don't delete old account until fully migrated

3. **API Rollback**: Keep old API tokens valid during transition

### Data Recovery

```bash
# Re-export from Linkrift if needed
curl -X GET "https://api.linkrift.io/v1/links/export" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o backup_export.json
```

### Communication Plan

- Notify users of any temporary issues
- Update status page
- Document lessons learned

---

## Getting Help

**Migration Support:**
- Documentation: [docs.linkrift.io/migration](https://docs.linkrift.io/migration)
- Email: migration@linkrift.io
- Enterprise: Dedicated migration specialist

**Common Issues:**
- [Troubleshooting Guide](./FAQ.md#migration-issues)
- [GitHub Issues](https://github.com/linkrift/linkrift/issues)
- [Community Discord](https://discord.gg/linkrift)

We're here to help make your migration smooth!
