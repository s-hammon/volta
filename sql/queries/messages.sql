-- name: CreateMessage :one
INSERT INTO messages (
    field_separator,
    encoding_characters,
    sending_application,
    sending_facility,
    receiving_application,
    receiving_facility,
    received_at,
    message_type,
    trigger_event,
    control_id,
    processing_id,
    version_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetMessageByID :one
SELECT *
FROM messages
WHERE id = $1;
