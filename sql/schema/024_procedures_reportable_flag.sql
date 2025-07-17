-- +goose Up
ALTER TABLE procedures ADD COLUMN reportable BOOLEAN DEFAULT TRUE;

-- +goose Down
ALTER TABLE procedures DROP COLUMN reportable;
