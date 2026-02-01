-- name: CreateSCIMToken :one
INSERT INTO scim_tokens (
    workspace_id, token_hash, token_prefix, name, expires_at
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSCIMTokenByHash :one
SELECT * FROM scim_tokens
WHERE token_hash = $1 AND is_active = TRUE;

-- name: ListSCIMTokensForWorkspace :many
SELECT * FROM scim_tokens
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: RevokeSCIMToken :exec
UPDATE scim_tokens
SET is_active = FALSE
WHERE id = $1;

-- name: UpdateSCIMTokenLastUsed :exec
UPDATE scim_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: DeleteSCIMToken :exec
DELETE FROM scim_tokens WHERE id = $1;

-- name: CreateSCIMSyncLog :exec
INSERT INTO scim_sync_log (
    workspace_id, operation, resource_type, external_id, status, details
)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListSCIMSyncLogs :many
SELECT * FROM scim_sync_log
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSCIMSyncLogs :one
SELECT COUNT(*) FROM scim_sync_log
WHERE workspace_id = $1;
