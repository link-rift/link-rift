-- name: CreateWebhook :one
INSERT INTO webhooks (workspace_id, url, secret, events, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWebhookByID :one
SELECT * FROM webhooks
WHERE id = $1;

-- name: ListWebhooksForWorkspace :many
SELECT * FROM webhooks
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: UpdateWebhook :one
UPDATE webhooks
SET url = COALESCE($2, url),
    events = COALESCE($3, events),
    is_active = COALESCE($4, is_active),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE id = $1;

-- name: GetActiveWebhooksForEvent :many
SELECT * FROM webhooks
WHERE workspace_id = $1
  AND is_active = TRUE
  AND $2::text = ANY(events);

-- name: IncrementWebhookFailureCount :exec
UPDATE webhooks
SET failure_count = failure_count + 1
WHERE id = $1;

-- name: ResetWebhookFailureCount :exec
UPDATE webhooks
SET failure_count = 0
WHERE id = $1;

-- name: UpdateWebhookLastTriggered :exec
UPDATE webhooks
SET last_triggered_at = NOW(), last_success_at = NOW(), failure_count = 0
WHERE id = $1;

-- name: DisableWebhook :exec
UPDATE webhooks
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: CreateWebhookDelivery :one
INSERT INTO webhook_deliveries (webhook_id, event, payload, max_attempts)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhookDeliveryByID :one
SELECT * FROM webhook_deliveries
WHERE id = $1;

-- name: ListWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE webhook_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWebhookDeliveries :one
SELECT COUNT(*) FROM webhook_deliveries
WHERE webhook_id = $1;

-- name: UpdateWebhookDelivery :exec
UPDATE webhook_deliveries
SET response_status = $2,
    response_body = $3,
    attempts = $4,
    last_attempt_at = NOW(),
    completed_at = $5
WHERE id = $1;

-- name: GetPendingWebhookDeliveries :many
SELECT * FROM webhook_deliveries
WHERE completed_at IS NULL
  AND attempts < max_attempts
  AND (last_attempt_at IS NULL OR last_attempt_at < NOW() - INTERVAL '30 seconds')
ORDER BY created_at ASC
LIMIT 50;

-- name: CountRecentWebhookFailures :one
SELECT COUNT(*) FROM webhook_deliveries
WHERE webhook_id = $1
  AND created_at > NOW() - INTERVAL '24 hours'
  AND completed_at IS NOT NULL
  AND (response_status IS NULL OR response_status >= 400);
