-- +goose Up
ALTER TABLE IF EXISTS physicians DROP CONSTRAINT IF EXISTS physicians_name_npi_unique;
ALTER TABLE IF EXISTS physicians ADD COLUMN app_code TEXT;
ALTER TABLE physicians ADD CONSTRAINT physicians_name_app_code_unique UNIQUE (first_name, last_name, app_code);

-- +goose Down
ALTER TABLE physicians DROP CONSTRAINT IF EXISTS physicians_name_app_code_unique;
ALTER TABLE IF EXISTS physicians DROP COLUMN IF EXISTS app_code;
ALTER TABLE physicians ADD CONSTRAINT physicians_name_npi_unique UNIQUE (first_name, last_name, npi);
