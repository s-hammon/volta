-- name: CreateOrder :one
INSERT INTO orders (
    outside_system_id,
    site_id,
    visit_id,
    mrn_id,
    ordering_physician_id,
    arrival,
    number,
    current_status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetOrderBySiteIDNumber :one
SELECT *
FROM orders
WHERE
    site_id = $1
    AND number = $2;