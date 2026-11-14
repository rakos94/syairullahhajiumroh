# Use Docker if installed, otherwise Podman (override with: make COMPOSE="docker compose" ...)
COMPOSE ?= $(shell command -v docker >/dev/null 2>&1 && echo "docker compose" || echo "podman compose")

.PHONY: run docker-up docker-down docker-app docker-https docker-logs build build-web

build-web:
	cd web && npm install && npm run build

run: build-web
	go run cmd/main.go

build: build-web
	go build -o bin/server cmd/main.go

# MongoDB only, for local development with `make run`
docker-up:
	$(COMPOSE) up -d mongodb

# Full app (MongoDB + app) in Docker on http://localhost:8080
docker-app:
	$(COMPOSE) up -d --build

# Production with HTTPS via Caddy; requires DOMAIN in .env
docker-https:
	$(COMPOSE) --profile https up -d --build

docker-logs:
	$(COMPOSE) logs -f app

docker-down:
	$(COMPOSE) --profile https down
