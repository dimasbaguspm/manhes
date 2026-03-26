.PHONY: init \
        dev-backend dev-web dev-reset \
        staging-up staging-down \
        prod-build \
        test lint swagger

## Setup — install Go tools and frontend dependencies
init:
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go mod download all
	cd web && npm install
	mkdir -p library

## Development
# Start infra (Redpanda + MinIO) and the Go API with hot-reload.
dev-backend:
	docker compose -f infra/docker-compose.dev.yml --env-file .env up --build

# Start the Vite dev server (proxies /api → localhost:8080).
dev-web:
	cd web && npm run gen:types && npm run dev

# Wipe all local state: Docker volumes, SQLite db, downloaded library files.
dev-reset:
	docker compose -f infra/docker-compose.dev.yml --env-file .env down -v
	docker run --rm -v "$(PWD)/library:/library" alpine sh -c "rm -rf /library/*"
	rm -f manhes.db manhes.db-shm manhes.db-wal

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
