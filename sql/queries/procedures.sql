-- name: CreateProcedure :one
INSERT INTO procedures (code, description, specialty, modality)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProcedureBySiteIDCode :one
SELECT *
FROM procedures
WHERE
    site_id = $1
    AND code = $2;