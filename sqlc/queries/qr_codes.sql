-- name: CreateQRCode :one
INSERT INTO qr_codes (
    link_id,
    qr_type,
    error_correction,
    foreground_color,
    background_color,
    logo_url,
    png_url,
    svg_url,
    dot_style,
    corner_style,
    size,
    margin
) VALUES (
    $1, $2, $3, $4, $5,
    sqlc.narg('logo_url'),
    sqlc.narg('png_url'),
    sqlc.narg('svg_url'),
    $6, $7, $8, $9
)
RETURNING *;

-- name: GetQRCodeByID :one
SELECT * FROM qr_codes
WHERE id = $1;

-- name: GetQRCodeByLinkID :one
SELECT * FROM qr_codes
WHERE link_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListQRCodesForLink :many
SELECT * FROM qr_codes
WHERE link_id = $1
ORDER BY created_at DESC;

-- name: UpdateQRCode :one
UPDATE qr_codes SET
    qr_type = COALESCE(sqlc.narg('qr_type'), qr_type),
    error_correction = COALESCE(sqlc.narg('error_correction'), error_correction),
    foreground_color = COALESCE(sqlc.narg('foreground_color'), foreground_color),
    background_color = COALESCE(sqlc.narg('background_color'), background_color),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    png_url = COALESCE(sqlc.narg('png_url'), png_url),
    svg_url = COALESCE(sqlc.narg('svg_url'), svg_url),
    dot_style = COALESCE(sqlc.narg('dot_style'), dot_style),
    corner_style = COALESCE(sqlc.narg('corner_style'), corner_style),
    size = COALESCE(sqlc.narg('size'), size),
    margin = COALESCE(sqlc.narg('margin'), margin),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteQRCode :exec
DELETE FROM qr_codes
WHERE id = $1;

-- name: IncrementQRScanCount :exec
UPDATE qr_codes
SET scan_count = scan_count + 1
WHERE id = $1;
