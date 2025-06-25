-- +goose Up
ALTER TABLE IF EXISTS procedures ADD COLUMN updated_by TEXT NOT NULL DEFAULT '';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION procedures_set_updated_fields()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.specialty IS DISTINCT FROM OLD.specialty
    OR NEW.modality IS DISTINCT FROM OLD.modality THEN
      NEW.updated_by := current_user;
      NEW.updated_at := CURRENT_TIMESTAMP;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_on_new_proc_specialty
BEFORE UPDATE ON procedures
FOR EACH ROW
EXECUTE FUNCTION procedures_set_updated_fields();

-- +goose Down
DROP TRIGGER IF EXISTS update_on_new_proc_specialty ON procedures;
DROP FUNCTION IF EXISTS procedures_set_updated_fields();
ALTER TABLE IF EXISTS procedures DROP COLUMN updated_by;
