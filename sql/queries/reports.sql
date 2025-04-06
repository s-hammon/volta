-- name: CreateReport :one
INSERT INTO reports (
    radiologist_id,
    body,
    impression,
    report_status,
    submitted_dt
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllReports :many

SELECT *
FROM reports;

-- name: GetReportByUniqueFields :one
SELECT *
FROM reports
WHERE
    radiologist_id = $1
    AND impression = $2
    AND report_status = $3
    AND submitted_dt = $4;