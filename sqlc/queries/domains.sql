-- name: CreateDomain :one
INSERT INTO domains (workspace_id, domain)
VALUES ($1, $2)
RETURNING *;

-- name: GetDomainByID :one
SELECT * FROM domains
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDomainByDomain :one
SELECT * FROM domains
WHERE domain = $1 AND deleted_at IS NULL;

-- name: ListDomainsForWorkspace :many
SELECT * FROM domains
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateDomain :one
UPDATE domains
SET
    is_verified = COALESCE(sqlc.narg('is_verified'), is_verified),
    verified_at = COALESCE(sqlc.narg('verified_at'), verified_at),
    ssl_status = COALESCE(sqlc.narg('ssl_status'), ssl_status),
    ssl_expires_at = COALESCE(sqlc.narg('ssl_expires_at'), ssl_expires_at),
    dns_records = COALESCE(sqlc.narg('dns_records'), dns_records),
    last_dns_check_at = COALESCE(sqlc.narg('last_dns_check_at'), last_dns_check_at),
    default_redirect_url = COALESCE(sqlc.narg('default_redirect_url'), default_redirect_url),
    custom_404_url = COALESCE(sqlc.narg('custom_404_url'), custom_404_url),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteDomain :exec
UPDATE domains
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
