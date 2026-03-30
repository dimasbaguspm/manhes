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
- **Event-driven pipeline** — an in-memory event bus decouples retrieval, file upload, and S3 storage
- **Web reader** — vertical strip layout, auto-scroll, zoom, header auto-hide, progress bar
- **Background sync** — add a manga once, background daemon keeps it up to date automatically
- **Single binary** — React SPA embedded into the Go binary via `//go:embed`

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
make staging-up             # pulls nightly image + starts MinIO
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

The binary serves the React SPA at `/` and the REST API at `/api/v1/`. Wire up MinIO and a reverse proxy separately.

---

## Codebase Overview

```
manhes/
├── cmd/manhes/           # Application entry point and dependency wiring
│   ├── main.go           # main() — bootstraps everything
│   ├── wiring.go         # Builds and connects all components
│   └── resource.go        # Embedded React SPA (//go:embed)
│
├── config/               # Configuration from environment variables
│   ├── config.go         # Main Config struct
│   ├── database.go       # MySQL connection settings
│   ├── s3.go             # S3/MinIO connection settings
│   ├── sources.go        # Per-source (MangaDex, Atsu) settings
│   ├── bus.go            # Event bus channel names
│   └── env.go            # Env var parsing helpers
│
├── internal/
│   ├── domain/           # Pure domain types — no external dependencies
│   │   ├── manga.go      # Manga aggregate root
│   │   ├── chapter.go    # Chapter entity
│   │   ├── dictionary.go # DictionaryEntry (cross-source index)
│   │   ├── events.go     # All event types (DictionaryUpdated, etc.)
│   │   ├── ports.go      # Repository & object store interfaces
│   │   ├── scraper.go    # Scraper port interface
│   │   ├── subscriber.go # Subscriber interface definition
│   │   └── handler.go    # Handler interfaces (HTTP layer contracts)
│   │
│   ├── handler/          # HTTP request handlers (Chi)
│   │   ├── handler.go    # Handlers struct + constructor
│   │   ├── manga.go      # ListManga, GetManga, GetChaptersByLang
│   │   ├── chapter.go    # ReadChapter (with prev/next nav)
│   │   ├── dictionary.go # Search, Refresh (triggers ingest)
│   │   └── helpers.go    # View-model converters (toMangaSummary, etc.)
│   │
│   ├── subscriber/       # Event consumers — react to events from the bus
│   │   ├── dictionary.go # DictionarySubscriber: handles DictionaryRefreshed
│   │   ├── manga.go      # MangaSubscriber: handles DictionaryUpdated, ChapterUploaded, MangaAvailable
│   │   ├── retrieval.go  # RetrievalSubscriber: handles IngestRequested
│   │   └── file_upload.go# FileUploadSubscriber: handles ChaptersFound
│   │
│   ├── daemon/           # Background workers
│   │   └── ingest.go     # IngestDaemon: periodic dictionary refresh + orphan cleanup
│   │
│   ├── ui/               # Embedded frontend (React SPA)
│   │   └── ui.go        # Serves the embedded web app
│   │
│   └── infrastructure/  # Adapters implementing domain ports
│       ├── persistence/ # MySQL adapter
│       │   ├── mysql.go  # DB connection + implements domain.Repository
│       │   └── queries/  # Raw SQL grouped by entity (manga, chapter, dictionary)
│       ├── scraper/     # Source adapters (MangaDex, Atsu)
│       │   ├── registry.go      # Priority-based scraper registry
│       │   ├── circuit_breaker.go
│       │   ├── mangadex/        # MangaDex scraper (priority 1)
│       │   └── atsu/            # Atsu.moe scraper (priority 2)
│       ├── storage/   # Disk storage (temp files before S3 upload)
│       ├── s3/        # S3/MinIO adapter
│       ├── eventbus/  # In-memory event bus (pub/sub via goroutines)
│       ├── http/      # HTTP setup (router, middleware, CORS)
│       └── downloader/ # HTTP client for downloading pages/covers
│
└── pkg/                 # Shared utilities (no internal deps)
    ├── eventbus/        # Bus interface + InMemBus implementation
    ├── log/             # Structured logging (slog wrapper)
    ├── retry/           # Retry with backoff
    ├── concurrent/      # concurrent.Collect (fan-in helper)
    ├── circuitbreaker/ # Circuit breaker pattern
    ├── httputil/        # HTTP client + response helpers
    ├── reqctx/          # Request ID middleware
    └── lifecycle/       # Startup/shutdown helpers
```

**Layer rules:**
- `domain/` — pure Go, zero external imports. Defines interfaces (ports) that infrastructure implements.
- `handler/` — HTTP only. Depends on domain interfaces, never on infrastructure directly.
- `subscriber/` — event consumers. Depend on domain interfaces to do work.
- `infrastructure/` — implements the interfaces defined in `domain/`. Never imported by `domain/`.

---

## Architecture

The entire ingest pipeline is event-driven. When you call `POST /api/v1/dictionary/refresh`, it kicks off a chain of async work — no blocking, no polling:

```
POST /dictionary/refresh
  │
  ▼
DictionaryRefreshed  ──► DictionarySubscriber updates source stats + best source
                              │
                              ▼
                         DictionaryUpdated (TriggerIngest=true)
                              │
                              ▼
                         MangaSubscriber upserts manga metadata, sets state=fetching
                              │
                              ▼
                         IngestRequested
                              │
                              ▼
                         RetrievalSubscriber fetches chapter list from all sources, diffs against DB
                              │
                              ▼
                         ChaptersFound (only NEW chapters)
                              │
                              ▼
                         FileUploadSubscriber downloads pages → uploads to S3 → cleans up disk
                              │
                              ▼
                         MangaAvailable (state=available, manga readable)
```

### Subscribers (event consumers)

| Subscriber | Listens to | What it does |
|---|---|---|
| `DictionarySubscriber` | `DictionaryRefreshed` | Re-searches all sources for a manga, updates `SourceStats` and `BestSource`, publishes `DictionaryUpdated` |
| `MangaSubscriber` | `DictionaryUpdated` | Fetches manga metadata (title, cover, authors) from best source, upserts to DB, publishes `IngestRequested` |
| `RetrievalSubscriber` | `IngestRequested` | Fetches chapter lists from all sources concurrently, diffs against stored chapters, publishes `ChaptersFound` for new ones |
| `FileUploadSubscriber` | `ChaptersFound` | Downloads all pages for each new chapter to disk, uploads to S3, cleans up local files, publishes `ChapterUploaded`, then `MangaAvailable` when all done |
| `MangaSubscriber` | `ChapterUploaded` / `MangaAvailable` | Updates manga state (`fetching` → `available`) |

### Background daemon

`IngestDaemon` runs every `DICTIONARY_REFRESH_INTERVAL` (default 4h):
1. Picks manga entries that were previously ingested
2. Calls `DictionaryService.Refresh()` for each — same flow as the manual refresh endpoint
3. Cleans up any orphaned disk directories left behind after S3 migration

### Event bus

The bus (`pkg/eventbus`) is an **in-memory pub/sub** — no Kafka, no Redpanda, no external broker. Events are delivered via goroutines, so handlers run asynchronously. This keeps ops simple (single binary) while preserving the decoupling benefits of an event-driven architecture.

---

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
| `INGEST_CONCURRENCY` | `4` | Max concurrent download/upload workers per chapter job |
| `DOWNLOADER_TIMEOUT` | `30s` | HTTP timeout for page/cover downloads |
| `DICTIONARY_REFRESH_INTERVAL` | `4h` | How often the ingest daemon refreshes manga entries |
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
GET    /api/v1/manga/{mangaId}/{lang}         Chapters for a language
GET    /api/v1/read/{chapterId}              Chapter pages (with prev/next links)
GET    /api/v1/dictionary?q=...               Search all sources, upsert into dictionary
POST   /api/v1/dictionary/refresh             Refresh a dictionary entry (async re-fetch from all sources)
```

### Workflow

```sh
# 1. Search for a manga — stores results in the dictionary
curl "http://localhost:8080/api/v1/dictionary?q=tower+of+god"

# 2. Refresh dictionary entry — triggers ingestion in the background
curl -X POST http://localhost:8080/api/v1/dictionary/refresh \
  -H 'Content-Type: application/json' \
  -d '{"id": "<dictionary-uuid>"}'

# 3. Read once state becomes "available"
curl "http://localhost:8080/api/v1/read/<chapterId>"
```

---

## Adding a new source

1. Create `internal/infrastructure/scraper/{name}/adapter.go` implementing `domain.Scraper` (and optionally `domain.Searcher`)
2. Register it in `BuildScraperRegistry` in `cmd/manhes/wiring.go` with a priority value (lower = higher priority)
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
