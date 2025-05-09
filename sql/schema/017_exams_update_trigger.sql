-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_current_status_cm()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.end_exam_dt IS NOT NULL THEN
    NEW.current_status := 'CM';
  ELSIF (TG_OP = 'UPDATE' AND OLD.end_exam_dt IS NOT NULL) THEN
    NEW.current_status := 'CM';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER ensure_cm_on_end_dt
BEFORE UPDATE ON exams
FOR EACH ROW
EXECUTE FUNCTION set_current_status_cm();

-- +goose Down
DROP TRIGGER IF EXISTS ensure_cm_on_end_dt ON exams;
DROP FUNCTION IF EXISTS set_current_status_cm();
