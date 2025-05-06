-- +goose Up
DROP TABLE IF EXISTS orders;

-- +goose Down
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    outside_system_id INT REFERENCES outside_systems(id) ON DELETE CASCADE,
    site_id INT REFERENCES sites(id) ON DELETE CASCADE,
    visit_id BIGINT REFERENCES visits(id) ON DELETE CASCADE,
    mrn_id BIGINT REFERENCES mrns(id) ON DELETE CASCADE,
    ordering_physician_id BIGINT REFERENCES physicians(id) ON DELETE CASCADE,
    arrival TIMESTAMP,
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
