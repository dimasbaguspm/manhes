# manhes

> **For educational purposes only.** This project is not intended for commercial use. See the [license notice](#license) for details.

A self-hosted manga library. Manhes scrapes metadata and chapters from multiple sources, stores images in S3-compatible object storage, and serves them through a clean web reader — all from a single binary.

**Sources:** MangaDex, Atsu.moe (extensible via `Scraper` interface)

---

## Preview

<video src="https://github.com/user-attachments/assets/5814d153-ad77-4b49-9bac-043d2608941f" autoplay loop muted playsinline width="100%"></video>

---

## Features

- **Multi-source ingestion** — scrapes in parallel, picks the best source per language by chapter count
- **Event-driven pipeline** — Kafka/Redpanda decouples scraping, downloading, and S3 upload
- **Web reader** — vertical strip layout, auto-scroll, zoom, header auto-hide, progress bar
- **Watchlist** — add a manga once, background daemons keep it up to date
- **Single binary** — React SPA embedded into the Go binary via `//go:embed`

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25, Chi, MySQL 8 |
| Queue | Kafka / Redpanda |
| Storage | MinIO (S3-compatible) |
| Frontend | React 18, TypeScript, Vite, Tailwind CSS |
| API docs | Swagger / OpenAPI |

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- Go 1.25+ (for local development)
- Node.js 22+ (for frontend development)

### Quick start — nightly image

The fastest way to run manhes locally. No build step required.

```sh
cp .env.example .env        # review and fill in your values
make staging-up             # pulls nightly image + starts Redpanda + MinIO
```

The app will be available at `http://localhost:8080`.

```sh
make staging-down           # tear everything down and wipe volumes
```

### Local development

```sh
make init           # install Go tools (air, swag) + npm dependencies
make dev-backend    # start infra + Go API with hot-reload (localhost:8080)
make dev-web        # second terminal — Vite dev server (localhost:5173)
```

The Vite dev server proxies `/api` to `localhost:8080`.

```sh
make dev-reset      # wipe Docker volumes (including MySQL data) and library files
```

### Build from source

```sh
make prod-build     # → bin/manhes  (frontend embedded, no CGO)
```

The binary serves the React SPA at `/` and the REST API at `/api/v1/`. Wire up Redpanda, MinIO, and a reverse proxy separately.

---

## Configuration

All config is via environment variables (`.env` file or injected at runtime). Copy `.env.example` to get started.

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` | `manhes` | MySQL user |
| `DB_PASS` | `manhes` | MySQL password |
| `DB_NAME` | `manhes` | MySQL database name |
| `DB_MAX_OPEN_CONNS` | `25` | Max open connections in the pool |
| `DB_MAX_IDLE_CONNS` | `5` | Max idle connections in the pool |
| `INGEST_WORKERS` | `5` | Max parallel ingest jobs |
| `INGEST_INTERVAL` | `30m` | How often the watchlist daemon re-publishes ingest events |
| `SYNC_INTERVAL` | `15m` | How often the sync daemon scans for missed chapters |
| `DOWNLOADER_TIMEOUT` | `30s` | HTTP timeout for page/cover downloads |
| `DICTIONARY_REFRESH_INTERVAL` | `4h` | How often source stats are refreshed |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated broker list |
| `S3_ENDPOINT` | `minio:9000` | MinIO/S3 endpoint |
| `S3_BUCKET` | `manga` | Bucket name |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | `minioadmin` | Credentials |
| `S3_USE_SSL` | `false` | Use HTTPS for S3 connection |
| `S3_PUBLIC_ENDPOINT` | `localhost:9000` | Public base URL for image URLs served to the frontend |
| `MANGADEX_BASE_URL` | `https://api.mangadex.org` | |
| `MANGADEX_RATE_LIMIT` | `5` | Requests/sec |
| `ATSU_BASE_URL` | `https://atsu.moe` | |
| `ATSU_RATE_LIMIT` | `3` | Requests/sec |

---

## Published images

Images are published to the GitHub Container Registry on every release and nightly:

```
ghcr.io/dimasbaguspm/manhes:latest       # latest stable release
ghcr.io/dimasbaguspm/manhes:v1.2.3       # pinned version
ghcr.io/dimasbaguspm/manhes:nightly      # nightly build from main
```

`infra/docker-compose.prod.yml` uses the `nightly` tag by default. To pin to a stable release, create an override file:

```yaml
# docker-compose.override.yml
services:
  manhes:
    image: ghcr.io/dimasbaguspm/manhes:latest
```

```sh
docker compose -f infra/docker-compose.prod.yml -f docker-compose.override.yml --env-file .env up -d
```

---

## API

Swagger UI: `http://localhost:8080/swagger/index.html`

```
GET    /api/v1/manga                          List catalog (paginated, filterable)
GET    /api/v1/manga/{mangaId}                Manga detail
GET    /api/v1/manga/{mangaId}/{lang}         Uploaded chapters for a language
GET    /api/v1/manga/{mangaId}/{lang}/read    Chapter pages (with prev/next links)
GET    /api/v1/dictionary?q=...              Search all sources + upsert into dictionary
POST   /api/v1/watchlist                     Add manga to watchlist by dictionary ID
```

### Workflow

```sh
# 1. Search for a manga — stores results in the dictionary
curl "http://localhost:8080/api/v1/dictionary?q=tower+of+god"

# 2. Add to watchlist — triggers ingestion in the background
curl -X POST http://localhost:8080/api/v1/watchlist \
  -H 'Content-Type: application/json' \
  -d '{"dictionaryId": "<uuid>"}'

# 3. Read once state becomes "available"
curl "http://localhost:8080/api/v1/manga/<mangaId>/en/read?chapter=1"
```

---

## Adding a new source

1. Create `internal/infrastructure/scraper/{name}/adapter.go` implementing `domain.Scraper` (and optionally `domain.Searcher`)
2. Register it in `buildScraperRegistry` in `cmd/manhes/main.go` with a priority value
3. Add `BaseURL` / `RateLimit` fields to `config/config.go` and `.env.example`

---

## Contributing

Contributions are welcome. Please open an issue first to discuss significant changes.

This project follows **Trunk Based Development** — all work is merged directly into `main` via short-lived branches. Keep branches small and focused; avoid long-running feature branches.

```sh
make test     # go test ./...
make lint     # go vet ./...
make swagger  # regenerate OpenAPI docs after handler changes
```

Keep pull requests focused — one concern per PR. Follow the existing layer boundaries: domain has no external deps, infrastructure implements ports, handlers are HTTP-only.

---

## License

MIT License — Copyright (c) 2026 dimasbaguspm

> **Non-commercial use only.** This project is built for personal and educational purposes. It is not intended to be used commercially or to infringe upon the rights of manga authors, publishers, or distributors. Always support the official releases.
