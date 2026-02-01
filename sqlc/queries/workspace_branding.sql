-- name: GetWorkspaceBranding :one
SELECT * FROM workspace_branding WHERE workspace_id = $1;

-- name: UpsertWorkspaceBranding :one
INSERT INTO workspace_branding (
    workspace_id,
    logo_url,
    logo_dark_url,
    favicon_url,
    primary_color,
    secondary_color,
    accent_color,
    custom_css,
    hide_powered_by,
    custom_footer_text,
    custom_footer_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (workspace_id) DO UPDATE SET
    logo_url = EXCLUDED.logo_url,
    logo_dark_url = EXCLUDED.logo_dark_url,
    favicon_url = EXCLUDED.favicon_url,
    primary_color = EXCLUDED.primary_color,
    secondary_color = EXCLUDED.secondary_color,
    accent_color = EXCLUDED.accent_color,
    custom_css = EXCLUDED.custom_css,
    hide_powered_by = EXCLUDED.hide_powered_by,
    custom_footer_text = EXCLUDED.custom_footer_text,
    custom_footer_url = EXCLUDED.custom_footer_url,
    updated_at = NOW()
RETURNING *;

-- name: DeleteWorkspaceBranding :exec
DELETE FROM workspace_branding WHERE workspace_id = $1;
