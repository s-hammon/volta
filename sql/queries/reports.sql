-- name: CreateReport :one
INSERT INTO reports (
    exam_id,
    radiologist_id,
    body,
    impression,
    report_status,
    submitted_dt
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
