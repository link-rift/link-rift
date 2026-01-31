#!/usr/bin/env bash
set -euo pipefail

# Migration helper script for Linkrift
# Usage: ./scripts/migrate.sh <command> [args]
#
# Commands:
#   up        Run all pending migrations
#   down      Rollback last migration
#   status    Show migration status
#   force N   Force set migration version to N

MIGRATE_PATH="migrations/postgres"
DATABASE_URL="${DATABASE_URL:-postgres://linkrift:linkrift_dev@localhost:5432/linkrift?sslmode=disable}"

command="${1:-help}"
shift || true

case "$command" in
    up)
        echo "Running pending migrations..."
        migrate -path "$MIGRATE_PATH" -database "$DATABASE_URL" up "$@"
        echo "Migrations applied."
        ;;
    down)
        echo "Rolling back last migration..."
        migrate -path "$MIGRATE_PATH" -database "$DATABASE_URL" down 1
        echo "Rollback complete."
        ;;
    status)
        echo "Migration status:"
        migrate -path "$MIGRATE_PATH" -database "$DATABASE_URL" version
        ;;
    force)
        if [ -z "${1:-}" ]; then
            echo "Usage: $0 force <version>"
            exit 1
        fi
        echo "Forcing migration version to $1..."
        migrate -path "$MIGRATE_PATH" -database "$DATABASE_URL" force "$1"
        echo "Version forced to $1."
        ;;
    help|*)
        echo "Linkrift Migration Helper"
        echo ""
        echo "Usage: $0 <command> [args]"
        echo ""
        echo "Commands:"
        echo "  up        Run all pending migrations"
        echo "  down      Rollback last migration"
        echo "  status    Show current migration version"
        echo "  force N   Force set migration version to N"
        echo ""
        echo "Environment:"
        echo "  DATABASE_URL  PostgreSQL connection string"
        ;;
esac
