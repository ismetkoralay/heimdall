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

up:
	docker compose up --build -d

down:
	docker compose down

down-clean:
	docker compose down -v
