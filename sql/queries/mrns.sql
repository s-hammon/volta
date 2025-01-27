-- name: CreateMrn :one
INSERT INTO mrns (site_id, patient_id, mrn)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMrnBySitePatient :one
SELECT *
FROM mrns
WHERE
    site_id = $1
    AND patient_id = $2;