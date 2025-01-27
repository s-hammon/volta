// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: exams.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createExam = `-- name: CreateExam :one
INSERT INTO exams (
    order_id,
    visit_id,
    mrn_id,
    site_id,
    procedure_id,
    accession,
    current_status,
    schedule_dt,
    begin_exam_dt,
    end_exam_dt
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
`

type CreateExamParams struct {
	OrderID       pgtype.Int8
	VisitID       pgtype.Int8
	MrnID         pgtype.Int8
	SiteID        pgtype.Int4
	ProcedureID   pgtype.Int4
	Accession     string
	CurrentStatus string
	ScheduleDt    pgtype.Timestamp
	BeginExamDt   pgtype.Timestamp
	EndExamDt     pgtype.Timestamp
}

func (q *Queries) CreateExam(ctx context.Context, arg CreateExamParams) (Exam, error) {
	row := q.db.QueryRow(ctx, createExam,
		arg.OrderID,
		arg.VisitID,
		arg.MrnID,
		arg.SiteID,
		arg.ProcedureID,
		arg.Accession,
		arg.CurrentStatus,
		arg.ScheduleDt,
		arg.BeginExamDt,
		arg.EndExamDt,
	)
	var i Exam
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.OutsideSystemID,
		&i.OrderID,
		&i.VisitID,
		&i.MrnID,
		&i.SiteID,
		&i.ProcedureID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}

const getExamBySiteIDAccession = `-- name: GetExamBySiteIDAccession :one
SELECT id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
FROM exams
WHERE
    site_id = $1
    AND accession = $2
`

type GetExamBySiteIDAccessionParams struct {
	SiteID    pgtype.Int4
	Accession string
}

func (q *Queries) GetExamBySiteIDAccession(ctx context.Context, arg GetExamBySiteIDAccessionParams) (Exam, error) {
	row := q.db.QueryRow(ctx, getExamBySiteIDAccession, arg.SiteID, arg.Accession)
	var i Exam
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.OutsideSystemID,
		&i.OrderID,
		&i.VisitID,
		&i.MrnID,
		&i.SiteID,
		&i.ProcedureID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}

const updateExamByID = `-- name: UpdateExamByID :one
UPDATE exams
SET
    order_id = $2,
    visit_id = $3,
    mrn_id = $4,
    site_id = $5,
    procedure_id = $6,
    accession = $7,
    current_status = $8,
    schedule_dt = $9,
    begin_exam_dt = $10,
    end_exam_dt = $11,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
`

type UpdateExamByIDParams struct {
	ID            int64
	OrderID       pgtype.Int8
	VisitID       pgtype.Int8
	MrnID         pgtype.Int8
	SiteID        pgtype.Int4
	ProcedureID   pgtype.Int4
	Accession     string
	CurrentStatus string
	ScheduleDt    pgtype.Timestamp
	BeginExamDt   pgtype.Timestamp
	EndExamDt     pgtype.Timestamp
}

func (q *Queries) UpdateExamByID(ctx context.Context, arg UpdateExamByIDParams) (Exam, error) {
	row := q.db.QueryRow(ctx, updateExamByID,
		arg.ID,
		arg.OrderID,
		arg.VisitID,
		arg.MrnID,
		arg.SiteID,
		arg.ProcedureID,
		arg.Accession,
		arg.CurrentStatus,
		arg.ScheduleDt,
		arg.BeginExamDt,
		arg.EndExamDt,
	)
	var i Exam
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.OutsideSystemID,
		&i.OrderID,
		&i.VisitID,
		&i.MrnID,
		&i.SiteID,
		&i.ProcedureID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}
