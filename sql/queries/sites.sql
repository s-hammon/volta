-- name: CreateSite :one
INSERT INTO sites (code, name, address, is_cms)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code, name) DO UPDATE
SET name = EXCLUDED.name,
    address = EXCLUDED.address,
    is_cms = EXCLUDED.is_cms
WHERE sites.name IS DISTINCT FROM EXCLUDED.name
    OR sites.address IS DISTINCT FROM EXCLUDED.address
    OR sites.is_cms IS DISTINCT FROM EXCLUDED.is_cms
RETURNING *;

-- name: GetSiteByCode :one
SELECT *
FROM sites
WHERE code = $1;