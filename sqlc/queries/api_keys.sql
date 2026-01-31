-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, workspace_id, name, key_hash, key_prefix, scopes, rate_limit, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAPIKeyByPrefix :one
SELECT * FROM api_keys
WHERE key_prefix = $1 AND revoked_at IS NULL;

-- name: ListAPIKeysForWorkspace :many
SELECT * FROM api_keys
WHERE workspace_id = $1 AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: RevokeAPIKey :exec
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;
