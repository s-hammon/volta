-- name: CreateVisit :one
WITH upsert AS (
    INSERT INTO visits (
        outside_system_id,
        site_id,
        mrn_id,
        number,
        patient_type,
        message_id
    )
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
    )
    ON CONFLICT (site_id, mrn_id, number) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
        outside_system_id = EXCLUDED.outside_system_id,
        patient_type = EXCLUDED.patient_type
    WHERE
        visits.outside_system_id IS DISTINCT FROM EXCLUDED.outside_system_id
        OR visits.patient_type IS DISTINCT FROM EXCLUDED.patient_type
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM visits
WHERE
    site_id = $2
    AND mrn_id = $3
    AND number = $4
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetVisitById :one
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
WHERE v.id = $1;

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
