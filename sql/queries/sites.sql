-- name: CreateSite :one
WITH upsert AS (
    INSERT INTO SITES (
        code,
        name,
        address,
        is_cms
    )
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (code, name) DO UPDATE
    SET
        address = EXCLUDED.address,
        is_cms = EXCLUDED.is_cms
    WHERE
        sites.address IS DISTINCT FROM EXCLUDED.address
        OR sites.is_cms IS DISTINCT FROM EXCLUDED.is_cms
    RETURNING *
)
SELECT * FROM upsert
UNION ALL
SELECT * FROM sites
WHERE code = $1
    AND name = $2
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetSiteById :one
SELECT *
FROM sites
WHERE id = $1;

-- name: GetSiteByCode :one
SELECT *
FROM sites
WHERE code = $1;