PROJECT_NAME := volta
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
SCHEMA_DIR := sql/schema
DATABASE_URL := ${DATABASE_URL}

build:
	@GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/${PROJECT_NAME} cmd/service/main.go

clean: reset up
	@rm -rf bin
	@go mod tidy

up:
	@goose -dir ${SCHEMA_DIR} postgres ${DATABASE_URL} up

down:
	@goose -dir ${SCHEMA_DIR} postgres ${DATABASE_URL} down

status:
	@goose -dir ${SCHEMA_DIR} postgres ${DATABASE_URL} status

reset:
	@goose -dir ${SCHEMA_DIR} postgres ${DATABASE_URL} reset

test:
	@go test -cover ./...

vet:
	@go vet ./...

test-packages:
	go test -json $$(go list ./... | grep -v -e /bin -e /cmd -e /vendor -e /internal/api/models) |\
		tparse --follow -sort=elapsed -trimpath=auto -all

test-packages-short:
	go test -test.short -json $$(go list ./... | grep -v -e /bin -e /cmd -e /vendor -e /internal/api/models) |\
		tparse --follow -sort=elapsed

prod-build: build 
	@scripts/build-prod.sh $(ARGS)

goose-build:
	@scripts/build-goose.sh $(ARGS)

ready: clean 
	go vet ./...
	go test -cover ./...
	golangci-lint run ./...
	gosec -terse ./...

bench-mllp-send: vet
	go test -bench=Send -benchmem -benchtime=5s -cpu=1,2,4 ./pkg/mllp

.PHONY: build clean up down status reset test test-packages test-packages-short artifact ready vet
