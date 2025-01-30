-- name: CreatePhysician :one
INSERT INTO physicians (
    first_name,
    last_name,
    middle_name,
    suffix,
    prefix,
    degree,
    npi,
    specialty
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

-- name: GetPhysicianByNameNPI :one
SELECT *
FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND npi = $3;

-- name: UpdatePhysician :one
UPDATE physicians
SET
    updated_at = CURRENT_TIMESTAMP,
    first_name = $2,
    last_name = $3,
    middle_name = $4,
    suffix = $5,
    prefix = $6,
    degree = $7,
    npi = $8,
    specialty = $9
WHERE id = $1
RETURNING *;