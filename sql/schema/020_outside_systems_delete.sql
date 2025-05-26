-- +goose Up
ALTER TABLE exams DROP COLUMN outside_system_id;
ALTER TABLE visits DROP COLUMN outside_system_id;

DROP TABLE IF EXISTS outside_systems;

ALTER TABLE sites ADD COLUMN metasite_id INT REFERENCES metasite(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE sites DROP COLUMN metasite_id;

CREATE TABLE IF NOT EXISTS outside_systems (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metasite_id INT REFERENCES metasite(id) ON DELETE CASCADE,
    name TEXT NOT NULL
);

CREATE INDEX outside_systems_name_metasite_id_idx ON outside_systems(name ASC, metasite_id ASC);
CREATE INDEX outside_systems_metasite_id_idx ON outside_systems(metasite_id ASC);

ALTER TABLE visits ADD COLUMN outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE;
ALTER TABLE exams ADD COLUMN outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE;

