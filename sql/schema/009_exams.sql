-- +goose Up
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
    final_report_id BIGINT REFERENCES reports(id) ON DELETE CASCADE,
    addendum_report_id BIGINT REFERENCES reports(id) ON DELETE CASCADE,
    accession TEXT NOT NULL,
    current_status TEXT NOT NULL,
    schedule_dt TIMESTAMP,
    begin_exam_dt TIMESTAMP,
    end_exam_dt TIMESTAMP
);

ALTER TABLE exams ADD CONSTRAINT unique_site_id_accession UNIQUE (site_id, accession);
CREATE INDEX exams_accession ON exams(accession ASC);
CREATE INDEX exams_order_id ON exams(order_id ASC);
CREATE INDEX exams_visit_id ON exams(visit_id ASC);
CREATE INDEX exams_mrn_id ON exams(mrn_id ASC);
CREATE INDEX exams_site_id ON exams(site_id ASC);
CREATE INDEX exams_procedure_id ON exams(procedure_id ASC);

-- +goose Down
DROP TABLE IF EXISTS exams;
