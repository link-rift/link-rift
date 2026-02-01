-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    workspace_id, user_id, action, resource_type, resource_id,
    old_values, new_values, metadata, ip_address, user_agent
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: ListAuditLogsForWorkspace :many
SELECT * FROM audit_logs
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAuditLogsForWorkspace :one
SELECT COUNT(*) FROM audit_logs
WHERE workspace_id = $1;

-- name: GetAuditLogByID :one
SELECT * FROM audit_logs
WHERE id = $1 AND workspace_id = $2;

-- name: ListAuditLogsFiltered :many
SELECT * FROM audit_logs
WHERE workspace_id = $1
  AND (sqlc.narg('action')::VARCHAR IS NULL OR action = sqlc.narg('action')::VARCHAR)
  AND (sqlc.narg('resource_type')::VARCHAR IS NULL OR resource_type = sqlc.narg('resource_type')::VARCHAR)
  AND (sqlc.narg('user_id')::UUID IS NULL OR user_id = sqlc.narg('user_id')::UUID)
  AND (sqlc.narg('start_date')::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg('start_date')::TIMESTAMPTZ)
  AND (sqlc.narg('end_date')::TIMESTAMPTZ IS NULL OR created_at <= sqlc.narg('end_date')::TIMESTAMPTZ)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAuditLogsFiltered :one
SELECT COUNT(*) FROM audit_logs
WHERE workspace_id = $1
  AND (sqlc.narg('action')::VARCHAR IS NULL OR action = sqlc.narg('action')::VARCHAR)
  AND (sqlc.narg('resource_type')::VARCHAR IS NULL OR resource_type = sqlc.narg('resource_type')::VARCHAR)
  AND (sqlc.narg('user_id')::UUID IS NULL OR user_id = sqlc.narg('user_id')::UUID)
  AND (sqlc.narg('start_date')::TIMESTAMPTZ IS NULL OR created_at >= sqlc.narg('start_date')::TIMESTAMPTZ)
  AND (sqlc.narg('end_date')::TIMESTAMPTZ IS NULL OR created_at <= sqlc.narg('end_date')::TIMESTAMPTZ);
