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
SELECT
    v.*,
    s.created_at as site_created_at,
    s.updated_at as site_updated_at,
    s.code as site_code,
    s.name as site_name,
    s.address as site_address,
    s.is_cms as site_is_cms,
    m.created_at as mrn_created_at,
    m.updated_at as mrn_updated_at,
    m.mrn as mrn_value
FROM visits as v
LEFT JOIN sites as s ON v.site_id = s.id
LEFT JOIN mrns as m ON v.mrn_id = m.id
WHERE
    v.site_id = $1
    AND v.number = $2;

-- name: UpdateVisit :one
UPDATE visits
SET
    updated_at = CURRENT_TIMESTAMP,
    outside_system_id = $2,
    site_id = $3,
    mrn_id = $4,
    number = $5,
    patient_type = $6
WHERE id = $1
RETURNING *;