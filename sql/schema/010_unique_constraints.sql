-- +goose Up
ALTER TABLE mrns ADD CONSTRAINT mrns_site_id_patient_id_unique UNIQUE (site_id, patient_id);
ALTER TABLE procedures ADD CONSTRAINT procedures_site_id_code_unique UNIQUE (site_id, code);

-- +goose Down
ALTER TABLE mrns DROP CONSTRAINT IF EXISTS mrns_site_id_patient_id_unique;
ALTER TABLE procedures DROP CONSTRAINT IF EXISTS procedures_site_id_code_unique;