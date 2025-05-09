-- name: CreateMrn :one
WITH upsert AS (
    INSERT INTO mrns (site_id, patient_id, mrn)
    VALUES ($1, $2, $3)
    ON CONFLICT (site_id, patient_id) DO UPDATE
    SET mrn = COALESCE(NULLIF(EXCLUDED.mrn, ''), mrns.mrn)
    WHERE mrns.mrn IS DISTINCT FROM COALESCE(NULLIF(EXCLUDED.mrn, ''), mrns.mrn)
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM mrns
WHERE
    site_id = $1
    AND patient_id = $2
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetMrnById :one
SELECT *
FROM mrns
WHERE id = $1;

-- name: GetMrnBySitePatient :one
SELECT *
FROM mrns
WHERE
    site_id = $1
    AND patient_id = $2;
