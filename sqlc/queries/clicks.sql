-- name: InsertClick :exec
INSERT INTO clicks (
    link_id, clicked_at, visitor_id, ip_address, user_agent, referer,
    country_code, region, city, device_type, browser, browser_version,
    os, os_version, is_bot, utm_source, utm_medium, utm_campaign
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: GetClicksByLinkID :many
SELECT * FROM clicks
WHERE link_id = $1
    AND clicked_at >= $2
    AND clicked_at <= $3
ORDER BY clicked_at DESC
LIMIT $4 OFFSET $5;
