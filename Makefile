include .env
export

.PHONY: run build start test vet migrate migrate-down lint fmt migrate-create migrate-up migrate-down migrate-status migrate-reset

## DB CONFIG
DB_DRIVER=postgres
MIGRATIONS_DIR=migrations
DB_STRING=host=${DB_HOST} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=${DB_SSLMODE} port=${DB_PORT}

## SERVER OPERATION COMMANDS
run:
	go run ./cmd/api

build: 
	go build -o bin/api/main ./cmd/api

start: build
	./bin/api/main

air:
	air

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run


## GOOSE DB MIGRATIONS SETUP
## Create a new migration: make migrate-create name=create_users_table
migrate-create:
ifndef name
	$(error name is not set. Usage: make migrate-create name=create_users_table)
endif
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

## Apply all pending migrations
migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" up

## Roll back the last migration
migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" down

## Apply only the next pending migration
migrate-up-by-one:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" up-by-one

## Redo the last migration (down then up)
migrate-redo:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" redo

## Roll back all migrations then reapply
migrate-reset:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" reset

## Show current migration status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_STRING)" status