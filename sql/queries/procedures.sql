-- name: CreateProcedure :one
INSERT INTO procedures (site_id, code, description, specialty, modality)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProcedureBySiteIDCode :one
SELECT *
FROM procedures
WHERE
    site_id = $1
    AND code = $2;