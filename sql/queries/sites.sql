-- name: CreateSite :one
WITH upsert AS (
    INSERT INTO SITES (
        code, -- $1
        name, -- $2
        address, -- $3
        is_cms -- $4
    )
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (code) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
        name = EXCLUDED.name,
        address = EXCLUDED.address,
        is_cms = EXCLUDED.is_cms
    WHERE
        sites.name IS DISTINCT FROM EXCLUDED.name
        OR sites.address IS DISTINCT FROM EXCLUDED.address
        OR sites.is_cms IS DISTINCT FROM EXCLUDED.is_cms
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
