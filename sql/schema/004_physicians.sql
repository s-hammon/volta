-- +goose Up
CREATE TABLE IF NOT EXISTS physicians (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT,
    suffix TEXT,
    prefix TEXT,
    degree TEXT,
    npi TEXT NOT NULL,
    specialty TEXT
);

ALTER TABLE physicians ADD CONSTRAINT physicians_name_npi_unique UNIQUE (first_name, last_name, npi);
CREATE INDEX physicians_last_first_name_idx ON physicians(last_name ASC, first_name ASC);

-- +goose Down
DROP TABLE IF EXISTS physicians;
