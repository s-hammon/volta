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