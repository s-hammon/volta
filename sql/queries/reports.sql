-- name: CreateReport :one
WITH upsert as (
    INSERT INTO reports (
        radiologist_id,
        body,
        impression,
        report_status,
        submitted_dt,
        message_id
    )
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (radiologist_id, impression, report_status, submitted_dt) DO UPDATE
    SET
        updated_at = CURRENT_TIMESTAMP,
        body = COALESCE(EXCLUDED.body, reports.body)
    WHERE
        COALESCE(EXCLUDED.body, reports.body) IS DISTINCT FROM EXCLUDED.body
    RETURNING id
)

SELECT id FROM upsert
UNION ALL
SELECT id FROM reports
WHERE
    radiologist_id = $1
    AND impression = $3
    AND report_status = $4
    AND submitted_dt = $5;

-- name: GetReportById :one
SELECT * FROM reports
WHERE id = $1;

-- name: GetReportByRadID :one
SELECT * FROM reports
where radiologist_id = $1;

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
