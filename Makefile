.PHONY: run docker-up docker-down build build-web

build-web:
	cd web && npm install && npm run build

run: build-web
	go run cmd/main.go

build: build-web
	go build -o bin/server cmd/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down
