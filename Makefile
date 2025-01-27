SCHEMA_DIR := sql/schema
CONN_STR := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"

up:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} up

down:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} down

status:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} status

reset:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} reset

.PHONY: up down status