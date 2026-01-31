-- name: CreateWorkspace :one
INSERT INTO workspaces (name, slug, owner_id, plan, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkspaceByID :one
SELECT * FROM workspaces
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkspaceBySlug :one
SELECT * FROM workspaces
WHERE slug = $1 AND deleted_at IS NULL;

-- name: ListWorkspacesForUser :many
SELECT w.* FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1 AND w.deleted_at IS NULL
ORDER BY w.created_at DESC;

-- name: UpdateWorkspace :one
UPDATE workspaces
SET
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    plan = COALESCE(sqlc.narg('plan'), plan),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkspace :exec
UPDATE workspaces
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkspaceCountForUser :one
SELECT COUNT(*) FROM workspaces w
JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1 AND wm.role = 'owner' AND w.deleted_at IS NULL;

-- name: UpdateWorkspaceOwner :one
UPDATE workspaces
SET owner_id = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;
