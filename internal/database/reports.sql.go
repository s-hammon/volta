// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: reports.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createReport = `-- name: CreateReport :one
INSERT INTO reports (
    radiologist_id,
    body,
    impression,
    report_status,
    submitted_dt
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, radiologist_id, body, impression, report_status, submitted_dt
`

type CreateReportParams struct {
	RadiologistID pgtype.Int8
	Body          string
	Impression    string
	ReportStatus  string
	SubmittedDt   pgtype.Timestamp
}

func (q *Queries) CreateReport(ctx context.Context, arg CreateReportParams) (Report, error) {
	row := q.db.QueryRow(ctx, createReport,
		arg.RadiologistID,
		arg.Body,
		arg.Impression,
		arg.ReportStatus,
		arg.SubmittedDt,
	)
	var i Report
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RadiologistID,
		&i.Body,
		&i.Impression,
		&i.ReportStatus,
		&i.SubmittedDt,
	)
	return i, err
}

const getAllReports = `-- name: GetAllReports :many
SELECT id, created_at, updated_at, radiologist_id, body, impression, report_status, submitted_dt
FROM reports
`

func (q *Queries) GetAllReports(ctx context.Context) ([]Report, error) {
	rows, err := q.db.Query(ctx, getAllReports)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Report
	for rows.Next() {
		var i Report
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.RadiologistID,
			&i.Body,
			&i.Impression,
			&i.ReportStatus,
			&i.SubmittedDt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getReportById = `-- name: GetReportById :one
SELECT id, created_at, updated_at, radiologist_id, body, impression, report_status, submitted_dt FROM reports
WHERE id = $1
`

func (q *Queries) GetReportById(ctx context.Context, id int64) (Report, error) {
	row := q.db.QueryRow(ctx, getReportById, id)
	var i Report
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RadiologistID,
		&i.Body,
		&i.Impression,
		&i.ReportStatus,
		&i.SubmittedDt,
	)
	return i, err
}

const getReportByUniqueFields = `-- name: GetReportByUniqueFields :one
SELECT id, created_at, updated_at, radiologist_id, body, impression, report_status, submitted_dt
FROM reports
WHERE
    radiologist_id = $1
    AND impression = $2
    AND report_status = $3
    AND submitted_dt = $4
`

type GetReportByUniqueFieldsParams struct {
	RadiologistID pgtype.Int8
	Impression    string
	ReportStatus  string
	SubmittedDt   pgtype.Timestamp
}

func (q *Queries) GetReportByUniqueFields(ctx context.Context, arg GetReportByUniqueFieldsParams) (Report, error) {
	row := q.db.QueryRow(ctx, getReportByUniqueFields,
		arg.RadiologistID,
		arg.Impression,
		arg.ReportStatus,
		arg.SubmittedDt,
	)
	var i Report
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RadiologistID,
		&i.Body,
		&i.Impression,
		&i.ReportStatus,
		&i.SubmittedDt,
	)
	return i, err
}
