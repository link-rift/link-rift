CREATE TABLE workspace_branding (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    logo_url VARCHAR(500),
    logo_dark_url VARCHAR(500),
    favicon_url VARCHAR(500),
    primary_color VARCHAR(7),
    secondary_color VARCHAR(7),
    accent_color VARCHAR(7),
    custom_css TEXT,
    hide_powered_by BOOLEAN NOT NULL DEFAULT FALSE,
    custom_footer_text VARCHAR(255),
    custom_footer_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_workspace_branding_workspace UNIQUE (workspace_id)
);

CREATE INDEX idx_workspace_branding_workspace ON workspace_branding(workspace_id);
