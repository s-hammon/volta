-- +goose Up
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    field_separator CHAR NOT NULL,
    encoding_characters TEXT NOT NULL,
    sending_application TEXT NOT NULL,
    sending_facility TEXT NOT NULL,
    receiving_application TEXT NOT NULL,
    receiving_facility TEXT NOT NULL,
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    message_type TEXT NOT NULL,
    trigger_event TEXT NOT NULL,
    control_id TEXT NOT NULL,
    processing_id TEXT NOT NULL,
    version_id TEXT NOT NULL
);

CREATE INDEX messages_control_id_idx ON messages(control_id ASC);
CREATE INDEX messages_message_type_idx ON messages(message_type ASC);
CREATE INDEX messages_processing_id_idx ON messages(processing_id ASC);
CREATE INDEX messages_received_at_idx ON messages(received_at ASC);
CREATE INDEX messages_sending_facility_idx ON messages(sending_facility ASC);

-- +goose Down
DROP TABLE IF EXISTS messages;