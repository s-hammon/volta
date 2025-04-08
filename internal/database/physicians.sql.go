// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: physicians.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPhysician = `-- name: CreatePhysician :one
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
    RETURNING id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty
)
SELECT id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty FROM upsert
UNION ALL
SELECT id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND npi = $7
    AND NOT EXISTS (SELECT 1 FROM upsert)
`

type CreatePhysicianParams struct {
	FirstName  string
	LastName   string
	MiddleName pgtype.Text
	Suffix     pgtype.Text
	Prefix     pgtype.Text
	Degree     pgtype.Text
	Npi        string
	Specialty  pgtype.Text
}

type CreatePhysicianRow struct {
	ID         int64
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
	FirstName  string
	LastName   string
	MiddleName pgtype.Text
	Suffix     pgtype.Text
	Prefix     pgtype.Text
	Degree     pgtype.Text
	Npi        string
	Specialty  pgtype.Text
}

func (q *Queries) CreatePhysician(ctx context.Context, arg CreatePhysicianParams) (CreatePhysicianRow, error) {
	row := q.db.QueryRow(ctx, createPhysician,
		arg.FirstName,
		arg.LastName,
		arg.MiddleName,
		arg.Suffix,
		arg.Prefix,
		arg.Degree,
		arg.Npi,
		arg.Specialty,
	)
	var i CreatePhysicianRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.FirstName,
		&i.LastName,
		&i.MiddleName,
		&i.Suffix,
		&i.Prefix,
		&i.Degree,
		&i.Npi,
		&i.Specialty,
	)
	return i, err
}

const getPhysicianById = `-- name: GetPhysicianById :one
SELECT id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty
FROM physicians
WHERE id = $1
`

func (q *Queries) GetPhysicianById(ctx context.Context, id int64) (Physician, error) {
	row := q.db.QueryRow(ctx, getPhysicianById, id)
	var i Physician
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.FirstName,
		&i.LastName,
		&i.MiddleName,
		&i.Suffix,
		&i.Prefix,
		&i.Degree,
		&i.Npi,
		&i.Specialty,
	)
	return i, err
}

const getPhysicianByNameNPI = `-- name: GetPhysicianByNameNPI :one
SELECT id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty
FROM physicians
WHERE
    first_name = $1
    AND last_name = $2
    AND npi = $3
`

type GetPhysicianByNameNPIParams struct {
	FirstName string
	LastName  string
	Npi       string
}

func (q *Queries) GetPhysicianByNameNPI(ctx context.Context, arg GetPhysicianByNameNPIParams) (Physician, error) {
	row := q.db.QueryRow(ctx, getPhysicianByNameNPI, arg.FirstName, arg.LastName, arg.Npi)
	var i Physician
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.FirstName,
		&i.LastName,
		&i.MiddleName,
		&i.Suffix,
		&i.Prefix,
		&i.Degree,
		&i.Npi,
		&i.Specialty,
	)
	return i, err
}

const updatePhysician = `-- name: UpdatePhysician :one
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
RETURNING id, created_at, updated_at, first_name, last_name, middle_name, suffix, prefix, degree, npi, specialty
`

type UpdatePhysicianParams struct {
	ID         int64
	FirstName  string
	LastName   string
	MiddleName pgtype.Text
	Suffix     pgtype.Text
	Prefix     pgtype.Text
	Degree     pgtype.Text
	Npi        string
	Specialty  pgtype.Text
}

func (q *Queries) UpdatePhysician(ctx context.Context, arg UpdatePhysicianParams) (Physician, error) {
	row := q.db.QueryRow(ctx, updatePhysician,
		arg.ID,
		arg.FirstName,
		arg.LastName,
		arg.MiddleName,
		arg.Suffix,
		arg.Prefix,
		arg.Degree,
		arg.Npi,
		arg.Specialty,
	)
	var i Physician
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.FirstName,
		&i.LastName,
		&i.MiddleName,
		&i.Suffix,
		&i.Prefix,
		&i.Degree,
		&i.Npi,
		&i.Specialty,
	)
	return i, err
}
