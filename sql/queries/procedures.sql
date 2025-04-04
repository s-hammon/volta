-- name: CreateProcedure :one
INSERT INTO procedures (site_id, code, description, specialty, modality)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (site_id, code) DO UPDATE
SET description = EXCLUDED.description,
    specialty = EXCLUDED.specialty,
    modality = EXCLUDED.modality
WHERE procedures.description IS DISTINCT FROM EXCLUDED.description
    OR procedures.specialty IS DISTINCT FROM EXCLUDED.specialty
    OR procedures.modality IS DISTINCT FROM EXCLUDED.modality
RETURNING *;

-- name: GetProcedureBySiteIDCode :one
SELECT *
FROM procedures
WHERE
    site_id = $1
    AND code = $2;