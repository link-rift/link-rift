#!/usr/bin/env bash
# Linkrift Backup Verification Script
# Verifies that backup and restore procedures work correctly.
#
# Usage: ./scripts/verify-backup.sh [--restore-test]
#
# Without --restore-test: Checks that backups exist and are recent
# With --restore-test: Actually restores to a test database (destructive to test DB)

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

RESTORE_TEST=false
if [[ "${1:-}" == "--restore-test" ]]; then
    RESTORE_TEST=true
fi

BACKUP_DIR="${BACKUP_DIR:-/backups}"
DB_URL="${DATABASE_URL:-postgres://linkrift:linkrift_dev@localhost:5432/linkrift}"
TEST_DB_URL="${TEST_DATABASE_URL:-postgres://linkrift:linkrift_dev@localhost:5432/linkrift_backup_test}"

echo "============================================"
echo " Linkrift Backup Verification"
echo " $(date)"
echo "============================================"
echo

# ── 1. Check Backup Directory ────────────────
echo "1. Backup Directory"
echo "---"

if [ -d "$BACKUP_DIR" ]; then
    echo "  Backup directory: $BACKUP_DIR"

    # Count backup files
    PG_BACKUPS=$(find "$BACKUP_DIR" -name "linkrift_*.dump" -o -name "linkrift_*.sql.gz" 2>/dev/null | wc -l || echo 0)
    echo "  PostgreSQL backups found: $PG_BACKUPS"

    if [ "$PG_BACKUPS" -gt 0 ]; then
        # Check most recent backup
        LATEST=$(find "$BACKUP_DIR" -name "linkrift_*" -type f -printf '%T@ %p\n' 2>/dev/null | sort -n | tail -1 | cut -d' ' -f2 || echo "")
        if [ -n "$LATEST" ]; then
            LATEST_SIZE=$(du -sh "$LATEST" | cut -f1)
            LATEST_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST" 2>/dev/null || stat -f %m "$LATEST" 2>/dev/null)) / 3600 ))
            echo -e "  ${GREEN}Latest backup${NC}: $LATEST ($LATEST_SIZE, ${LATEST_AGE}h ago)"

            if [ "$LATEST_AGE" -gt 24 ]; then
                echo -e "  ${RED}WARNING${NC}: Latest backup is more than 24 hours old!"
            fi
        fi
    else
        echo -e "  ${YELLOW}No backups found${NC}"
    fi
else
    echo -e "  ${YELLOW}Backup directory $BACKUP_DIR does not exist${NC}"
    echo "  Set BACKUP_DIR environment variable to your backup location"
fi
echo

# ── 2. Create Test Backup ─────────────────────
echo "2. Test Backup Creation"
echo "---"

TEST_BACKUP="/tmp/linkrift_verify_$(date +%Y%m%d_%H%M%S).dump"
echo "  Creating test backup to: $TEST_BACKUP"

if pg_dump -Fc "$DB_URL" -f "$TEST_BACKUP" 2>/dev/null; then
    BACKUP_SIZE=$(du -sh "$TEST_BACKUP" | cut -f1)
    echo -e "  ${GREEN}PASS${NC} Backup created successfully ($BACKUP_SIZE)"
else
    echo -e "  ${RED}FAIL${NC} pg_dump failed — check database connectivity"
    echo "  Ensure DATABASE_URL is set correctly"
    exit 1
fi
echo

# ── 3. Verify Backup Integrity ────────────────
echo "3. Backup Integrity Check"
echo "---"

# List backup contents to verify it's not corrupt
TABLE_COUNT=$(pg_restore -l "$TEST_BACKUP" 2>/dev/null | grep "TABLE" | wc -l || echo 0)
INDEX_COUNT=$(pg_restore -l "$TEST_BACKUP" 2>/dev/null | grep "INDEX" | wc -l || echo 0)
echo "  Tables in backup: $TABLE_COUNT"
echo "  Indexes in backup: $INDEX_COUNT"

if [ "$TABLE_COUNT" -gt 0 ]; then
    echo -e "  ${GREEN}PASS${NC} Backup contains valid data"
else
    echo -e "  ${RED}FAIL${NC} Backup appears empty or corrupt"
fi
echo

# ── 4. Restore Test (optional) ────────────────
if [ "$RESTORE_TEST" = true ]; then
    echo "4. Restore Test"
    echo "---"

    echo "  Restoring to test database..."
    echo "  WARNING: This will drop and recreate the test database"

    # Drop and recreate test database
    psql "$DB_URL" -c "DROP DATABASE IF EXISTS linkrift_backup_test;" 2>/dev/null || true
    psql "$DB_URL" -c "CREATE DATABASE linkrift_backup_test;" 2>/dev/null

    if pg_restore -d "$TEST_DB_URL" "$TEST_BACKUP" 2>/dev/null; then
        echo -e "  ${GREEN}PASS${NC} Restore completed successfully"

        # Verify data in restored database
        RESTORED_TABLES=$(psql "$TEST_DB_URL" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
        echo "  Tables in restored database: $RESTORED_TABLES"

        if [ "$RESTORED_TABLES" -gt 0 ]; then
            echo -e "  ${GREEN}PASS${NC} Restored database has valid tables"
        else
            echo -e "  ${RED}FAIL${NC} Restored database appears empty"
        fi

        # Cleanup test database
        psql "$DB_URL" -c "DROP DATABASE IF EXISTS linkrift_backup_test;" 2>/dev/null || true
    else
        echo -e "  ${RED}FAIL${NC} Restore failed"
    fi
    echo
else
    echo "4. Restore Test"
    echo "---"
    echo "  Skipped (run with --restore-test to perform restore verification)"
    echo
fi

# ── 5. Cleanup ────────────────────────────────
rm -f "$TEST_BACKUP"

echo "============================================"
echo " Backup verification complete"
echo "============================================"
