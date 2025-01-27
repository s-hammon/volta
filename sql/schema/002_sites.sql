-- +goose Up
CREATE TABLE IF NOT EXISTS sites (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    is_cms BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX sites_code_idx ON sites(code ASC);
ALTER TABLE sites ADD CONSTRAINT sites_code_name_unique UNIQUE (code, name);

ALTER TABLE mrns ADD COLUMN site_id INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE;
ALTER TABLE mrns ADD CONSTRAINT mrns_site_id_mrn_patient_id UNIQUE (site_id, mrn, patient_id);

CREATE INDEX mrns_site_id_idx ON mrns(site_id ASC);
CREATE INDEX mrns_site_id_mrn_idx ON mrns(site_id ASC, mrn ASC);

CREATE TABLE IF NOT EXISTS metasite (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    is_org BOOLEAN NOT NULL
);

CREATE INDEX metasite_name_idx ON metasite(name ASC);

CREATE TABLE IF NOT EXISTS outside_systems (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metasite_id INT REFERENCES metasite(id) ON DELETE CASCADE,
    name TEXT NOT NULL
);

CREATE INDEX outside_systems_name_metasite_id_idx ON outside_systems(name ASC, metasite_id ASC);
CREATE INDEX outside_systems_metasite_id_idx ON outside_systems(metasite_id ASC);

-- +goose Down
DROP TABLE IF EXISTS outside_systems;
DROP TABLE IF EXISTS metasite;
ALTER TABLE mrns DROP COLUMN site_id;
DROP TABLE IF EXISTS sites;
