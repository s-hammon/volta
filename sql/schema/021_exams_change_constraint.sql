-- +goose Up
ALTER TABLE IF EXISTS exams ADD COLUMN sending_app TEXT NOT NULL;
ALTER TABLE IF EXISTS exams ADD CONSTRAINT unique_sending_app_accession UNIQUE (sending_app, accession);
ALTER TABLE IF EXISTS exams DROP CONSTRAINT IF EXISTS unique_site_id_accession;
DROP INDEX IF EXISTS exams_site_id_accession;

-- +goose Down
CREATE UNIQUE INDEX IF NOT EXISTS exams_site_id_accession ON exams(site_id, accession);
ALTER TABLE IF EXISTS exams ADD CONSTRAINT unique_site_id_accession UNIQUE (site_id, accession);
ALTER TABLE IF EXISTS exams DROP CONSTRAINT IF EXISTS unique_sending_app_accession;
ALTER TABLE IF EXISTS exams DROP COLUMN sending_app;
