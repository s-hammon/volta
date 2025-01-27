-- name: CreateExam :one
INSERT INTO exams (
    order_id,
    visit_id,
    mrn_id,
    site_id,
    procedure_id,
    arrival,
    accession,
    current_status,
    schedule_dt,
    begin_exam_dt,
    end_exam_dt
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- -- name: GetExamBySiteIDAccession :one
SELECT *
FROM exams
WHERE
    site_id = $1
    AND accession = $2;