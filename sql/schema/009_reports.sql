-- +goose Up
CREATE TABLE IF NOT EXISTS reports (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    exam_id BIGINT REFERENCES exams(id) ON DELETE CASCADE,
    radiologist_id BIGINT REFERENCES physicians(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    impression TEXT NOT NULL,
    report_status TEXT NOT NULL,
    submitted_dt TIMESTAMP
);

ALTER TABLE reports ADD CONSTRAINT reports_exam_id_radiologist_id_status_unique UNIQUE (exam_id, radiologist_id, report_status);
CREATE INDEX reports_body_idx ON reports(body ASC);
CREATE INDEX reports_impression_idx ON reports(impression ASC);
CREATE INDEX reports_radiologist_id_idx ON reports(radiologist_id ASC);
CREATE INDEX reports_exam_id_idx ON reports(exam_id ASC);
CREATE INDEX reports_submitted_idx ON reports(submitted_dt ASC);

-- +goose Down
DROP TABLE IF EXISTS reports;