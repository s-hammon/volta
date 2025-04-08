-- name: CreatePhysician :one
WITH upsert AS (
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
    ON CONFLICT (first_name, last_name, npi) DO UPDATE
    SET first_name = EXCLUDED.first_name,
        last_name = EXCLUDED.last_name,
        middle_name = EXCLUDED.middle_name,
        suffix = EXCLUDED.suffix,
        prefix = EXCLUDED.prefix,
        degree = EXCLUDED.degree,
        npi = EXCLUDED.npi,
        specialty = EXCLUDED.specialty
    WHERE physicians.first_name IS DISTINCT FROM EXCLUDED.first_name
        OR physicians.last_name IS DISTINCT FROM EXCLUDED.last_name
        OR physicians.middle_name IS DISTINCT FROM EXCLUDED.middle_name
        OR physicians.suffix IS DISTINCT FROM EXCLUDED.suffix
        OR physicians.prefix IS DISTINCT FROM EXCLUDED.prefix
        OR physicians.degree IS DISTINCT FROM EXCLUDED.degree
        OR physicians.npi IS DISTINCT FROM EXCLUDED.npi
        OR physicians.specialty IS DISTINCT FROM EXCLUDED.specialty
    RETURNING *
)
SELECT * FROM upsert
UNION ALL
SELECT * FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND npi = $7
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetPhysicianById :one
SELECT *
FROM physicians
WHERE id = $1;

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