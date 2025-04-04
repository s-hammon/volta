-- name: CreatePatient :one
INSERT INTO patients (
    first_name,
    last_name,
    middle_name,
    suffix,
    prefix,
    degree,
    dob,
    sex,
    ssn,
    home_phone,
    work_phone,
    cell_phone
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
SET first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    middle_name = EXCLUDED.middle_name,
    suffix = EXCLUDED.suffix,
    prefix = EXCLUDED.prefix,
    degree = EXCLUDED.degree,
    dob = EXCLUDED.dob,
    sex = EXCLUDED.sex,
    home_phone = EXCLUDED.home_phone,
    work_phone = EXCLUDED.work_phone,
    cell_phone = EXCLUDED.cell_phone
WHERE patients.first_name IS DISTINCT FROM EXCLUDED.first_name
    OR patients.last_name IS DISTINCT FROM EXCLUDED.last_name
    OR patients.middle_name IS DISTINCT FROM EXCLUDED.middle_name
    OR patients.suffix IS DISTINCT FROM EXCLUDED.suffix
    OR patients.prefix IS DISTINCT FROM EXCLUDED.prefix
    OR patients.degree IS DISTINCT FROM EXCLUDED.degree
    OR patients.dob IS DISTINCT FROM EXCLUDED.dob
    OR patients.sex IS DISTINCT FROM EXCLUDED.sex
    OR patients.home_phone IS DISTINCT FROM EXCLUDED.home_phone
    OR patients.work_phone IS DISTINCT FROM EXCLUDED.work_phone
    OR patients.cell_phone IS DISTINCT FROM EXCLUDED.cell_phone
RETURNING *;

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