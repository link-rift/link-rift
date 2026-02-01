# Maintenance Guide

**Last Updated: 2025-01-24**

This document covers regular maintenance tasks, dependency updates, database maintenance, backup procedures, and disaster recovery for Linkrift.

---

## Table of Contents

- [Overview](#overview)
- [Regular Maintenance Tasks](#regular-maintenance-tasks)
  - [Daily Tasks](#daily-tasks)
  - [Weekly Tasks](#weekly-tasks)
  - [Monthly Tasks](#monthly-tasks)
- [Go Dependency Updates](#go-dependency-updates)
  - [Checking for Updates](#checking-for-updates)
  - [Updating Dependencies](#updating-dependencies)
  - [Security Updates](#security-updates)
  - [Dependency Audit](#dependency-audit)
- [Database Maintenance](#database-maintenance)
  - [PostgreSQL Maintenance](#postgresql-maintenance)
  - [VACUUM Operations](#vacuum-operations)
  - [REINDEX Operations](#reindex-operations)
  - [Statistics Updates](#statistics-updates)
- [ClickHouse Maintenance](#clickhouse-maintenance)
  - [Data Compaction](#data-compaction)
  - [Partition Management](#partition-management)
  - [Query Optimization](#query-optimization)
- [Redis Maintenance](#redis-maintenance)
  - [Memory Management](#memory-management)
  - [Persistence Configuration](#persistence-configuration)
  - [Key Expiration](#key-expiration)
- [Log Rotation](#log-rotation)
  - [Application Logs](#application-logs)
  - [Access Logs](#access-logs)
  - [Audit Logs](#audit-logs)
- [Backup Procedures](#backup-procedures)
  - [PostgreSQL Backups](#postgresql-backups)
  - [ClickHouse Backups](#clickhouse-backups)
  - [Redis Backups](#redis-backups)
  - [Configuration Backups](#configuration-backups)
- [Disaster Recovery](#disaster-recovery)
  - [Recovery Plan](#recovery-plan)
  - [PostgreSQL Recovery](#postgresql-recovery)
  - [ClickHouse Recovery](#clickhouse-recovery)
  - [Full System Recovery](#full-system-recovery)
- [Health Checks](#health-checks)
- [Maintenance Scripts](#maintenance-scripts)

---

## Overview

Regular maintenance ensures Linkrift operates reliably and efficiently. This guide covers:

| Category | Frequency | Purpose |
|----------|-----------|---------|
| Database maintenance | Daily/Weekly | Performance optimization |
| Dependency updates | Weekly/Monthly | Security and stability |
| Backups | Daily | Data protection |
| Log rotation | Daily | Disk space management |
| Health checks | Continuous | Early issue detection |

---

## Regular Maintenance Tasks

### Daily Tasks

```bash
#!/bin/bash
# scripts/maintenance/daily.sh

set -e

echo "=== Linkrift Daily Maintenance - $(date) ==="

# 1. Health check
echo "Checking service health..."
curl -sf http://localhost:8080/health || echo "WARNING: Health check failed"

# 2. Check disk space
echo "Checking disk space..."
df -h | grep -E '(Filesystem|/dev/)'
if [ $(df / | tail -1 | awk '{print $5}' | tr -d '%') -gt 85 ]; then
    echo "WARNING: Disk usage above 85%"
fi

# 3. Check log sizes
echo "Checking log sizes..."
du -sh /var/log/linkrift/*

# 4. Verify backups completed
echo "Verifying backup completion..."
BACKUP_DATE=$(date +%Y-%m-%d)
if [ ! -f "/backups/postgres/linkrift_${BACKUP_DATE}.sql.gz" ]; then
    echo "WARNING: Today's PostgreSQL backup not found"
fi

# 5. Check error rates
echo "Checking error rates..."
ERROR_COUNT=$(grep -c "level\":\"error" /var/log/linkrift/app.log 2>/dev/null || echo "0")
if [ "$ERROR_COUNT" -gt 100 ]; then
    echo "WARNING: High error count today: $ERROR_COUNT"
fi

echo "=== Daily maintenance complete ==="
```

### Weekly Tasks

```bash
#!/bin/bash
# scripts/maintenance/weekly.sh

set -e

echo "=== Linkrift Weekly Maintenance - $(date) ==="

# 1. PostgreSQL maintenance
echo "Running PostgreSQL maintenance..."
psql -U linkrift -d linkrift -c "
    -- Update statistics
    ANALYZE VERBOSE;

    -- Check for bloated tables
    SELECT schemaname, tablename,
           pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
    FROM pg_tables
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
    LIMIT 10;
"

# 2. Check for expired links
echo "Cleaning expired links..."
psql -U linkrift -d linkrift -c "
    DELETE FROM links
    WHERE expires_at IS NOT NULL
    AND expires_at < NOW() - INTERVAL '30 days';
"

# 3. ClickHouse optimization
echo "Optimizing ClickHouse tables..."
clickhouse-client --query "OPTIMIZE TABLE clicks FINAL"

# 4. Redis memory check
echo "Checking Redis memory..."
redis-cli INFO memory | grep -E "(used_memory_human|maxmemory_human)"

# 5. Check dependency updates
echo "Checking for dependency updates..."
cd /app && go list -u -m all 2>/dev/null | grep '\[' | head -20

# 6. SSL certificate expiry check
echo "Checking SSL certificates..."
echo | openssl s_client -servername linkrift.io -connect linkrift.io:443 2>/dev/null | \
    openssl x509 -noout -dates

echo "=== Weekly maintenance complete ==="
```

### Monthly Tasks

```bash
#!/bin/bash
# scripts/maintenance/monthly.sh

set -e

echo "=== Linkrift Monthly Maintenance - $(date) ==="

# 1. Full VACUUM on PostgreSQL
echo "Running full VACUUM..."
psql -U linkrift -d linkrift -c "VACUUM FULL VERBOSE;"

# 2. Reindex important tables
echo "Reindexing tables..."
psql -U linkrift -d linkrift -c "
    REINDEX TABLE links;
    REINDEX TABLE users;
    REINDEX TABLE clicks;
"

# 3. Archive old analytics data
echo "Archiving old analytics data..."
ARCHIVE_DATE=$(date -d '90 days ago' +%Y-%m-%d)
clickhouse-client --query "
    ALTER TABLE clicks
    DROP PARTITION WHERE toDate(clicked_at) < '${ARCHIVE_DATE}'
"

# 4. Clean old backups
echo "Cleaning old backups..."
find /backups -type f -mtime +90 -delete

# 5. Security audit
echo "Running security audit..."
cd /app && gosec ./...

# 6. Dependency vulnerability scan
echo "Scanning for vulnerabilities..."
cd /app && govulncheck ./...

# 7. Generate monthly report
echo "Generating monthly report..."
psql -U linkrift -d linkrift -c "
    SELECT
        date_trunc('day', created_at) as date,
        COUNT(*) as links_created,
        COUNT(DISTINCT user_id) as active_users
    FROM links
    WHERE created_at > NOW() - INTERVAL '30 days'
    GROUP BY date_trunc('day', created_at)
    ORDER BY date;
"

echo "=== Monthly maintenance complete ==="
```

---

## Go Dependency Updates

### Checking for Updates

```bash
# List all dependencies with available updates
go list -u -m all

# Check for specific module updates
go list -u -m github.com/go-chi/chi/v5

# Show direct dependencies only
go list -u -m -f '{{if not .Indirect}}{{.}}{{end}}' all

# Check for major version updates
go list -u -m all 2>&1 | grep -E '\[v[0-9]+\.'
```

### Updating Dependencies

```bash
#!/bin/bash
# scripts/maintenance/update-deps.sh

set -e

echo "=== Updating Go Dependencies ==="

# Backup go.mod and go.sum
cp go.mod go.mod.backup
cp go.sum go.sum.backup

# Update all dependencies to latest minor/patch versions
go get -u ./...

# Tidy up
go mod tidy

# Verify the build
go build ./...

# Run tests
go test ./...

# Show changes
echo "=== Changes to go.mod ==="
diff go.mod.backup go.mod || true

# Cleanup backups if successful
rm go.mod.backup go.sum.backup

echo "=== Update complete ==="
```

### Security Updates

```go
// scripts/security-check.go
package main

import (
    "fmt"
    "os/exec"
    "strings"
)

func main() {
    // Run govulncheck
    cmd := exec.Command("govulncheck", "./...")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Println("Vulnerabilities found:")
        fmt.Println(string(output))

        // Parse and report critical vulnerabilities
        lines := strings.Split(string(output), "\n")
        for _, line := range lines {
            if strings.Contains(line, "CRITICAL") || strings.Contains(line, "HIGH") {
                fmt.Printf("ALERT: %s\n", line)
            }
        }
    } else {
        fmt.Println("No vulnerabilities found")
    }
}
```

```bash
# Install and run govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Check specific packages
govulncheck -test ./internal/...

# JSON output for CI/CD
govulncheck -json ./... > vulnerabilities.json
```

### Dependency Audit

```bash
#!/bin/bash
# scripts/maintenance/audit-deps.sh

echo "=== Dependency Audit Report ==="
echo "Date: $(date)"
echo ""

# Count dependencies
echo "## Dependency Count"
echo "Direct: $(go list -m -f '{{if not .Indirect}}1{{end}}' all | wc -l)"
echo "Indirect: $(go list -m -f '{{if .Indirect}}1{{end}}' all | wc -l)"
echo ""

# List outdated dependencies
echo "## Outdated Dependencies"
go list -u -m all 2>&1 | grep '\[' | while read line; do
    echo "- $line"
done
echo ""

# Check for replaced modules
echo "## Replaced Modules"
grep "replace" go.mod || echo "None"
echo ""

# Module graph summary
echo "## Dependency Tree (top-level)"
go mod graph | head -20
```

---

## Database Maintenance

### PostgreSQL Maintenance

```sql
-- Check database size
SELECT pg_size_pretty(pg_database_size('linkrift')) as db_size;

-- Check table sizes
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as index_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check for unused indexes
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_scan = 0
AND schemaname = 'public';

-- Check for missing indexes
SELECT
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    seq_tup_read / NULLIF(seq_scan, 0) as avg_seq_tup_read
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_tup_read DESC
LIMIT 10;

-- Check for long-running queries
SELECT
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';
```

### VACUUM Operations

```bash
#!/bin/bash
# scripts/maintenance/vacuum.sh

# Standard VACUUM (can run during operation)
psql -U linkrift -d linkrift << 'EOF'
-- VACUUM with verbose output
VACUUM VERBOSE;

-- VACUUM specific high-churn tables
VACUUM VERBOSE links;
VACUUM VERBOSE clicks;
VACUUM VERBOSE sessions;
EOF

# Full VACUUM (requires exclusive lock - schedule during maintenance window)
# This reclaims more space but locks tables
psql -U linkrift -d linkrift -c "VACUUM FULL VERBOSE links;"
```

```sql
-- Automated VACUUM tuning
ALTER TABLE links SET (
    autovacuum_vacuum_scale_factor = 0.05,  -- VACUUM when 5% of tuples are dead
    autovacuum_analyze_scale_factor = 0.02,  -- ANALYZE when 2% of tuples changed
    autovacuum_vacuum_cost_delay = 10        -- Reduce I/O impact
);

ALTER TABLE clicks SET (
    autovacuum_vacuum_scale_factor = 0.01,  -- More aggressive for high-volume table
    autovacuum_analyze_scale_factor = 0.005
);
```

### REINDEX Operations

```sql
-- Check index bloat
SELECT
    indexrelname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    idx_scan as index_scans
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Reindex specific index (locks table briefly)
REINDEX INDEX CONCURRENTLY idx_links_short_code;

-- Reindex entire table
REINDEX TABLE CONCURRENTLY links;

-- Reindex entire database (use with caution)
REINDEX DATABASE CONCURRENTLY linkrift;
```

```bash
#!/bin/bash
# scripts/maintenance/reindex.sh

# Concurrent reindex (PostgreSQL 12+)
psql -U linkrift -d linkrift << 'EOF'
-- Reindex critical indexes concurrently
REINDEX INDEX CONCURRENTLY idx_links_short_code;
REINDEX INDEX CONCURRENTLY idx_links_user_id;
REINDEX INDEX CONCURRENTLY idx_links_created_at;
REINDEX INDEX CONCURRENTLY idx_clicks_link_id;
REINDEX INDEX CONCURRENTLY idx_clicks_clicked_at;
EOF
```

### Statistics Updates

```sql
-- Update statistics for query planner
ANALYZE VERBOSE;

-- Update specific tables
ANALYZE VERBOSE links;
ANALYZE VERBOSE users;
ANALYZE VERBOSE clicks;

-- Check statistics
SELECT
    schemaname,
    tablename,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze,
    vacuum_count,
    analyze_count
FROM pg_stat_user_tables
WHERE schemaname = 'public';
```

---

## ClickHouse Maintenance

### Data Compaction

```sql
-- Check part counts and sizes
SELECT
    database,
    table,
    count() as parts,
    sum(rows) as total_rows,
    formatReadableSize(sum(bytes_on_disk)) as size
FROM system.parts
WHERE active AND database = 'linkrift'
GROUP BY database, table
ORDER BY sum(bytes_on_disk) DESC;

-- Manual optimization (merges parts)
OPTIMIZE TABLE clicks FINAL;

-- Optimize specific partition
OPTIMIZE TABLE clicks PARTITION '2025-01' FINAL;

-- Check merge progress
SELECT * FROM system.merges;
```

### Partition Management

```sql
-- View partitions
SELECT
    partition,
    count() as parts,
    sum(rows) as rows,
    formatReadableSize(sum(bytes_on_disk)) as size
FROM system.parts
WHERE active AND table = 'clicks'
GROUP BY partition
ORDER BY partition;

-- Drop old partitions (data retention)
ALTER TABLE clicks DROP PARTITION '2024-01';

-- Detach partition (keeps data on disk)
ALTER TABLE clicks DETACH PARTITION '2024-06';

-- Reattach partition
ALTER TABLE clicks ATTACH PARTITION '2024-06';

-- Move partition to different disk
ALTER TABLE clicks MOVE PARTITION '2024-12' TO DISK 'cold_storage';
```

### Query Optimization

```sql
-- Check slow queries
SELECT
    query_id,
    user,
    query,
    elapsed,
    read_rows,
    read_bytes,
    memory_usage
FROM system.query_log
WHERE type = 'QueryFinish'
AND event_date = today()
AND elapsed > 1
ORDER BY elapsed DESC
LIMIT 20;

-- Analyze query
EXPLAIN SYNTAX SELECT * FROM clicks WHERE link_id = 'abc123';
EXPLAIN PLAN SELECT * FROM clicks WHERE link_id = 'abc123';
EXPLAIN PIPELINE SELECT * FROM clicks WHERE link_id = 'abc123';

-- Check table engine settings
SELECT
    name,
    value
FROM system.settings
WHERE name LIKE '%merge%';
```

---

## Redis Maintenance

### Memory Management

```bash
#!/bin/bash
# scripts/maintenance/redis-memory.sh

# Check memory usage
redis-cli INFO memory

# Get memory stats
redis-cli MEMORY STATS

# Find big keys
redis-cli --bigkeys

# Memory doctor recommendations
redis-cli MEMORY DOCTOR

# Sample key memory usage
redis-cli MEMORY USAGE "link:abc123"
```

```bash
# Redis configuration for memory management
# /etc/redis/redis.conf

# Set max memory (adjust based on server)
maxmemory 4gb

# Eviction policy for cache use case
maxmemory-policy allkeys-lru

# Sample keys for eviction
maxmemory-samples 10
```

### Persistence Configuration

```bash
# Redis persistence settings
# /etc/redis/redis.conf

# RDB snapshots
save 900 1      # Save if at least 1 key changed in 900 seconds
save 300 10     # Save if at least 10 keys changed in 300 seconds
save 60 10000   # Save if at least 10000 keys changed in 60 seconds

# RDB file settings
dbfilename dump.rdb
dir /var/lib/redis

# AOF persistence
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec

# AOF rewrite settings
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

```bash
#!/bin/bash
# Manual Redis maintenance

# Trigger RDB snapshot
redis-cli BGSAVE

# Check last save status
redis-cli LASTSAVE

# Trigger AOF rewrite
redis-cli BGREWRITEAOF

# Check persistence status
redis-cli INFO persistence
```

### Key Expiration

```go
// internal/cache/maintenance.go
package cache

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type CacheMaintenance struct {
    client *redis.Client
    logger *zap.Logger
}

// ScanAndCleanup removes orphaned or invalid keys
func (m *CacheMaintenance) ScanAndCleanup(ctx context.Context) error {
    var cursor uint64
    var cleaned int64

    for {
        var keys []string
        var err error
        keys, cursor, err = m.client.Scan(ctx, cursor, "link:*", 100).Result()
        if err != nil {
            return err
        }

        for _, key := range keys {
            // Check if key has TTL
            ttl, err := m.client.TTL(ctx, key).Result()
            if err != nil {
                continue
            }

            // Remove keys without TTL (orphaned)
            if ttl == -1 {
                m.client.Del(ctx, key)
                cleaned++
            }
        }

        if cursor == 0 {
            break
        }
    }

    m.logger.Info("Cache cleanup completed",
        zap.Int64("keys_cleaned", cleaned),
    )
    return nil
}

// SetKeyExpiration ensures all link keys have TTL
func (m *CacheMaintenance) SetKeyExpiration(ctx context.Context, pattern string, ttl time.Duration) error {
    var cursor uint64

    for {
        var keys []string
        var err error
        keys, cursor, err = m.client.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }

        pipe := m.client.Pipeline()
        for _, key := range keys {
            pipe.Expire(ctx, key, ttl)
        }
        _, err = pipe.Exec(ctx)
        if err != nil {
            return err
        }

        if cursor == 0 {
            break
        }
    }

    return nil
}
```

---

## Log Rotation

### Application Logs

```bash
# /etc/logrotate.d/linkrift
/var/log/linkrift/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 linkrift linkrift
    sharedscripts
    postrotate
        # Send SIGUSR1 to reopen log files
        /usr/bin/pkill -USR1 -f linkrift || true
    endscript
}
```

```go
// internal/logger/rotation.go
package logger

import (
    "os"
    "os/signal"
    "syscall"

    "go.uber.org/zap"
)

// SetupLogRotation handles SIGUSR1 for log rotation
func SetupLogRotation() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGUSR1)

    go func() {
        for range c {
            // Sync current logger
            Sync()

            // Reinitialize logger (reopens files)
            if err := Init(os.Getenv("APP_ENV")); err != nil {
                Log.Error("Failed to reinitialize logger", zap.Error(err))
            }
        }
    }()
}
```

### Access Logs

```bash
# /etc/logrotate.d/linkrift-access
/var/log/linkrift/access.log {
    daily
    rotate 90
    compress
    delaycompress
    missingok
    notifempty
    create 0640 linkrift linkrift
    dateext
    dateformat -%Y%m%d
}
```

### Audit Logs

```bash
# /etc/logrotate.d/linkrift-audit
/var/log/linkrift/audit.log {
    daily
    rotate 365
    compress
    delaycompress
    missingok
    notifempty
    create 0640 linkrift linkrift
    dateext
    dateformat -%Y%m%d
    # Keep audit logs longer for compliance
    maxage 730
}
```

---

## Backup Procedures

### PostgreSQL Backups

```bash
#!/bin/bash
# scripts/backup/postgres-backup.sh

set -e

BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y-%m-%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/linkrift_${DATE}.sql.gz"
RETENTION_DAYS=30

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

# Create backup with compression
echo "Starting PostgreSQL backup..."
pg_dump -U linkrift -d linkrift \
    --format=custom \
    --compress=9 \
    --file="${BACKUP_FILE%.gz}.dump"

# Also create SQL dump for portability
pg_dump -U linkrift -d linkrift | gzip > "$BACKUP_FILE"

# Verify backup
if pg_restore --list "${BACKUP_FILE%.gz}.dump" > /dev/null 2>&1; then
    echo "Backup verified successfully"
else
    echo "ERROR: Backup verification failed"
    exit 1
fi

# Calculate checksum
sha256sum "$BACKUP_FILE" > "${BACKUP_FILE}.sha256"

# Clean old backups
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +${RETENTION_DAYS} -delete
find "$BACKUP_DIR" -name "*.dump" -mtime +${RETENTION_DAYS} -delete
find "$BACKUP_DIR" -name "*.sha256" -mtime +${RETENTION_DAYS} -delete

# Upload to S3 (optional)
if [ -n "$S3_BUCKET" ]; then
    aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/postgres/"
    aws s3 cp "${BACKUP_FILE%.gz}.dump" "s3://${S3_BUCKET}/postgres/"
fi

echo "Backup completed: $BACKUP_FILE"
```

### ClickHouse Backups

```bash
#!/bin/bash
# scripts/backup/clickhouse-backup.sh

set -e

BACKUP_DIR="/backups/clickhouse"
DATE=$(date +%Y-%m-%d_%H%M%S)
BACKUP_NAME="linkrift_${DATE}"
RETENTION_DAYS=30

mkdir -p "$BACKUP_DIR"

# Create backup using clickhouse-backup tool
echo "Starting ClickHouse backup..."
clickhouse-backup create "$BACKUP_NAME"

# Verify backup
clickhouse-backup list | grep "$BACKUP_NAME"

# Compress backup
tar -czf "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" \
    -C /var/lib/clickhouse/backup "$BACKUP_NAME"

# Clean local backup
clickhouse-backup delete local "$BACKUP_NAME"

# Upload to S3
if [ -n "$S3_BUCKET" ]; then
    clickhouse-backup upload "$BACKUP_NAME"
fi

# Clean old backups
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +${RETENTION_DAYS} -delete

echo "ClickHouse backup completed: ${BACKUP_NAME}"
```

### Redis Backups

```bash
#!/bin/bash
# scripts/backup/redis-backup.sh

set -e

BACKUP_DIR="/backups/redis"
DATE=$(date +%Y-%m-%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/dump_${DATE}.rdb"
RETENTION_DAYS=7

mkdir -p "$BACKUP_DIR"

# Trigger background save
echo "Triggering Redis BGSAVE..."
redis-cli BGSAVE

# Wait for save to complete
while [ "$(redis-cli LASTSAVE)" == "$(redis-cli LASTSAVE)" ]; do
    sleep 1
done

# Copy dump file
cp /var/lib/redis/dump.rdb "$BACKUP_FILE"

# Compress
gzip "$BACKUP_FILE"

# Verify
if [ -f "${BACKUP_FILE}.gz" ]; then
    echo "Redis backup completed: ${BACKUP_FILE}.gz"
else
    echo "ERROR: Backup file not found"
    exit 1
fi

# Upload to S3
if [ -n "$S3_BUCKET" ]; then
    aws s3 cp "${BACKUP_FILE}.gz" "s3://${S3_BUCKET}/redis/"
fi

# Clean old backups
find "$BACKUP_DIR" -name "*.rdb.gz" -mtime +${RETENTION_DAYS} -delete

echo "Redis backup completed"
```

### Configuration Backups

```bash
#!/bin/bash
# scripts/backup/config-backup.sh

set -e

BACKUP_DIR="/backups/config"
DATE=$(date +%Y-%m-%d)
BACKUP_FILE="${BACKUP_DIR}/config_${DATE}.tar.gz"

mkdir -p "$BACKUP_DIR"

# Backup configuration files (excluding secrets)
tar -czf "$BACKUP_FILE" \
    --exclude='*.env' \
    --exclude='*secret*' \
    --exclude='*credential*' \
    /etc/linkrift/ \
    /app/config/ \
    /etc/nginx/sites-available/ \
    /etc/systemd/system/linkrift*

echo "Configuration backup completed: $BACKUP_FILE"

# Keep last 30 config backups
find "$BACKUP_DIR" -name "config_*.tar.gz" -mtime +30 -delete
```

---

## Disaster Recovery

### Recovery Plan

```markdown
## Disaster Recovery Runbook

### Priority Order
1. Restore DNS/Load Balancer
2. Restore PostgreSQL (primary data)
3. Restore Redis (caching layer)
4. Restore ClickHouse (analytics)
5. Verify application functionality

### Recovery Time Objectives (RTO)
- Critical services: 1 hour
- Full functionality: 4 hours
- Analytics: 24 hours

### Recovery Point Objectives (RPO)
- PostgreSQL: < 1 hour (continuous backups)
- ClickHouse: < 24 hours
- Redis: < 1 hour (can be rebuilt from PostgreSQL)
```

### PostgreSQL Recovery

```bash
#!/bin/bash
# scripts/recovery/restore-postgres.sh

set -e

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    echo "Available backups:"
    ls -la /backups/postgres/
    exit 1
fi

echo "=== PostgreSQL Recovery ==="
echo "Backup file: $BACKUP_FILE"

# Stop application
echo "Stopping application..."
systemctl stop linkrift

# Drop and recreate database
echo "Recreating database..."
psql -U postgres << EOF
DROP DATABASE IF EXISTS linkrift;
CREATE DATABASE linkrift OWNER linkrift;
EOF

# Restore from backup
echo "Restoring from backup..."
if [[ "$BACKUP_FILE" == *.dump ]]; then
    pg_restore -U linkrift -d linkrift "$BACKUP_FILE"
elif [[ "$BACKUP_FILE" == *.sql.gz ]]; then
    gunzip -c "$BACKUP_FILE" | psql -U linkrift -d linkrift
else
    echo "Unknown backup format"
    exit 1
fi

# Verify restoration
echo "Verifying restoration..."
psql -U linkrift -d linkrift -c "SELECT COUNT(*) FROM links;"
psql -U linkrift -d linkrift -c "SELECT COUNT(*) FROM users;"

# Update statistics
echo "Updating statistics..."
psql -U linkrift -d linkrift -c "ANALYZE VERBOSE;"

# Start application
echo "Starting application..."
systemctl start linkrift

echo "=== Recovery complete ==="
```

### ClickHouse Recovery

```bash
#!/bin/bash
# scripts/recovery/restore-clickhouse.sh

set -e

BACKUP_NAME=$1

if [ -z "$BACKUP_NAME" ]; then
    echo "Usage: $0 <backup_name>"
    echo "Available backups:"
    clickhouse-backup list
    exit 1
fi

echo "=== ClickHouse Recovery ==="

# Download from S3 if needed
if ! clickhouse-backup list local | grep -q "$BACKUP_NAME"; then
    echo "Downloading backup from S3..."
    clickhouse-backup download "$BACKUP_NAME"
fi

# Restore backup
echo "Restoring backup..."
clickhouse-backup restore "$BACKUP_NAME"

# Verify
echo "Verifying restoration..."
clickhouse-client --query "SELECT count() FROM clicks"

echo "=== ClickHouse recovery complete ==="
```

### Full System Recovery

```bash
#!/bin/bash
# scripts/recovery/full-recovery.sh

set -e

echo "=== FULL SYSTEM RECOVERY ==="
echo "Starting at: $(date)"

# 1. Check prerequisites
echo "Checking prerequisites..."
command -v psql >/dev/null || { echo "psql not found"; exit 1; }
command -v redis-cli >/dev/null || { echo "redis-cli not found"; exit 1; }
command -v clickhouse-client >/dev/null || { echo "clickhouse-client not found"; exit 1; }

# 2. Find latest backups
LATEST_PG=$(ls -t /backups/postgres/*.dump 2>/dev/null | head -1)
LATEST_CH=$(ls -t /backups/clickhouse/*.tar.gz 2>/dev/null | head -1)
LATEST_REDIS=$(ls -t /backups/redis/*.rdb.gz 2>/dev/null | head -1)

echo "Latest backups found:"
echo "  PostgreSQL: $LATEST_PG"
echo "  ClickHouse: $LATEST_CH"
echo "  Redis: $LATEST_REDIS"

# 3. Restore PostgreSQL
if [ -n "$LATEST_PG" ]; then
    echo "Restoring PostgreSQL..."
    ./restore-postgres.sh "$LATEST_PG"
fi

# 4. Restore Redis (or rebuild cache)
if [ -n "$LATEST_REDIS" ]; then
    echo "Restoring Redis..."
    systemctl stop redis
    gunzip -c "$LATEST_REDIS" > /var/lib/redis/dump.rdb
    chown redis:redis /var/lib/redis/dump.rdb
    systemctl start redis
else
    echo "No Redis backup found, cache will rebuild"
    redis-cli FLUSHALL
fi

# 5. Restore ClickHouse
if [ -n "$LATEST_CH" ]; then
    BACKUP_NAME=$(basename "$LATEST_CH" .tar.gz)
    echo "Restoring ClickHouse..."
    tar -xzf "$LATEST_CH" -C /var/lib/clickhouse/backup/
    clickhouse-backup restore "$BACKUP_NAME"
fi

# 6. Start application
echo "Starting application..."
systemctl start linkrift

# 7. Verify system health
echo "Verifying system health..."
sleep 10
curl -sf http://localhost:8080/health && echo "Health check passed" || echo "Health check failed"

# 8. Run smoke tests
echo "Running smoke tests..."
./smoke-tests.sh

echo "=== RECOVERY COMPLETE ==="
echo "Finished at: $(date)"
```

---

## Health Checks

```go
// internal/health/checks.go
package health

import (
    "context"
    "database/sql"
    "time"

    "github.com/redis/go-redis/v9"
)

type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

type HealthStatus struct {
    Status    string            `json:"status"`
    Checks    map[string]Check  `json:"checks"`
    Timestamp time.Time         `json:"timestamp"`
}

type Check struct {
    Status  string `json:"status"`
    Latency string `json:"latency,omitempty"`
    Error   string `json:"error,omitempty"`
}

func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Status:    "healthy",
        Checks:    make(map[string]Check),
        Timestamp: time.Now(),
    }

    // Database check
    status.Checks["database"] = h.checkDatabase(ctx)
    if status.Checks["database"].Status != "healthy" {
        status.Status = "unhealthy"
    }

    // Redis check
    status.Checks["redis"] = h.checkRedis(ctx)
    if status.Checks["redis"].Status != "healthy" {
        status.Status = "unhealthy"
    }

    return status
}

func (h *HealthChecker) checkDatabase(ctx context.Context) Check {
    start := time.Now()
    err := h.db.PingContext(ctx)
    latency := time.Since(start)

    if err != nil {
        return Check{Status: "unhealthy", Error: err.Error()}
    }
    return Check{Status: "healthy", Latency: latency.String()}
}

func (h *HealthChecker) checkRedis(ctx context.Context) Check {
    start := time.Now()
    err := h.redis.Ping(ctx).Err()
    latency := time.Since(start)

    if err != nil {
        return Check{Status: "unhealthy", Error: err.Error()}
    }
    return Check{Status: "healthy", Latency: latency.String()}
}
```

---

## Maintenance Scripts

### Crontab Configuration

```cron
# Linkrift Maintenance Crontab
# /etc/cron.d/linkrift-maintenance

# Daily maintenance (3 AM)
0 3 * * * root /opt/linkrift/scripts/maintenance/daily.sh >> /var/log/linkrift/maintenance.log 2>&1

# Weekly maintenance (Sunday 4 AM)
0 4 * * 0 root /opt/linkrift/scripts/maintenance/weekly.sh >> /var/log/linkrift/maintenance.log 2>&1

# Monthly maintenance (1st of month, 5 AM)
0 5 1 * * root /opt/linkrift/scripts/maintenance/monthly.sh >> /var/log/linkrift/maintenance.log 2>&1

# Hourly PostgreSQL backup
0 * * * * root /opt/linkrift/scripts/backup/postgres-backup.sh >> /var/log/linkrift/backup.log 2>&1

# Daily ClickHouse backup (2 AM)
0 2 * * * root /opt/linkrift/scripts/backup/clickhouse-backup.sh >> /var/log/linkrift/backup.log 2>&1

# Hourly Redis backup
30 * * * * root /opt/linkrift/scripts/backup/redis-backup.sh >> /var/log/linkrift/backup.log 2>&1

# Daily config backup (1 AM)
0 1 * * * root /opt/linkrift/scripts/backup/config-backup.sh >> /var/log/linkrift/backup.log 2>&1
```

---

## Related Documentation

- [MONITORING_LOGGING.md](./MONITORING_LOGGING.md) - Monitoring and observability
- [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Issue diagnosis and resolution
- [../security/SECURITY.md](../security/SECURITY.md) - Security procedures
