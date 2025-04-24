-- +goose Up
DROP INDEX IF EXISTS reports_body_idx;

-- +goose Down
CREATE INDEX reports_body_idx ON reports(body ASC);
