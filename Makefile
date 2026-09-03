.PHONY: run build start test vet migrate migrate-down lint fmt

run:
	go run ./cmd/api

build: 
	go build -o bin/api/main ./cmd/api

start: build
	./bin/api/main

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run