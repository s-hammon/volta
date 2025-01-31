PROJECT_NAME := volta
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
SCHEMA_DIR := sql/schema
CONN_STR := ${CONN_STR}

build:
	@GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/${PROJECT_NAME} cmd/service/main.go

clean: reset up
	@rm -rf bin
	@go mod tidy

up:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} up

down:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} down

status:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} status

reset:
	@goose -dir ${SCHEMA_DIR} postgres ${CONN_STR} reset

test:
	@go test -cover ./...

artifact: build
	@docker build -t ${PROJECT_NAME} .

.PHONY: build clean up down status reset test artifact