-- name: AddWorkspaceMember :one
INSERT INTO workspace_members (workspace_id, user_id, role, invited_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkspaceMember :one
SELECT * FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: ListWorkspaceMembers :many
SELECT wm.*, u.email, u.name AS user_name, u.avatar_url
FROM workspace_members wm
JOIN users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.joined_at;

-- name: UpdateMemberRole :one
UPDATE workspace_members
SET role = $3
WHERE workspace_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: GetMemberCountForWorkspace :one
SELECT COUNT(*) FROM workspace_members WHERE workspace_id = $1;
