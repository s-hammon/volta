-- name: CreateProcedure :one
WITH upsert AS (
    INSERT INTO procedures (site_id, code, description, specialty, modality, message_id)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (site_id, code) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
        description = COALESCE(NULLIF(EXCLUDED.description, ''), procedures.description)
    WHERE
        COALESCE(NULLIF(EXCLUDED.description, ''), procedures.description) IS DISTINCT FROM EXCLUDED.description
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM procedures
WHERE
    site_id = $1
    AND code = $2
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetProcedureById :one
SELECT *
FROM procedures
WHERE id = $1;

-- name: GetProcedureBySiteIDCode :one
SELECT *
FROM procedures
WHERE
    site_id = $1
    AND code = $2;

-- name: GetProceduresForSpecialtyUpdate :many
SELECT *
FROM procedures
WHERE
    specialty is null
    AND id > $1
ORDER BY id -- so one can move cursor value $1 to max(id)
LIMIT 100;

-- name: UpdateProcedureSpecialty :exec
UPDATE procedures
SET specialty = $2
WHERE id = $1;
