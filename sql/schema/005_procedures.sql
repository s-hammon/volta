-- +goose Up
CREATE TABLE IF NOT EXISTS procedures (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    code TEXT NOT NULL,
    description TEXT NOT NULL,
    specialty TEXT,
    modality TEXT
);

CREATE INDEX procedures_code_idx ON procedures(code ASC);
CREATE INDEX procedures_specialty_idx ON procedures(specialty ASC);

-- +goose Down
DROP TABLE IF EXISTS procedures;