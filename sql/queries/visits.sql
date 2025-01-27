-- name: CreateVisit :one
INSERT INTO visits (
    outside_system_id,
    site_id,
    mrn_id,
    number,
    patient_type
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetVisitBySiteIdNumber :one
SELECT *
FROM visits
WHERE
    site_id = $1
    AND number = $2;