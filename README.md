# manhes

A self-hosted manga library. Scrapes metadata from MangaDex and Atsu.moe, stores images in S3, and serves them through a web reader — single binary, no CGO.

> **For educational purposes only.** See [license](#license).

---

## Quick start

```sh
cp .env.example .env
make staging-up   # starts app at http://localhost:8080
make staging-down # teardown
```

## Local development

```sh
make init         # install tools + deps
make dev-backend  # Go API with hot-reload (localhost:8080)
make dev-web      # Vite dev server (localhost:5173, proxies /api)
make dev-reset    # wipe volumes and library files
```

## Build

```sh
make prod-build   # → bin/manhes
```

---

## API

Swagger UI: `http://localhost:8080/swagger/index.html`

```
GET  /api/v1/manga                     List catalog (paginated, filterable)
GET  /api/v1/manga/{mangaId}           Manga detail
GET  /api/v1/manga/{mangaId}/{lang}    Chapters for a language
GET  /api/v1/read/{chapterId}          Chapter pages
GET  /api/v1/dictionary?q=...          Search sources
POST /api/v1/dictionary/refresh         Refresh a dictionary entry
```

---

## Architecture

Event-driven pipeline: `POST /dictionary/refresh` → scrapes sources → downloads chapters → uploads to S3 → serves via web reader.

- **Background daemon** — `IngestDaemon` runs every 4h (configurable via `DICTIONARY_REFRESH_INTERVAL`), refreshes available/fetching manga and cleans up orphaned disk dirs.
- **Event bus** — in-memory pub/sub via goroutines, no external broker.
- **Scrapers** — MangaDex (priority 1) and Atsu.moe, pluggable via `Scraper` interface.

### Layer rules

- `domain/` — pure Go, zero external imports. Defines interfaces.
- `handler/` — HTTP only, depends on domain interfaces.
- `infrastructure/` — implements domain ports, never imported by `domain/`.

---

## Environment variables

See `.env.example`. Key ones:

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | MySQL host |
| `DB_PORT` | `3306` | MySQL port |
| `DB_USER` / `DB_PASS` | `manhes` | Credentials |
| `DB_NAME` | `manhes` | Database name |
| `S3_ENDPOINT` | `minio:9000` | MinIO/S3 endpoint |
| `S3_BUCKET` | `manga` | Bucket name |
| `DICTIONARY_REFRESH_INTERVAL` | `4h` | Daemon refresh interval |

---

## Adding a source

1. Create `internal/infrastructure/scraper/{name}/adapter.go` implementing `domain.Scraper`
2. Register in `BuildScraperRegistry` in `cmd/manhes/wiring.go` with a priority value
3. Add `BaseURL` / `RateLimit` fields to `config/` and `.env.example`

---

## License

MIT — Copyright (c) 2026 dimasbaguspm

> **Non-commercial use only.** For personal and educational purposes only.
