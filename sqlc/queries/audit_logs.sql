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
