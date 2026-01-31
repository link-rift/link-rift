-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token_hash, ip_address, user_agent, device_name, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE refresh_token_hash = $1
    AND is_revoked = FALSE
    AND expires_at > NOW();

-- name: RevokeSession :exec
UPDATE sessions
SET is_revoked = TRUE
WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW() OR is_revoked = TRUE;

-- name: RevokeAllUserSessions :exec
UPDATE sessions
SET is_revoked = TRUE
WHERE user_id = $1 AND is_revoked = FALSE;

-- name: ListUserSessions :many
SELECT * FROM sessions
WHERE user_id = $1 AND is_revoked = FALSE AND expires_at > NOW()
ORDER BY last_active_at DESC;
