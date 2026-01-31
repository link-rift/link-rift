-- name: GetActiveRulesForLink :many
SELECT * FROM link_rules
WHERE link_id = $1 AND is_active = TRUE
ORDER BY priority ASC;

-- name: CreateLinkRule :one
INSERT INTO link_rules (
    link_id, rule_type, priority, is_active, conditions, destination_url, weight
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateLinkRule :one
UPDATE link_rules
SET
    rule_type = COALESCE(sqlc.narg('rule_type'), rule_type),
    priority = COALESCE(sqlc.narg('priority'), priority),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    conditions = COALESCE(sqlc.narg('conditions'), conditions),
    destination_url = COALESCE(sqlc.narg('destination_url'), destination_url),
    weight = COALESCE(sqlc.narg('weight'), weight),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLinkRule :exec
DELETE FROM link_rules WHERE id = $1;

-- name: GetLinkRuleByID :one
SELECT * FROM link_rules WHERE id = $1;
