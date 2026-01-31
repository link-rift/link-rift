-- name: CreateLink :one
INSERT INTO links (
    user_id, workspace_id, domain_id, url, short_code,
    title, description, is_active, password_hash,
    expires_at, max_clicks,
    utm_source, utm_medium, utm_campaign, utm_term, utm_content
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetLinkByID :one
SELECT * FROM links
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetLinkByShortCode :one
SELECT * FROM links
WHERE short_code = $1 AND deleted_at IS NULL;

-- name: ListLinksForWorkspace :many
SELECT
    l.*,
    COUNT(*) OVER() AS total_count
FROM links l
WHERE l.workspace_id = $1
    AND l.deleted_at IS NULL
    AND (sqlc.narg('search')::text IS NULL OR
         to_tsvector('english', COALESCE(l.title, '') || ' ' || COALESCE(l.description, '')) @@
         plainto_tsquery('english', sqlc.narg('search')::text))
ORDER BY l.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateLink :one
UPDATE links
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    url = COALESCE(sqlc.narg('url'), url),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    max_clicks = COALESCE(sqlc.narg('max_clicks'), max_clicks),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteLink :exec
UPDATE links
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: IncrementLinkClicks :exec
UPDATE links
SET total_clicks = total_clicks + 1, updated_at = NOW()
WHERE id = $1;
