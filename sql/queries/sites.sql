-- name: CreateSite :one
WITH upsert AS (
    INSERT INTO sites (
        code, -- $1
        name, -- $2
        address, -- $3
        is_cms, -- $4
        message_id -- $5
    )
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (code) DO NOTHING
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM sites
WHERE code = $1
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetSiteById :one
SELECT *
FROM sites
WHERE id = $1;

-- name: GetSiteByCode :one
SELECT *
FROM sites
WHERE code = $1;
