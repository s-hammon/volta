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