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