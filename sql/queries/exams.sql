-- name: CreateExam :one
WITH upsert as (
    INSERT INTO exams (
        visit_id, -- $1
        mrn_id, -- $2
        site_id, -- $3
        procedure_id, -- $4
        ordering_physician_id, -- $5
        accession, -- $6
        current_status, -- $7
        schedule_dt, -- $8
        begin_exam_dt, -- $9
        end_exam_dt, -- $10
        exam_cancelled_dt -- $11
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    ON CONFLICT (site_id, accession) DO UPDATE
    SET
        visit_id = EXCLUDED.visit_id,
        mrn_id = EXCLUDED.mrn_id,
        procedure_id = EXCLUDED.procedure_id,
        ordering_physician_id = EXCLUDED.ordering_physician_id,
        current_status = COALESCE(NULLIF(EXCLUDED.current_status, ''), exams.current_status),
        schedule_dt = COALESCE(EXCLUDED.schedule_dt, exams.schedule_dt),
        begin_exam_dt = COALESCE(EXCLUDED.begin_exam_dt, exams.begin_exam_dt),
        end_exam_dt = COALESCE(EXCLUDED.end_exam_dt, exams.end_exam_dt),
        exam_cancelled_dt = COALESCE(EXCLUDED.exam_cancelled_dt, exams.exam_cancelled_dt)
    WHERE
        exams.visit_id IS DISTINCT FROM EXCLUDED.visit_id
        OR exams.mrn_id IS DISTINCT FROM EXCLUDED.mrn_id
        OR exams.site_id IS DISTINCT FROM EXCLUDED.site_id
        OR exams.procedure_id IS DISTINCT FROM EXCLUDED.procedure_id
        OR exams.ordering_physician_id IS DISTINCT FROM EXCLUDED.ordering_physician_id
        OR COALESCE(NULLIF(EXCLUDED.current_status, ''), exams.current_status) IS DISTINCT FROM exams.current_status
        OR COALESCE(EXCLUDED.schedule_dt, exams.schedule_dt) IS DISTINCT FROM exams.schedule_dt
        OR COALESCE(EXCLUDED.begin_exam_dt, exams.begin_exam_dt) IS DISTINCT FROM exams.begin_exam_dt
        OR COALESCE(EXCLUDED.end_exam_dt, exams.end_exam_dt) IS DISTINCT FROM exams.end_exam_dt
        OR COALESCE(EXCLUDED.exam_cancelled_dt, exams.exam_cancelled_dt) IS DISTINCT FROM exams.exam_cancelled_dt
    RETURNING id
)
SELECT id FROM upsert
UNION ALL
SELECT id FROM exams
WHERE
    site_id = $4
    AND accession = $6
    AND NOT EXISTS (SELECT 1 FROM upsert);

-- name: GetExamById :one
SELECT * FROM exams
WHERE id = $1;

-- name: GetAllExams :many
SELECT *
FROM exams;

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
    o.created_at AS provider_created_at,
    o.updated_at AS provider_updated_at,
    o.first_name AS provider_first_name,
    o.last_name AS provider_last_name,
    o.middle_name AS provider_middle_name,
    o.suffix AS provider_suffix,
    o.prefix AS provider_prefix,
    o.degree AS provider_degree,
    o.npi AS provider_npi,
    o.specialty AS provider_specialty,
    s.created_at AS site_created_at,
    s.updated_at AS site_updated_at,
    s.code AS site_code,
    s.name AS site_name,
    s.address AS site_address,
    s.is_cms AS site_is_cms
FROM exams AS e
LEFT JOIN mrns AS m ON e.mrn_id = m.id
LEFT JOIN procedures AS p ON e.procedure_id = p.id and e.site_id = p.site_id
LEFT JOIN physicians AS o ON e.ordering_physician_id = o.id
LEFT JOIN sites AS s ON e.site_id = s.id
WHERE
    e.site_id = $1
    AND e.accession = $2;

-- name: UpdateExam :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    visit_id = $2,
    mrn_id = $3,
    site_id = $4,
    procedure_id = $5,
    ordering_physician_id = $6,
    accession = $7,
    current_status = $8,
    schedule_dt = $9,
    begin_exam_dt = $10,
    end_exam_dt = $11
WHERE id = $1
RETURNING *;

-- name: UpdateExamFinalReport :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    final_report_id = $2
WHERE id = $1
RETURNING *;

-- name: UpdateExamAddendumReport :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    addendum_report_id = $2
WHERE id = $1
RETURNING *;

-- name: UpdateExamPrelimReport :one
UPDATE exams
SET
    updated_at = CURRENT_TIMESTAMP,
    prelim_report_id = $2
WHERE id = $1
RETURNING *;
