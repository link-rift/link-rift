-- ============================================================================
-- Bio Pages
-- ============================================================================

-- name: CreateBioPage :one
INSERT INTO bio_pages (workspace_id, slug, title, bio, avatar_url, theme_id, custom_css, meta_title, meta_description, og_image_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetBioPageByID :one
SELECT * FROM bio_pages
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetBioPageBySlug :one
SELECT * FROM bio_pages
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListBioPagesForWorkspace :many
SELECT * FROM bio_pages
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateBioPage :one
UPDATE bio_pages
SET
    slug = COALESCE(sqlc.narg('slug'), slug),
    title = COALESCE(sqlc.narg('title'), title),
    bio = COALESCE(sqlc.narg('bio'), bio),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    theme_id = COALESCE(sqlc.narg('theme_id'), theme_id),
    custom_css = COALESCE(sqlc.narg('custom_css'), custom_css),
    meta_title = COALESCE(sqlc.narg('meta_title'), meta_title),
    meta_description = COALESCE(sqlc.narg('meta_description'), meta_description),
    og_image_url = COALESCE(sqlc.narg('og_image_url'), og_image_url),
    is_published = COALESCE(sqlc.narg('is_published'), is_published),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteBioPage :exec
UPDATE bio_pages
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetBioPageCountForWorkspace :one
SELECT COUNT(*) AS count FROM bio_pages
WHERE workspace_id = $1 AND deleted_at IS NULL;

-- ============================================================================
-- Bio Page Links
-- ============================================================================

-- name: CreateBioPageLink :one
INSERT INTO bio_page_links (bio_page_id, title, url, icon, position, is_visible, visible_from, visible_until)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetBioPageLinkByID :one
SELECT * FROM bio_page_links
WHERE id = $1;

-- name: ListBioPageLinks :many
SELECT * FROM bio_page_links
WHERE bio_page_id = $1
ORDER BY position ASC;

-- name: UpdateBioPageLink :one
UPDATE bio_page_links
SET
    title = COALESCE(sqlc.narg('title'), title),
    url = COALESCE(sqlc.narg('url'), url),
    icon = COALESCE(sqlc.narg('icon'), icon),
    is_visible = COALESCE(sqlc.narg('is_visible'), is_visible),
    visible_from = COALESCE(sqlc.narg('visible_from'), visible_from),
    visible_until = COALESCE(sqlc.narg('visible_until'), visible_until),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBioPageLink :exec
DELETE FROM bio_page_links
WHERE id = $1;

-- name: UpdateBioPageLinkPosition :exec
UPDATE bio_page_links
SET position = $2, updated_at = NOW()
WHERE id = $1;

-- name: IncrementBioPageLinkClickCount :exec
UPDATE bio_page_links
SET click_count = click_count + 1
WHERE id = $1;

-- name: GetMaxBioPageLinkPosition :one
SELECT COALESCE(MAX(position), -1)::integer AS max_position FROM bio_page_links
WHERE bio_page_id = $1;
