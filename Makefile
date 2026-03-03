.PHONY: run docker-up docker-down build

run:
	go run cmd/main.go

build:
	go build -o bin/server cmd/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down
