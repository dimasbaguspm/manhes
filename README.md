# manhes

Manga scraper and ingestion engine with a web reader. Manages a watchlist, scrapes metadata and chapters from multiple sources in parallel, uploads to S3, and exposes a REST catalog API consumed by a React frontend.

**Sources:** MangaDex, Atsu.moe (extensible via `Scraper` interface)
**Storage:** SQLite (catalog) + MinIO/S3 (images)
**Backend stack:** Go 1.25, Chi, Kafka/Redpanda, no CGO (`modernc.org/sqlite`)
**Frontend stack:** React 18, TypeScript, Vite, Tailwind CSS, React Router v6

---

## Development

Copy `.env.example` to `.env`, then:

```sh
make init           # install Go tools (air, swag) + frontend npm dependencies
make dev-backend    # start infra (Redpanda + MinIO) + Go API with hot-reload
make dev-web        # second terminal — generate types + start Vite (http://localhost:5173)
```

The Vite dev server proxies `/api` to `localhost:8080`, so both run side by side without touching each other.

```sh
make dev-reset   # wipe Docker volumes, SQLite db, and library files (clean slate)
```

### Preview (nightly)

Run the latest nightly image with the full stack

```sh
cp .env.example .env   # fill in your values once
make staging-up        # pulls ghcr.io/dimasbaguspm/manhes:nightly + starts Redpanda + MinIO
make staging-down      # tear everything down and wipe volumes
```

`infra/docker-compose.prod.yml` uses the `nightly` image by default, so this is always up to date with the latest main-branch build.

### Production

Produces a single binary with the frontend embedded:

```sh
make prod-build  # → bin/manhes
```

No infra included — wire up Redpanda, MinIO, and a reverse proxy however you prefer. The binary serves the React SPA at `/` and the REST API at `/api/v1/`.

#### Pinning to a stable release

Images are published to the GitHub Container Registry on every release:

```
ghcr.io/dimasbaguspm/manhes:latest       # latest stable release
ghcr.io/dimasbaguspm/manhes:v1.2.3       # pinned version
ghcr.io/dimasbaguspm/manhes:nightly      # nightly build from main
```

To pin to a specific version, create an override file next to `docker-compose.prod.yml`:

```yaml
# docker-compose.override.yml
services:
  manhes:
    image: ghcr.io/dimasbaguspm/manhes:latest
```

Then bring it up:

```sh
docker compose -f infra/docker-compose.prod.yml -f docker-compose.override.yml --env-file .env up -d
```

All configuration is provided via `.env` — see the [Configuration](#configuration) table below. The container exposes port `8080`; put a reverse proxy in front of it as needed.

### Utilities

```sh
make test     # go test ./...
make lint     # go vet ./...
make swagger  # regenerate OpenAPI docs (re-run after handler changes)
```


---

## Configuration

All config is via environment variables (`.env` file or injected at runtime):

| Variable | Default | Description |
|----------|---------|-------------|
| `INGEST_WORKERS` | `5` | Max parallel ingest jobs |
| `SYNC_INTERVAL` | `15m` | How often the sync daemon scans for missed chapters |
| `DOWNLOADER_TIMEOUT` | `30s` | HTTP timeout for page/cover downloads |
| `DICTIONARY_REFRESH_INTERVAL` | `4h` | How often the dictionary daemon refreshes source stats |
| `KAFKA_BROKERS` | `localhost:9092` | Comma-separated broker list |
| `S3_ENDPOINT` | `minio:9000` | MinIO/S3 endpoint (internal hostname in compose) |
| `S3_BUCKET` | `manga` | |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | `minioadmin` | |
| `S3_USE_SSL` | `false` | Use HTTPS for S3 connection |
| `S3_PUBLIC_ENDPOINT` | `localhost:9000` | Public base URL served to the frontend for image URLs |
| `MANGADEX_BASE_URL` | `https://api.mangadex.org` | |
| `MANGADEX_RATE_LIMIT` | `5` | Requests/sec |
| `ATSU_BASE_URL` | `https://atsu.moe` | |
| `ATSU_RATE_LIMIT` | `3` | Requests/sec |

---

## API

Swagger UI: `http://localhost:8080/swagger/index.html`

```
# Catalog
GET    /api/v1/manga                              List catalog (paginated, filterable)
GET    /api/v1/manga/{mangaId}                    Manga detail (mangaId is dictionary UUID)
GET    /api/v1/manga/{mangaId}/{lang}             Uploaded chapters for a language
GET    /api/v1/manga/{mangaId}/{lang}/read        Chapter pages (with prev/next links)

# Dictionary — cross-source index
GET    /api/v1/dictionary?q=...                   Search all sources + upsert results into dictionary

# Watchlist
POST   /api/v1/watchlist                          Add manga to watchlist by dictionary ID
```

### Workflow

**1. Discover a manga** via search — results are stored in the dictionary:
```sh
curl "http://localhost:8080/api/v1/dictionary?q=tower+of+god"
# Returns a list of DictionaryEntry objects with their IDs
```

**2. Add to watchlist** using the dictionary entry ID:
```sh
curl -X POST http://localhost:8080/api/v1/watchlist \
  -H 'Content-Type: application/json' \
  -d '{"dictionaryId": "<uuid>"}'
# Returns: {"status":"accepted","slug":"...","dictionaryId":"..."}
```

**3. Read chapters** once ingestion completes (state becomes `available`):
```sh
# List uploaded chapters
curl "http://localhost:8080/api/v1/manga/<mangaId>/en"

# Read a chapter
curl "http://localhost:8080/api/v1/manga/<mangaId>/en/read?chapter=1"
```

---

## Event flow

```
WatchlistDaemon → Kafka (manga.ingest)
  → IngestConsumer (up to INGEST_WORKERS parallel)
    → scrape sources in parallel, download chapters to disk
    → Kafka (manga.sync) per chapter
      → SyncConsumer → upload pages to S3, mark uploaded, delete local files
SyncDaemon → periodic full scan (SYNC_INTERVAL) for any missed chapters

DictionaryDaemon → periodic refresh of source stats (DICTIONARY_REFRESH_INTERVAL)
```

Ingest jobs run concurrently up to `INGEST_WORKERS`. Per-source rate limits are enforced by each adapter's `rate.Limiter` — concurrent jobs to the same source self-throttle.

---

## Library layout

```
library/{slug}/
├── metadata.json
├── cover.jpg              # deleted after S3 upload
└── {lang}/
    ├── metadata.json
    └── ch-{###}/
        ├── chapter.json
        └── {###}.jpg      # deleted after S3 upload
```

JSON metadata files are kept as source-of-truth for re-sync. Image files are removed once successfully uploaded to S3.

---

## Adding a new source

1. Create `internal/infrastructure/scraper/{name}/adapter.go` implementing `domain.Scraper` (and optionally `domain.Searcher`)
2. Register it in `buildScraperRegistry` in `cmd/manhes/main.go` with a priority value
3. Add `BaseURL` / `RateLimit` config fields to `config/config.go` and `.env.example`
