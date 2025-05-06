-- +goose Up
ALTER TABLE IF EXISTS exams ADD COLUMN ordering_physician_id BIGINT REFERENCES physicians(id) ON DELETE CASCADE;
CREATE INDEX exams_ordering_physician_id ON exams(ordering_physician_id ASC);

DROP INDEX IF EXISTS exams_order_id;
ALTER TABLE IF EXISTS exams DROP COLUMN order_id;

-- +goose Down
ALTER TABLE IF EXISTS exams ADD COLUMN order_id BIGINT REFERENCES orders(id) ON DELETE CASCADE;
CREATE INDEX exams_order_id ON exams(order_id ASC);

DROP INDEX IF EXISTS exams_ordering_physician_id;
ALTER TABLE IF EXISTS exams DROP COLUMN IF EXISTS ordering_physician_id;
