-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE,
    site_id INT REFERENCES sites(id) ON DELETE CASCADE,
    visit_id BIGINT REFERENCES visits(id) ON DELETE CASCADE,
    mrn_id BIGINT REFERENCES mrns(id) ON DELETE CASCADE,
    ordering_physician_id BIGINT REFERENCES physicians(id) ON DELETE CASCADE,
    arrival TIMESTAMP NOT NULL,
    number TEXT NOT NULL,
    current_status TEXT NOT NULL
);

ALTER TABLE orders ADD CONSTRAINT unique_site_id_number UNIQUE (site_id, number);
CREATE INDEX orders_number_outside_system_id_idx ON orders(number ASC, outside_system_id ASC);
CREATE INDEX orders_outside_system_id_idx ON orders(outside_system_id ASC);
CREATE INDEX orders_site_id_idx ON orders(site_id ASC);
CREATE INDEX orders_visit_id_idx ON orders(visit_id ASC);
CREATE INDEX orders_mrn_id_idx ON orders(mrn_id ASC);
CREATE INDEX orders_ordering_physician_id_idx ON orders(ordering_physician_id ASC);
CREATE INDEX orders_arrival_idx ON orders(arrival ASC);
CREATE INDEX orders_number_idx ON orders(number ASC);

CREATE TABLE IF NOT EXISTS exams (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE,
    site_id INT REFERENCES sites(id) ON DELETE CASCADE,
    order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE,
    visit_id BIGINT REFERENCES visits(id) ON DELETE CASCADE,
    mrn_id BIGINT REFERENCES mrns(id) ON DELETE CASCADE,
    procedure_id INT REFERENCES procedures(id) ON DELETE CASCADE,
    accession TEXT NOT NULL,
    priority TEXT NOT NULL
);

CREATE INDEX exams_accession_idx ON exams(accession ASC);
CREATE INDEX exams_accession_outside_system_id_idx ON exams(accession ASC, outside_system_id ASC);
CREATE INDEX exams_outside_system_id_idx ON exams(outside_system_id ASC);
CREATE INDEX exams_site_id_idx ON exams(site_id ASC);
CREATE INDEX exams_order_id_idx ON exams(order_id ASC);
CREATE INDEX exams_visit_id_idx ON exams(visit_id ASC);
CREATE INDEX exams_mrn_id_idx ON exams(mrn_id ASC);
CREATE INDEX exams_procedure_id_idx ON exams(procedure_id ASC);
CREATE INDEX exams_priority_idx ON exams(priority ASC);

-- +goose Down
DROP TABLE IF EXISTS exams;
DROP TABLE IF EXISTS orders;