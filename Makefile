# Use Docker if installed, otherwise Podman (override with: make COMPOSE="docker compose" ...)
COMPOSE ?= $(shell command -v docker >/dev/null 2>&1 && echo "docker compose" || echo "podman compose")
ENGINE ?= $(shell command -v docker >/dev/null 2>&1 && echo docker || echo podman)

.PHONY: run docker-up docker-down docker-app docker-https docker-logs backup backup-list restore build build-web

build-web:
	cd web && npm install && npm run build

run: build-web
	go run cmd/main.go

build: build-web
	go build -o bin/server cmd/main.go

# MongoDB only, for local development with `make run`
docker-up:
	$(COMPOSE) up -d mongodb

# Full app (MongoDB + app) in Docker on http://localhost:8080.
# Rebuilds the image and always replaces the app container. The container is
# removed by name because podman-compose ignores --force-recreate in some
# cases (e.g. a crash-looping container keeps its old config).
docker-app:
	$(COMPOSE) up -d mongodb backup
	$(COMPOSE) build app
	-$(ENGINE) rm -f syairullah-app
	$(COMPOSE) up -d --no-deps app

# Production with HTTPS via Caddy; requires DOMAIN in .env
# App is only published on 127.0.0.1 here, so all traffic goes through Caddy.
docker-https:
	APP_BIND=127.0.0.1 $(MAKE) docker-app
	$(COMPOSE) --profile https up -d caddy

# Backups (stored in ./backups, or BACKUP_DIR from .env)
backup:
	$(COMPOSE) exec backup bash /backup.sh now

backup-list:
	$(COMPOSE) exec backup bash /backup.sh list

# Restore REPLACES the current database and uploads: make restore BACKUP=backup-YYYYmmdd-HHMMSS
restore:
	@test -n "$(BACKUP)" || (echo "pakai: make restore BACKUP=backup-YYYYmmdd-HHMMSS (lihat: make backup-list)"; exit 1)
	@read -p "Restore $(BACKUP)? Data sekarang akan DIGANTI. Ketik 'ya': " ok && [ "$$ok" = "ya" ]
	$(COMPOSE) stop app
	$(COMPOSE) --profile tools run --rm restore $(BACKUP)
	$(COMPOSE) start app

docker-logs:
	$(COMPOSE) logs -f app

docker-down:
	$(COMPOSE) --profile https down
