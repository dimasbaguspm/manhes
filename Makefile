.PHONY: init \
        dev-backend dev-web dev-reset \
        staging-up staging-down \
        prod-build \
        test lint swagger sqlc

## Setup — install Go tools and frontend dependencies
init:
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go mod download all
	cd web && npm install
	mkdir -p library

## Development
# Migrations run automatically in-app via golang-migrate. sqlc runs via .air.toml pre_cmd.
dev-backend:
	docker compose -f infra/docker-compose.dev.yml --env-file .env up --build

# Start the Vite dev server (proxies /api → localhost:8080).
dev-web:
	cd web && npm run gen:types && npm run dev

# Wipe all local state: Docker volumes (including MySQL data) and downloaded library files.
dev-reset:
	docker compose -f infra/docker-compose.dev.yml --env-file .env down -v
	docker run --rm -v "$(PWD)/library:/library" alpine sh -c "rm -rf /library/*"

## Staging — production image tested locally (full stack: infra + app)
staging-up:
	docker compose -f infra/docker-compose.prod.yml --env-file .env up --build

staging-down:
	docker compose -f infra/docker-compose.prod.yml --env-file .env down -v

## Production — single binary with the frontend embedded
prod-build:
	cd web && npm run gen:types && npm run build
	cp -r web/dist internal/ui/
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/manhes ./cmd/manhes

## Utilities
test:
	go test ./...

lint:
	go vet ./...

swagger:
	swag init -g ./cmd/manhes/main.go -o ./docs/manhes --packageName manhes --parseInternal --quiet

## Database
MYSQL_DSN ?= $(DB_USER):$(DB_PASS)@tcp(localhost:3306)/$(DB_NAME)?parseTime=true&charset=utf8mb4

sqlc:
	sqlc generate
