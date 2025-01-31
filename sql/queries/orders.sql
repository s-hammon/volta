-- name: CreateOrder :one
INSERT INTO orders (
    outside_system_id,
    site_id,
    visit_id,
    mrn_id,
    ordering_physician_id,
    arrival,
    number,
    current_status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetOrderBySiteIDNumber :one
SELECT
    o.*,
    s.created_at as site_created_at,
    s.updated_at as site_updated_at,
    s.code as site_code,
    s.name as site_name,
    s.address as site_address,
    s.is_cms as site_is_cms,
    v.created_at as visit_created_at,
    v.updated_at as visit_updated_at,
    v.outside_system_id as visit_outside_system_id,
    v.number as visit_number,
    v.patient_type as visit_patient_type,
    m.created_at as mrn_created_at,
    m.updated_at as mrn_updated_at,
    m.mrn as mrn_value,
    p.created_at as physician_created_at,
    p.updated_at as physician_updated_at,
    p.first_name as physician_first_name,
    p.last_name as physician_last_name,
    p.middle_name as physician_middle_name,
    p.suffix as physician_suffix,
    p.prefix as physician_prefix,
    p.degree as physician_degree,
    p.npi as physician_npi,
    p.specialty as physician_specialty
FROM orders as o
LEFT JOIN sites as s ON o.site_id = s.id
LEFT JOIN visits as v ON o.visit_id = v.id
LEFT JOIN mrns as m ON o.mrn_id = m.id
LEFT JOIN physicians as p ON o.ordering_physician_id = p.id
WHERE
    o.site_id = $1
    AND o.number = $2;

-- name: UpdateOrder :one
UPDATE orders
SET
    updated_at = CURRENT_TIMESTAMP,
    outside_system_id = $2,
    site_id = $3,
    visit_id = $4,
    mrn_id = $5,
    ordering_physician_id = $6,
    arrival = $7,
    number = $8,
    current_status = $9
WHERE
    id = $1
RETURNING *;