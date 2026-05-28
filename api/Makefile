.PHONY: build run dev test fmt migrate-up migrate-down docker-up docker-down

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	air

test:
	go test ./...

fmt:
	gofmt -w .
	goimports -w .

migrate-up:
	go run ./cmd/server -migrate-up

migrate-down:
	go run ./cmd/server -migrate-down

docker-up:
	docker compose up --build

docker-down:
	docker compose down
