CREATE TABLE sso_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL DEFAULT 'saml',
    entity_id VARCHAR(500) NOT NULL,
    sso_url VARCHAR(500) NOT NULL,
    slo_url VARCHAR(500),
    certificate TEXT NOT NULL,
    idp_metadata_url VARCHAR(500),
    idp_metadata_xml TEXT,
    attribute_mapping JSONB NOT NULL DEFAULT '{}',
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    enforce_sso BOOLEAN NOT NULL DEFAULT FALSE,
    allowed_domains TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sso_configs_workspace UNIQUE (workspace_id)
);

CREATE INDEX idx_sso_configs_workspace ON sso_configs(workspace_id);

CREATE TABLE sso_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL DEFAULT 'saml',
    external_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    raw_attributes JSONB NOT NULL DEFAULT '{}',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sso_identities_workspace_external UNIQUE (workspace_id, external_id)
);

CREATE INDEX idx_sso_identities_user ON sso_identities(user_id);
CREATE INDEX idx_sso_identities_workspace ON sso_identities(workspace_id);
CREATE INDEX idx_sso_identities_external ON sso_identities(workspace_id, external_id);
