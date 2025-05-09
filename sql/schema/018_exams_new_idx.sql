-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS exams_site_id_accession ON exams(site_id, accession);

-- +goose Down
DROP INDEX IF EXISTS exams_site_id_accession;
