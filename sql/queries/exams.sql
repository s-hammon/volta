-- name: CreateExam :one
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
RETURNING *;

-- name: GetExamBySiteIDAccession :one
SELECT *
FROM exams
WHERE
    site_id = $1
    AND accession = $2;

-- name: UpdateExamByID :one
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
RETURNING *;