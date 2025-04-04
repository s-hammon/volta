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
ON CONFLICT (site_id, accession) DO UPDATE
SET order_id = EXCLUDED.order_id,
    visit_id = EXCLUDED.visit_id,
    mrn_id = EXCLUDED.mrn_id,
    site_id = EXCLUDED.site_id,
    procedure_id = EXCLUDED.procedure_id,
    current_status = EXCLUDED.current_status,
    schedule_dt = EXCLUDED.schedule_dt,
    begin_exam_dt = EXCLUDED.begin_exam_dt,
    end_exam_dt = EXCLUDED.end_exam_dt
WHERE exams.order_id IS DISTINCT FROM EXCLUDED.order_id
    OR exams.visit_id IS DISTINCT FROM EXCLUDED.visit_id
    OR exams.mrn_id IS DISTINCT FROM EXCLUDED.mrn_id
    OR exams.site_id IS DISTINCT FROM EXCLUDED.site_id
    OR exams.procedure_id IS DISTINCT FROM EXCLUDED.procedure_id
    OR exams.current_status IS DISTINCT FROM EXCLUDED.current_status
    OR exams.schedule_dt IS DISTINCT FROM EXCLUDED.schedule_dt
    OR exams.begin_exam_dt IS DISTINCT FROM EXCLUDED.begin_exam_dt
    OR exams.end_exam_dt IS DISTINCT FROM EXCLUDED.end_exam_dt
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, final_report_id, addendum_report_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
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
		&i.FinalReportID,
		&i.AddendumReportID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}

const getExamBySiteIDAccession = `-- name: GetExamBySiteIDAccession :one
SELECT
    e.id, e.created_at, e.updated_at, e.outside_system_id, e.order_id, e.visit_id, e.mrn_id, e.site_id, e.procedure_id, e.final_report_id, e.addendum_report_id, e.accession, e.current_status, e.schedule_dt, e.begin_exam_dt, e.end_exam_dt,
    m.created_at AS mrn_created_at,
    m.updated_at AS mrn_updated_at,
    m.mrn AS mrn_value,
    p.created_at AS procedure_created_at,
    p.updated_at AS procedure_updated_at,
    p.code AS procedure_code,
    p.description AS procedure_description,
    p.specialty AS procedure_specialty,
    p.modality AS procedure_modality,
    s.created_at AS site_created_at,
    s.updated_at AS site_updated_at,
    s.code AS site_code,
    s.name AS site_name,
    s.address AS site_address,
    s.is_cms AS site_is_cms
FROM exams AS e
LEFT JOIN mrns AS m ON e.mrn_id = m.id
LEFT JOIN procedures AS p ON e.procedure_id = p.id and e.site_id = p.site_id
LEFT JOIN sites AS s ON e.site_id = s.id
WHERE
    e.site_id = $1
    AND e.accession = $2
`

type GetExamBySiteIDAccessionParams struct {
	SiteID    pgtype.Int4
	Accession string
}

type GetExamBySiteIDAccessionRow struct {
	ID                   int64
	CreatedAt            pgtype.Timestamp
	UpdatedAt            pgtype.Timestamp
	OutsideSystemID      pgtype.Int4
	OrderID              pgtype.Int8
	VisitID              pgtype.Int8
	MrnID                pgtype.Int8
	SiteID               pgtype.Int4
	ProcedureID          pgtype.Int4
	FinalReportID        pgtype.Int8
	AddendumReportID     pgtype.Int8
	Accession            string
	CurrentStatus        string
	ScheduleDt           pgtype.Timestamp
	BeginExamDt          pgtype.Timestamp
	EndExamDt            pgtype.Timestamp
	MrnCreatedAt         pgtype.Timestamp
	MrnUpdatedAt         pgtype.Timestamp
	MrnValue             pgtype.Text
	ProcedureCreatedAt   pgtype.Timestamp
	ProcedureUpdatedAt   pgtype.Timestamp
	ProcedureCode        pgtype.Text
	ProcedureDescription pgtype.Text
	ProcedureSpecialty   pgtype.Text
	ProcedureModality    pgtype.Text
	SiteCreatedAt        pgtype.Timestamp
	SiteUpdatedAt        pgtype.Timestamp
	SiteCode             pgtype.Text
	SiteName             pgtype.Text
	SiteAddress          pgtype.Text
	SiteIsCms            pgtype.Bool
}

func (q *Queries) GetExamBySiteIDAccession(ctx context.Context, arg GetExamBySiteIDAccessionParams) (GetExamBySiteIDAccessionRow, error) {
	row := q.db.QueryRow(ctx, getExamBySiteIDAccession, arg.SiteID, arg.Accession)
	var i GetExamBySiteIDAccessionRow
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
		&i.FinalReportID,
		&i.AddendumReportID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
		&i.MrnCreatedAt,
		&i.MrnUpdatedAt,
		&i.MrnValue,
		&i.ProcedureCreatedAt,
		&i.ProcedureUpdatedAt,
		&i.ProcedureCode,
		&i.ProcedureDescription,
		&i.ProcedureSpecialty,
		&i.ProcedureModality,
		&i.SiteCreatedAt,
		&i.SiteUpdatedAt,
		&i.SiteCode,
		&i.SiteName,
		&i.SiteAddress,
		&i.SiteIsCms,
	)
	return i, err
}

const updateExam = `-- name: UpdateExam :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    order_id = $2,
    visit_id = $3,
    mrn_id = $4,
    site_id = $5,
    procedure_id = $6,
    accession = $7,
    current_status = $8,
    schedule_dt = $9,
    begin_exam_dt = $10,
    end_exam_dt = $11
WHERE id = $1
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, final_report_id, addendum_report_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
`

type UpdateExamParams struct {
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

func (q *Queries) UpdateExam(ctx context.Context, arg UpdateExamParams) (Exam, error) {
	row := q.db.QueryRow(ctx, updateExam,
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
		&i.FinalReportID,
		&i.AddendumReportID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}

const updateExamAddendumReport = `-- name: UpdateExamAddendumReport :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    addendum_report_id = $2
WHERE id = $1
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, final_report_id, addendum_report_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
`

type UpdateExamAddendumReportParams struct {
	ID               int64
	AddendumReportID pgtype.Int8
}

func (q *Queries) UpdateExamAddendumReport(ctx context.Context, arg UpdateExamAddendumReportParams) (Exam, error) {
	row := q.db.QueryRow(ctx, updateExamAddendumReport, arg.ID, arg.AddendumReportID)
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
		&i.FinalReportID,
		&i.AddendumReportID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}

const updateExamFinalReport = `-- name: UpdateExamFinalReport :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    final_report_id = $2
WHERE id = $1
RETURNING id, created_at, updated_at, outside_system_id, order_id, visit_id, mrn_id, site_id, procedure_id, final_report_id, addendum_report_id, accession, current_status, schedule_dt, begin_exam_dt, end_exam_dt
`

type UpdateExamFinalReportParams struct {
	ID            int64
	FinalReportID pgtype.Int8
}

func (q *Queries) UpdateExamFinalReport(ctx context.Context, arg UpdateExamFinalReportParams) (Exam, error) {
	row := q.db.QueryRow(ctx, updateExamFinalReport, arg.ID, arg.FinalReportID)
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
		&i.FinalReportID,
		&i.AddendumReportID,
		&i.Accession,
		&i.CurrentStatus,
		&i.ScheduleDt,
		&i.BeginExamDt,
		&i.EndExamDt,
	)
	return i, err
}
