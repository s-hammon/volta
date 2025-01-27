-- +goose Up
CREATE TABLE IF NOT EXISTS visits (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE,
    site_id INT REFERENCES sites(id) ON DELETE CASCADE,
    mrn_id BIGINT REFERENCES mrns(id) ON DELETE CASCADE,
    number TEXT NOT NULL,
    patient_type SMALLINT NOT NULL
);

ALTER TABLE visits ADD CONSTRAINT visits_site_id_mrn_id_number_unique UNIQUE (site_id, mrn_id, number);
CREATE INDEX visits_outside_system_id_number_idx ON visits(outside_system_id ASC, number ASC);
CREATE INDEX visits_outside_system_id_idx ON visits(outside_system_id ASC);
CREATE INDEX visits_mrn_id_idx ON visits(mrn_id ASC);
CREATE INDEX visits_site_id_idx ON visits(site_id ASC);
CREATE INDEX visits_number_idx ON visits(number ASC);

-- +goose Down
DROP TABLE IF EXISTS visits;