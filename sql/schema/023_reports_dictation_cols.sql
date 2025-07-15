-- +goose Up
ALTER TABLE reports ADD COLUMN dictation_start TIMESTAMP;
ALTER TABLE reports ADD COLUMN dictation_end TIMESTAMP;

-- +goose Down
ALTER TABLE reports DROP COLUMN dictation_end;
ALTER TABLE reports DROP COLUMN dictation_start;
