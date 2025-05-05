-- +goose Up
ALTER TABLE IF EXISTS exams ADD COLUMN exam_cancelled_dt TIMESTAMP;
ALTER TABLE IF EXISTS exams ADD COLUMN prelim_report_id BIGINT REFERENCES reports(id) ON DELETE CASCADE;

CREATE INDEX exams_prelim_report_id ON exams(prelim_report_id ASC);
CREATE INDEX exams_final_report_id ON exams(final_report_id ASC);
CREATE INDEX exams_addendum_report_id ON exams(addendum_report_id ASC);

-- +goose Down
ALTER TABLE IF EXISTS exams DROP COLUMN IF EXISTS exam_cancelled_dt;
ALTER TABLE IF EXISTS exams DROP COLUMN IF EXISTS prelim_report_id;

DROP INDEX IF EXISTS exams_prelim_report_id;
DROP INDEX IF EXISTS exams_final_report_id;
DROP INDEX IF EXISTS exams_addendum_report_id;
