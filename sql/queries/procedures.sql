-- name: CreateProcedure :one
WITH upsert AS (
    INSERT INTO procedures (site_id, code, description, specialty, modality, message_id)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (site_id, code) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
        description = COALESCE(NULLIF(EXCLUDED.description, ''), procedures.description),
        specialty = COALESCE(NULLIF(EXCLUDED.specialty, ''), procedures.specialty),
        modality = COALESCE(NULLIF(EXCLUDED.specialty, ''), procedures.specialty)
    WHERE
        COALESCE(NULLIF(EXCLUDED.description, ''), procedures.description) IS DISTINCT FROM EXCLUDED.description
        OR COALESCE(NULLIF(EXCLUDED.specialty, ''), procedures.specialty) IS DISTINCT FROM EXCLUDED.specialty
        OR COALESCE(NULLIF(EXCLUDED.specialty, ''), procedures.specialty) IS DISTINCT FROM EXCLUDED.modality
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

-- name: GetProceduresForSpecialtyAssignment :many
SELECT *
FROM procedures
WHERE
    id > $1
    AND specialty IS NULL
ORDER BY id
LIMIT 100;

-- name: UpdateProcedureSpecialty :exec
UPDATE procedures
SET
    updated_at = CURRENT_TIMESTAMP,
    specialty = $2
WHERE id = $1;
