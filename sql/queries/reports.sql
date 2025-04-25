-- name: CreateReport :one
WITH upsert as (
    INSERT INTO reports (
        radiologist_id,
        body,
        impression,
        report_status,
        submitted_dt
    )
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (radiologist_id, impression, report_status, submitted_dt) DO UPDATE
    SET
        body = COALESCE(EXCLUDED.body, reports.body)
    WHERE
        COALESCE(EXCLUDED.body, reports.body) IS DISTINCT FROM EXCLUDED.body
    RETURNING *
)

SELECT * FROM upsert
UNION ALL
SELECT * FROM reports
WHERE
    radiologist_id = $1
    AND impression = $3
    AND report_status = $4
    AND submitted_dt = $5;

-- name: GetReportById :one
SELECT * FROM reports
WHERE id = $1;

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
