-- name: CreatePhysician :one
WITH upsert AS (
    INSERT INTO physicians (
        first_name, -- $1
        last_name, -- $2
        middle_name, -- $3
        suffix, -- $4
        prefix, -- $5
        degree, -- $6
        app_code, -- $7
        npi, -- $8
        message_id -- $9
    )
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    )
    ON CONFLICT (first_name, last_name, app_code) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
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
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND app_code = $7
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

-- name: GetPhysicianByNameAppCode :one
SELECT *
FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND app_code = $3;

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
    app_code = $8,
    npi = $9,
    specialty = $10
WHERE id = $1
RETURNING *;
