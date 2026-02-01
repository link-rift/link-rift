-- name: GetSSOConfigByWorkspace :one
SELECT * FROM sso_configs WHERE workspace_id = $1;

-- name: CreateSSOConfig :one
INSERT INTO sso_configs (
    workspace_id, provider, entity_id, sso_url, slo_url,
    certificate, idp_metadata_url, idp_metadata_xml,
    attribute_mapping, is_enabled, enforce_sso, allowed_domains
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateSSOConfig :one
UPDATE sso_configs SET
    provider = $2,
    entity_id = $3,
    sso_url = $4,
    slo_url = $5,
    certificate = $6,
    idp_metadata_url = $7,
    idp_metadata_xml = $8,
    attribute_mapping = $9,
    is_enabled = $10,
    enforce_sso = $11,
    allowed_domains = $12,
    updated_at = NOW()
WHERE workspace_id = $1
RETURNING *;

-- name: DeleteSSOConfig :exec
DELETE FROM sso_configs WHERE workspace_id = $1;

-- name: GetSSOIdentityByExternalID :one
SELECT * FROM sso_identities WHERE workspace_id = $1 AND external_id = $2;

-- name: GetSSOIdentityByUserID :one
SELECT * FROM sso_identities WHERE user_id = $1 AND workspace_id = $2;

-- name: CreateSSOIdentity :one
INSERT INTO sso_identities (
    user_id, workspace_id, provider, external_id, email, name, raw_attributes
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateSSOIdentityLastLogin :exec
UPDATE sso_identities SET
    last_login_at = NOW(),
    email = $3,
    name = $4,
    raw_attributes = $5,
    updated_at = NOW()
WHERE user_id = $1 AND workspace_id = $2;

-- name: ListSSOIdentitiesForWorkspace :many
SELECT * FROM sso_identities WHERE workspace_id = $1 ORDER BY created_at DESC;

-- name: DeleteSSOIdentity :exec
DELETE FROM sso_identities WHERE id = $1;
