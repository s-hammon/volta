-- +goose up
CREATE TABLE IF NOT EXISTS exams (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    visit_id BIGINT REFERENCES visits(id) ON DELETE CASCADE,
    mrn_id BIGINT REFERENCES mrns(id) ON DELETE CASCADE,
    site_id INT REFERENCES sites(id) ON DELETE CASCADE,
    procedure_id INT REFERENCES procedures(id) ON DELETE CASCADE,
    arrival TIMESTAMP NOT NULL,
    accession TEXT NOT NULL,
    current_status TEXT NOT NULL,
    schedule_dt TIMESTAMP NOT NULL,
    begin_exam_dt TIMESTAMP NOT NULL,
    end_exam_dt TIMESTAMP NOT NULL
);

ALTER TABLE exams ADD CONSTRAINT unique_site_id_accession UNIQUE (site_id, accession);
CREATE INDEX exams_accession ON exams(accession ASC);
CREATE INDEX exams_order_id ON exams(order_id ASC);
CREATE INDEX exams_visit_id ON exams(visit_id ASC);
CREATE INDEX exams_mrn_id ON exams(mrn_id ASC);
CREATE INDEX exams_site_id ON exams(site_id ASC);
CREATE INDEX exams_procedure_id ON exams(procedure_id ASC);

CREATE TABLE IF NOT EXISTS reports (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    exam_id BIGINT REFERENCES exams(id) ON DELETE CASCADE,
    radiologist_id BIGINT REFERENCES physicians(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    impression TEXT NOT NULL,
    report_status TEXT NOT NULL,
    submitted_dt TIMESTAMP NOT NULL
);

ALTER TABLE reports ADD CONSTRAINT reports_exam_id_radiologist_id_status_unique UNIQUE (exam_id, radiologist_id, report_status);
CREATE INDEX reports_body_idx ON reports(body ASC);
CREATE INDEX reports_impression_idx ON reports(impression ASC);
CREATE INDEX reports_radiologist_id_idx ON reports(radiologist_id ASC);
CREATE INDEX reports_exam_id_idx ON reports(exam_id ASC);
CREATE INDEX reports_submitted_idx ON reports(submitted_dt ASC);

-- +goose down
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS exams;
