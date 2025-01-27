-- name: CreateSite :one
INSERT INTO sites (code, name, address, is_cms)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSiteByCode :one
SELECT *
FROM sites
WHERE code = $1;