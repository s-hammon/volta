-- name: CreateMrn :one
INSERT INTO mrns (site_id, patient_id, mrn)
VALUES ($1, $2, $3)
ON CONFLICT (site_id, patient_id) DO UPDATE
SET mrn = EXCLUDED.mrn
WHERE mrns.mrn IS DISTINCT FROM EXCLUDED.mrn
RETURNING *;

-- name: GetMrnBySitePatient :one
SELECT *
FROM mrns
WHERE
    site_id = $1
    AND patient_id = $2;