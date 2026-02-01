CREATE TABLE scim_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    token_prefix VARCHAR(12) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_scim_tokens_workspace ON scim_tokens(workspace_id);
CREATE INDEX idx_scim_tokens_hash ON scim_tokens(token_hash) WHERE is_active = TRUE;

CREATE TABLE scim_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    external_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scim_sync_log_workspace ON scim_sync_log(workspace_id, created_at DESC);
