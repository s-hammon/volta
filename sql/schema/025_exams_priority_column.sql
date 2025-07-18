-- +goose Up
ALTER TABLE exams ADD COLUMN priority TEXT;

-- +goose Down
ALTER TABLE exams DROP COLUMN priority;
