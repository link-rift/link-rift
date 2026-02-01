# Linkrift CLI Tool Documentation

**Last Updated: 2025-01-24**

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Using Go Install](#using-go-install)
  - [Binary Download](#binary-download)
  - [Package Managers](#package-managers)
  - [Docker](#docker)
- [Authentication](#authentication)
  - [API Key Authentication](#api-key-authentication)
  - [Interactive Login](#interactive-login)
  - [Environment Variables](#environment-variables)
- [Commands](#commands)
  - [shorten](#shorten)
  - [list](#list)
  - [stats](#stats)
  - [bulk](#bulk)
  - [domains](#domains)
  - [config](#config)
- [Configuration File](#configuration-file)
- [Shell Completion](#shell-completion)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Linkrift CLI is a command-line tool for creating and managing short links directly from your terminal. It provides a fast, scriptable interface for all Linkrift operations.

### Features

- Create short links with a single command
- Bulk import URLs from files or stdin
- View link analytics and statistics
- Manage custom domains
- Shell completion for bash, zsh, fish, and PowerShell
- JSON output for scripting and automation
- Configuration file support for multiple profiles

---

## Installation

### Using Go Install

If you have Go 1.21 or later installed:

```bash
go install github.com/link-rift/link-rift/cmd/linkrift@latest
```

This will install the `linkrift` binary to your `$GOPATH/bin` directory. Make sure this directory is in your `PATH`.

### Binary Download

Download pre-built binaries from the [GitHub Releases](https://github.com/link-rift/link-rift/releases) page.

#### macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/link-rift/link-rift/releases/latest/download/linkrift_darwin_arm64.tar.gz | tar xz
sudo mv linkrift /usr/local/bin/

# Intel
curl -L https://github.com/link-rift/link-rift/releases/latest/download/linkrift_darwin_amd64.tar.gz | tar xz
sudo mv linkrift /usr/local/bin/
```

#### Linux

```bash
# x86_64
curl -L https://github.com/link-rift/link-rift/releases/latest/download/linkrift_linux_amd64.tar.gz | tar xz
sudo mv linkrift /usr/local/bin/

# ARM64
curl -L https://github.com/link-rift/link-rift/releases/latest/download/linkrift_linux_arm64.tar.gz | tar xz
sudo mv linkrift /usr/local/bin/
```

#### Windows

Download `linkrift_windows_amd64.zip` from the releases page and add the extracted directory to your `PATH`.

```powershell
# Using PowerShell
Invoke-WebRequest -Uri "https://github.com/link-rift/link-rift/releases/latest/download/linkrift_windows_amd64.zip" -OutFile "linkrift.zip"
Expand-Archive -Path "linkrift.zip" -DestinationPath "C:\Program Files\Linkrift"
# Add to PATH manually or via System Properties
```

### Package Managers

#### Homebrew (macOS/Linux)

```bash
brew tap link-rift/tap
brew install linkrift
```

#### Scoop (Windows)

```powershell
scoop bucket add linkrift https://github.com/link-rift/scoop-bucket
scoop install linkrift
```

#### APT (Debian/Ubuntu)

```bash
curl -fsSL https://apt.linkrift.io/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/linkrift-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/linkrift-archive-keyring.gpg] https://apt.linkrift.io stable main" | sudo tee /etc/apt/sources.list.d/linkrift.list
sudo apt update
sudo apt install linkrift
```

#### YUM/DNF (RHEL/Fedora/CentOS)

```bash
sudo rpm --import https://rpm.linkrift.io/gpg.key
sudo tee /etc/yum.repos.d/linkrift.repo << 'EOF'
[linkrift]
name=Linkrift
baseurl=https://rpm.linkrift.io/stable
enabled=1
gpgcheck=1
gpgkey=https://rpm.linkrift.io/gpg.key
EOF
sudo yum install linkrift
```

### Docker

```bash
docker pull linkrift/cli:latest

# Run with API key
docker run --rm -e LINKRIFT_API_KEY=lr_live_sk_your_api_key linkrift/cli shorten https://example.com

# Run with config file
docker run --rm -v ~/.linkrift:/root/.linkrift linkrift/cli shorten https://example.com
```

### Verify Installation

```bash
linkrift version
```

Expected output:

```
linkrift version 1.0.0 (commit: abc1234, built: 2025-01-24T00:00:00Z)
```

---

## Authentication

The CLI requires authentication to interact with the Linkrift API. You can authenticate using an API key or by logging in interactively.

### API Key Authentication

1. Generate an API key from the [Linkrift Dashboard](https://app.linkrift.io/settings/api-keys)
2. Configure the CLI with your API key:

```bash
linkrift config set api-key lr_live_sk_your_api_key
```

Or use the `--api-key` flag with any command:

```bash
linkrift shorten https://example.com --api-key lr_live_sk_your_api_key
```

### Interactive Login

Login with your Linkrift account credentials:

```bash
linkrift login
```

This will:
1. Open your browser to authenticate (or prompt for email/password in non-interactive mode)
2. Store the access token securely in your system's keychain
3. Automatically refresh the token when needed

```bash
# Non-interactive login
linkrift login --email user@example.com --password yourpassword

# Check login status
linkrift auth status

# Logout
linkrift logout
```

### Environment Variables

You can also use environment variables for authentication:

```bash
# API Key (recommended for CI/CD)
export LINKRIFT_API_KEY=lr_live_sk_your_api_key

# Or use access token
export LINKRIFT_ACCESS_TOKEN=v4.public.eyJzdWIiOiJ1c2VyXzEyMzQ1...
```

Environment variables take precedence over configuration file values.

---

## Commands

### shorten

Create a short link.

```bash
linkrift shorten <url> [flags]
```

#### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--slug` | `-s` | Custom short code |
| `--domain` | `-d` | Custom domain ID or hostname |
| `--title` | `-t` | Link title |
| `--tags` | | Comma-separated tags |
| `--expires` | `-e` | Expiration (e.g., "24h", "7d", "2025-12-31") |
| `--password` | `-p` | Password protect the link |
| `--utm-source` | | UTM source parameter |
| `--utm-medium` | | UTM medium parameter |
| `--utm-campaign` | | UTM campaign parameter |
| `--qr` | `-q` | Generate QR code |
| `--qr-size` | | QR code size in pixels (default: 256) |
| `--qr-format` | | QR code format: png, svg (default: png) |
| `--output` | `-o` | Output format: text, json, yaml (default: text) |
| `--copy` | `-c` | Copy short URL to clipboard |

#### Examples

```bash
# Basic shortening
linkrift shorten https://example.com/very/long/url

# With custom slug
linkrift shorten https://example.com/page --slug my-custom-slug

# With custom domain
linkrift shorten https://example.com/page --domain links.mycompany.com

# With expiration
linkrift shorten https://example.com/page --expires 7d

# With password protection
linkrift shorten https://example.com/page --password secret123

# With UTM parameters
linkrift shorten https://example.com/page \
  --utm-source newsletter \
  --utm-medium email \
  --utm-campaign january-2025

# With tags
linkrift shorten https://example.com/page --tags "marketing,campaign,2025"

# Generate QR code alongside
linkrift shorten https://example.com/page --qr --qr-size 512

# Copy to clipboard
linkrift shorten https://example.com/page --copy

# JSON output
linkrift shorten https://example.com/page --output json
```

#### Output Examples

**Text (default):**

```
Short URL: https://lrift.co/abc123
Original:  https://example.com/very/long/url
Created:   2025-01-24T12:00:00Z
```

**JSON:**

```json
{
  "id": "lnk_1234567890abcdef",
  "short_url": "https://lrift.co/abc123",
  "original_url": "https://example.com/very/long/url",
  "slug": "abc123",
  "created_at": "2025-01-24T12:00:00Z"
}
```

---

### list

List your short links.

```bash
linkrift list [flags]
```

#### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--limit` | `-l` | Number of links to show (default: 20) |
| `--all` | `-a` | Show all links (paginate through all) |
| `--domain` | `-d` | Filter by domain |
| `--tag` | `-t` | Filter by tag |
| `--search` | `-s` | Search in URL, slug, or title |
| `--sort` | | Sort by: created, clicks, updated (default: created) |
| `--order` | | Sort order: asc, desc (default: desc) |
| `--output` | `-o` | Output format: table, json, yaml, csv (default: table) |

#### Examples

```bash
# List recent links
linkrift list

# List all links
linkrift list --all

# Filter by tag
linkrift list --tag marketing

# Search
linkrift list --search "campaign"

# Sort by clicks
linkrift list --sort clicks --order desc

# Limit results
linkrift list --limit 50

# JSON output
linkrift list --output json

# CSV output (useful for exports)
linkrift list --all --output csv > links.csv
```

#### Output Examples

**Table (default):**

```
ID                        SHORT URL                   CLICKS  CREATED
lnk_1234567890abcdef      https://lrift.co/abc123     1,234   2025-01-24
lnk_abcdef1234567890      https://lrift.co/xyz789       567   2025-01-23
lnk_fedcba0987654321      https://lrift.co/my-link      89   2025-01-22

Showing 3 of 150 links. Use --all to see all.
```

**CSV:**

```csv
id,short_url,original_url,slug,clicks,created_at
lnk_1234567890abcdef,https://lrift.co/abc123,https://example.com/page,abc123,1234,2025-01-24T12:00:00Z
lnk_abcdef1234567890,https://lrift.co/xyz789,https://example.com/other,xyz789,567,2025-01-23T10:00:00Z
```

---

### stats

View analytics for a link or your account.

```bash
linkrift stats [link_id_or_slug] [flags]
```

#### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--period` | `-p` | Time period: 24h, 7d, 30d, 90d, all (default: 30d) |
| `--start` | | Start date (YYYY-MM-DD) |
| `--end` | | End date (YYYY-MM-DD) |
| `--interval` | `-i` | Data interval: hour, day, week, month (default: day) |
| `--output` | `-o` | Output format: table, json, chart (default: table) |

#### Examples

```bash
# View stats for a specific link
linkrift stats lnk_1234567890abcdef

# Or use the slug
linkrift stats abc123

# View last 7 days
linkrift stats abc123 --period 7d

# Custom date range
linkrift stats abc123 --start 2025-01-01 --end 2025-01-24

# Account-wide stats (no link specified)
linkrift stats

# ASCII chart visualization
linkrift stats abc123 --output chart

# JSON output
linkrift stats abc123 --output json
```

#### Output Examples

**Table (default):**

```
Link: https://lrift.co/abc123
Period: 2024-12-25 to 2025-01-24 (30 days)

SUMMARY
-------
Total Clicks:     12,345
Unique Visitors:   8,765
Avg Daily Clicks:    411

TOP REFERRERS
-------------
google.com         3,456 (28.0%)
twitter.com        2,345 (19.0%)
direct             1,890 (15.3%)
facebook.com       1,234 (10.0%)
linkedin.com         987 (8.0%)

TOP COUNTRIES
-------------
United States      5,678 (46.0%)
United Kingdom     2,345 (19.0%)
Germany            1,234 (10.0%)
Canada               987 (8.0%)
France               765 (6.2%)

DEVICES
-------
Desktop            6,543 (53.0%)
Mobile             4,567 (37.0%)
Tablet             1,235 (10.0%)
```

**Chart:**

```
Clicks over the last 30 days

  1,200 |                    *
  1,000 |                   * *
    800 |          *       *   *
    600 |    *    * *  *  *     * *
    400 |   * * **   **  **       *
    200 | **                       **
      0 +--------------------------------
        Dec 25    Jan 01    Jan 10    Jan 20
```

---

### bulk

Bulk create or manage links.

```bash
linkrift bulk <subcommand> [flags]
```

#### Subcommands

| Subcommand | Description |
|------------|-------------|
| `create` | Create links from file or stdin |
| `export` | Export links to file |
| `delete` | Delete multiple links |
| `update` | Update multiple links |

#### bulk create

```bash
linkrift bulk create [file] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Input format: csv, json, txt (auto-detected) |
| `--domain` | `-d` | Default domain for all links |
| `--tags` | | Default tags for all links |
| `--dry-run` | | Preview without creating |
| `--output` | `-o` | Output format: table, json, csv |

**CSV Format:**

```csv
url,slug,title,tags
https://example.com/page1,page-1,First Page,"marketing,campaign"
https://example.com/page2,,Second Page,marketing
https://example.com/page3,page-3,,
```

**JSON Format:**

```json
[
  {"url": "https://example.com/page1", "slug": "page-1", "title": "First Page"},
  {"url": "https://example.com/page2", "title": "Second Page"},
  {"url": "https://example.com/page3", "slug": "page-3"}
]
```

**TXT Format (one URL per line):**

```
https://example.com/page1
https://example.com/page2
https://example.com/page3
```

**Examples:**

```bash
# From CSV file
linkrift bulk create urls.csv

# From JSON file
linkrift bulk create urls.json --format json

# From stdin
cat urls.txt | linkrift bulk create -

# With default domain and tags
linkrift bulk create urls.csv --domain links.mycompany.com --tags "bulk,import"

# Dry run to preview
linkrift bulk create urls.csv --dry-run

# Export results to file
linkrift bulk create urls.csv --output csv > results.csv
```

#### bulk export

```bash
linkrift bulk export [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format: csv, json (default: csv) |
| `--output` | `-o` | Output file (default: stdout) |
| `--domain` | `-d` | Filter by domain |
| `--tag` | `-t` | Filter by tag |
| `--include-stats` | | Include click statistics |

**Examples:**

```bash
# Export all links to CSV
linkrift bulk export > all-links.csv

# Export to JSON
linkrift bulk export --format json > all-links.json

# Export specific domain
linkrift bulk export --domain links.mycompany.com > company-links.csv

# Export with stats
linkrift bulk export --include-stats > links-with-stats.csv
```

#### bulk delete

```bash
linkrift bulk delete [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | File with link IDs (one per line) |
| `--tag` | `-t` | Delete all links with tag |
| `--older-than` | | Delete links older than (e.g., "90d") |
| `--confirm` | | Skip confirmation prompt |
| `--dry-run` | | Preview without deleting |

**Examples:**

```bash
# Delete from file
linkrift bulk delete --file link-ids.txt

# Delete by tag
linkrift bulk delete --tag "deprecated"

# Delete old links
linkrift bulk delete --older-than 365d

# Dry run
linkrift bulk delete --tag "deprecated" --dry-run
```

---

### domains

Manage custom domains.

```bash
linkrift domains <subcommand> [flags]
```

#### Subcommands

| Subcommand | Description |
|------------|-------------|
| `list` | List all domains |
| `add` | Add a new domain |
| `verify` | Verify domain DNS |
| `default` | Set default domain |
| `remove` | Remove a domain |

#### Examples

```bash
# List domains
linkrift domains list

# Add a domain
linkrift domains add links.mycompany.com

# Verify domain
linkrift domains verify dom_1234567890abcdef

# Or verify by hostname
linkrift domains verify links.mycompany.com

# Set as default
linkrift domains default links.mycompany.com

# Remove domain
linkrift domains remove links.mycompany.com
```

#### Output Examples

**domains list:**

```
HOSTNAME                  STATUS      DEFAULT  SSL       LINKS
lrift.co                  verified    yes      active    1,234
links.mycompany.com       verified    no       active      567
promo.example.com         pending     no       pending       0

DNS Records for pending domains:
  promo.example.com:
    CNAME: promo.example.com -> redirect.linkrift.io
    TXT:   _linkrift-verify.promo.example.com -> verify-abc123
```

**domains add:**

```
Domain added: links.mycompany.com

To verify your domain, add the following DNS records:

  CNAME Record:
    Name:  links.mycompany.com
    Value: redirect.linkrift.io

  Verification Record:
    Type:  TXT
    Name:  _linkrift-verify.links.mycompany.com
    Value: verify-abc123def456

After adding DNS records, run:
  linkrift domains verify links.mycompany.com
```

---

### config

Manage CLI configuration.

```bash
linkrift config <subcommand> [key] [value]
```

#### Subcommands

| Subcommand | Description |
|------------|-------------|
| `show` | Show current configuration |
| `set` | Set a configuration value |
| `get` | Get a configuration value |
| `unset` | Remove a configuration value |
| `path` | Show configuration file path |
| `edit` | Open configuration in editor |

#### Configuration Keys

| Key | Description |
|-----|-------------|
| `api-key` | API key for authentication |
| `default-domain` | Default domain for new links |
| `default-tags` | Default tags for new links (comma-separated) |
| `output-format` | Default output format (text, json, yaml) |
| `color` | Enable/disable colored output (true/false) |
| `workspace` | Default workspace ID |

#### Examples

```bash
# Show all configuration
linkrift config show

# Set API key
linkrift config set api-key lr_live_sk_your_api_key

# Set default domain
linkrift config set default-domain links.mycompany.com

# Set default tags
linkrift config set default-tags "cli,automated"

# Get a value
linkrift config get default-domain

# Unset a value
linkrift config unset default-domain

# Show config file path
linkrift config path

# Edit configuration in default editor
linkrift config edit
```

---

## Configuration File

The CLI stores configuration in a YAML file. The location depends on your operating system:

| OS | Path |
|----|------|
| macOS | `~/.config/linkrift/config.yaml` |
| Linux | `~/.config/linkrift/config.yaml` |
| Windows | `%APPDATA%\linkrift\config.yaml` |

### Configuration File Format

```yaml
# Linkrift CLI Configuration
# ~/.config/linkrift/config.yaml

# Authentication
api_key: lr_live_sk_your_api_key

# Default settings
defaults:
  domain: links.mycompany.com
  tags:
    - cli
    - automated
  output_format: text

# Appearance
color: true

# Workspace
workspace: ws_1234567890abcdef

# Profiles for multiple accounts
profiles:
  personal:
    api_key: lr_live_sk_personal_key
    defaults:
      domain: lrift.co

  work:
    api_key: lr_live_sk_work_key
    defaults:
      domain: links.company.com
    workspace: ws_company123
```

### Using Profiles

Switch between profiles using the `--profile` flag or `LINKRIFT_PROFILE` environment variable:

```bash
# Use a specific profile
linkrift shorten https://example.com --profile work

# Set default profile via environment
export LINKRIFT_PROFILE=work
linkrift shorten https://example.com
```

---

## Shell Completion

The CLI supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc
eval "$(linkrift completion bash)"

# Or generate completion file
linkrift completion bash > /etc/bash_completion.d/linkrift
```

### Zsh

```bash
# Add to ~/.zshrc
eval "$(linkrift completion zsh)"

# Or generate completion file
linkrift completion zsh > "${fpath[1]}/_linkrift"
```

### Fish

```bash
# Generate completion file
linkrift completion fish > ~/.config/fish/completions/linkrift.fish
```

### PowerShell

```powershell
# Add to your PowerShell profile
linkrift completion powershell | Out-String | Invoke-Expression

# Or save to a file
linkrift completion powershell > linkrift.ps1
# Then add to profile: . path\to\linkrift.ps1
```

---

## Usage Examples

### CI/CD Integration

**GitHub Actions:**

```yaml
name: Create Release Links

on:
  release:
    types: [published]

jobs:
  create-links:
    runs-on: ubuntu-latest
    steps:
      - name: Install Linkrift CLI
        run: |
          curl -L https://github.com/link-rift/link-rift/releases/latest/download/linkrift_linux_amd64.tar.gz | tar xz
          sudo mv linkrift /usr/local/bin/

      - name: Create short link for release
        env:
          LINKRIFT_API_KEY: ${{ secrets.LINKRIFT_API_KEY }}
        run: |
          linkrift shorten "${{ github.event.release.html_url }}" \
            --slug "release-${{ github.event.release.tag_name }}" \
            --tags "release,github" \
            --output json > link.json

          echo "Short URL: $(jq -r '.short_url' link.json)"
```

**GitLab CI:**

```yaml
create-short-link:
  image: alpine:latest
  script:
    - wget -qO- https://github.com/link-rift/link-rift/releases/latest/download/linkrift_linux_amd64.tar.gz | tar xz
    - ./linkrift shorten "$CI_PAGES_URL" --slug "docs-$CI_COMMIT_SHORT_SHA"
  variables:
    LINKRIFT_API_KEY: $LINKRIFT_API_KEY
```

### Scripting Examples

**Create links from a web page:**

```bash
#!/bin/bash
# Extract URLs from a web page and create short links

curl -s "https://example.com/sitemap.xml" | \
  grep -oP '(?<=<loc>)[^<]+' | \
  linkrift bulk create - --tags "sitemap,automated"
```

**Monitor link performance:**

```bash
#!/bin/bash
# Get daily stats for top links

linkrift list --sort clicks --limit 10 --output json | \
  jq -r '.[].id' | \
  while read id; do
    echo "=== Stats for $id ==="
    linkrift stats "$id" --period 24h
    echo
  done
```

**Export and backup links:**

```bash
#!/bin/bash
# Weekly backup of all links

DATE=$(date +%Y-%m-%d)
linkrift bulk export --format json --include-stats > "backup-$DATE.json"
gzip "backup-$DATE.json"
```

**Create links with QR codes:**

```bash
#!/bin/bash
# Create links with QR codes for event materials

URLS=(
  "https://example.com/event/registration"
  "https://example.com/event/schedule"
  "https://example.com/event/speakers"
)

for url in "${URLS[@]}"; do
  slug=$(basename "$url")
  linkrift shorten "$url" \
    --slug "event-$slug" \
    --qr \
    --qr-size 1024 \
    --qr-format svg \
    --tags "event,qr"
done
```

---

## Troubleshooting

### Common Issues

#### "command not found: linkrift"

Make sure the binary is in your PATH:

```bash
# Check if binary exists
which linkrift

# If using go install, ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

#### "authentication required"

```bash
# Check if API key is configured
linkrift config get api-key

# Set API key
linkrift config set api-key lr_live_sk_your_api_key

# Or use environment variable
export LINKRIFT_API_KEY=lr_live_sk_your_api_key
```

#### "rate limit exceeded"

The CLI respects rate limits and will automatically retry with backoff. For bulk operations, use the `bulk` command which handles rate limiting:

```bash
# Instead of a loop with shorten, use bulk create
linkrift bulk create urls.csv
```

#### "invalid API key"

- Verify the key is correct in your dashboard
- Check if the key has been revoked
- Ensure you're using the correct key for the environment (live vs test)

#### "SSL certificate error"

If you're behind a corporate proxy:

```bash
# Skip SSL verification (not recommended for production)
linkrift shorten https://example.com --insecure

# Or set custom CA bundle
export SSL_CERT_FILE=/path/to/ca-bundle.crt
```

### Debug Mode

Enable verbose output for debugging:

```bash
# Enable debug mode
linkrift shorten https://example.com --debug

# Or set environment variable
export LINKRIFT_DEBUG=true
linkrift shorten https://example.com
```

Debug output includes:
- Request/response headers
- API endpoint URLs
- Timing information
- Rate limit status

### Getting Help

```bash
# General help
linkrift help

# Command-specific help
linkrift help shorten
linkrift shorten --help

# Version information
linkrift version
```

---

## Additional Resources

- **API Documentation**: [docs.linkrift.io/api](https://docs.linkrift.io/api)
- **SDK Documentation**: [docs.linkrift.io/sdk](https://docs.linkrift.io/sdk)
- **GitHub Repository**: [github.com/link-rift/link-rift](https://github.com/link-rift/link-rift)
- **Issues**: [github.com/link-rift/link-rift/issues](https://github.com/link-rift/link-rift/issues)
- **Changelog**: [docs.linkrift.io/changelog](https://docs.linkrift.io/changelog)
- **Support**: [support@linkrift.io](mailto:support@linkrift.io)

---

*This documentation is generated from the Linkrift CLI specifications. For the most up-to-date information, visit [docs.linkrift.io/cli](https://docs.linkrift.io/cli).*
