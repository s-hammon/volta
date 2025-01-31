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
SELECT
    e.*,
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
    AND e.accession = $2;

-- name: UpdateExam :one
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
RETURNING *;