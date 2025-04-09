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
    SET
        middle_name = EXCLUDED.middle_name,
        suffix = EXCLUDED.suffix,
        prefix = EXCLUDED.prefix,
        degree = EXCLUDED.degree,
        specialty = COALESCE(NULLIF(EXCLUDED.specialty, ''), physicians.specialty)
    WHERE
        physicians.middle_name IS DISTINCT FROM EXCLUDED.middle_name
        OR physicians.suffix IS DISTINCT FROM EXCLUDED.suffix
        OR physicians.prefix IS DISTINCT FROM EXCLUDED.prefix
        OR physicians.degree IS DISTINCT FROM EXCLUDED.degree
        OR physicians.specialty IS DISTINCT FROM COALESCE(NULLIF(EXCLUDED.specialty, ''), physicians.specialty)
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