.PHONY: run build test lint tidy

run:
	go run ./cmd/heimdall

build:
	CGO_ENABLED=0 go build -o bin/service ./cmd/heimdall/

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

