-- +goose Up
CREATE TABLE IF NOT EXISTS patients (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    middle_name TEXT,
    suffix TEXT,
    prefix TEXT,
    degree TEXT,
    dob DATE NOT NULL,
    sex CHAR NOT NULL,
    ssn char(11),
    home_phone TEXT,
    work_phone TEXT,
    cell_phone TEXT
);

ALTER TABLE patients ADD CONSTRAINT patients_ssn_unique UNIQUE (ssn);
CREATE INDEX patients_dob_idx ON patients(dob ASC);
CREATE INDEX patients_gender_idx ON patients(sex ASC);
CREATE INDEX patients_name_idx ON patients(last_name ASC, first_name ASC);

CREATE TABLE IF NOT EXISTS mrns (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    patient_id BIGINT REFERENCES patients(id) ON DELETE CASCADE,
    mrn TEXT NOT NULL
);

CREATE INDEX mrns_mrn_idx ON mrns(mrn ASC);
CREATE INDEX mrns_patient_id_idx ON mrns(patient_id ASC);

-- +goose Down
DROP TABLE IF EXISTS mrns;
DROP TABLE IF EXISTS patients;