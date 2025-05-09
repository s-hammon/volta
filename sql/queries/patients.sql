-- name: CreatePatient :one
WITH upsert AS (
    INSERT INTO patients (
        first_name, -- $1
        last_name, -- $2
        middle_name, -- $3
        suffix, -- $4
        prefix, -- $5
        degree, -- $6
        dob, -- $7
        sex, -- $8
        ssn, -- $9
        home_phone, -- $10
        work_phone, -- $11
        cell_phone -- $12
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
        $9,
        $10,
        $11,
        $12
    )
    ON CONFLICT (ssn) DO UPDATE
    SET first_name = COALESCE(NULLIF(EXCLUDED.first_name, ''), patients.first_name),
        last_name = COALESCE(NULLIF(EXCLUDED.last_name, ''), patients.last_name),
        middle_name = COALESCE(NULLIF(EXCLUDED.middle_name, ''), patients.middle_name),
        suffix = COALESCE(NULLIF(EXCLUDED.suffix, ''), patients.suffix),
        prefix = COALESCE(NULLIF(EXCLUDED.prefix, ''), patients.prefix),
        degree = COALESCE(NULLIF(EXCLUDED.degree, ''), patients.degree),
        dob = EXCLUDED.dob,
        sex = EXCLUDED.sex,
        home_phone = EXCLUDED.home_phone,
        work_phone = EXCLUDED.work_phone,
        cell_phone = EXCLUDED.cell_phone
    WHERE
        COALESCE(NULLIF(EXCLUDED.first_name, ''), patients.first_name) IS DISTINCT FROM EXCLUDED.first_name
        OR COALESCE(NULLIF(EXCLUDED.last_name, ''), patients.last_name) IS DISTINCT FROM EXCLUDED.last_name
        OR COALESCE(NULLIF(EXCLUDED.middle_name, ''), patients.middle_name) IS DISTINCT FROM EXCLUDED.middle_name
        OR patients.suffix IS DISTINCT FROM EXCLUDED.suffix
        OR patients.prefix IS DISTINCT FROM EXCLUDED.prefix
        OR patients.degree IS DISTINCT FROM EXCLUDED.degree
        OR patients.dob IS DISTINCT FROM EXCLUDED.dob
        OR patients.sex IS DISTINCT FROM EXCLUDED.sex
        OR patients.home_phone IS DISTINCT FROM EXCLUDED.home_phone
        OR patients.work_phone IS DISTINCT FROM EXCLUDED.work_phone
        OR patients.cell_phone IS DISTINCT FROM EXCLUDED.cell_phone
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM patients
WHERE
    ssn = $9
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetPatientById :one
SELECT *
FROM patients
WHERE id = $1;

-- name: GetPatientByNameSSN :one
SELECT *
FROM patients
WHERE
    first_name = $1
    AND last_name = $2
    AND dob = $3
    AND ssn = $4;

-- name: UpdatePatient :one
UPDATE patients
SET
    updated_at = CURRENT_TIMESTAMP,
    first_name = $2,
    last_name = $3,
    middle_name = $4,
    suffix = $5,
    prefix = $6,
    degree = $7,
    dob = $8,
    sex = $9,
    ssn = $10,
    home_phone = $11,
    work_phone = $12,
    cell_phone = $13
WHERE id = $1
RETURNING *;
