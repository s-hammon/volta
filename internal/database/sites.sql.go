// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: sites.sql

package database

import (
	"context"
)

const createSite = `-- name: CreateSite :one
INSERT INTO sites (code, name, address, is_cms)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code, name) DO UPDATE
SET name = EXCLUDED.name,
    address = EXCLUDED.address,
    is_cms = EXCLUDED.is_cms
WHERE sites.name IS DISTINCT FROM EXCLUDED.name
    OR sites.address IS DISTINCT FROM EXCLUDED.address
    OR sites.is_cms IS DISTINCT FROM EXCLUDED.is_cms
RETURNING id, created_at, updated_at, code, name, address, is_cms
`

type CreateSiteParams struct {
	Code    string
	Name    string
	Address string
	IsCms   bool
}

func (q *Queries) CreateSite(ctx context.Context, arg CreateSiteParams) (Site, error) {
	row := q.db.QueryRow(ctx, createSite,
		arg.Code,
		arg.Name,
		arg.Address,
		arg.IsCms,
	)
	var i Site
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Code,
		&i.Name,
		&i.Address,
		&i.IsCms,
	)
	return i, err
}

const getSiteByCode = `-- name: GetSiteByCode :one
SELECT id, created_at, updated_at, code, name, address, is_cms
FROM sites
WHERE code = $1
`

func (q *Queries) GetSiteByCode(ctx context.Context, code string) (Site, error) {
	row := q.db.QueryRow(ctx, getSiteByCode, code)
	var i Site
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Code,
		&i.Name,
		&i.Address,
		&i.IsCms,
	)
	return i, err
}
