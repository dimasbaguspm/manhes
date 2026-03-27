-- Dictionary: cross-source manga index (source of truth for /manga endpoints)
CREATE TABLE IF NOT EXISTS dictionary (
    id           TEXT PRIMARY KEY,
    slug         TEXT UNIQUE NOT NULL,
    title        TEXT NOT NULL,
    sources      TEXT NOT NULL DEFAULT '{}',       -- JSON map[string]string
    source_stats TEXT NOT NULL DEFAULT '{}',       -- JSON map[string]SourceStat
    best_source  TEXT NOT NULL DEFAULT '{}',       -- JSON map[string]string (lang→source)
    state        TEXT NOT NULL DEFAULT 'unavailable',
    refreshed_at TEXT,
    created_at   TEXT NOT NULL
);

-- Watchlist: registered manga titles with polling schedule
CREATE TABLE IF NOT EXISTS watchlist (
    slug            TEXT PRIMARY KEY,
    title           TEXT NOT NULL DEFAULT '',
    interval        INT  NOT NULL DEFAULT 6,
    period          TEXT NOT NULL DEFAULT 'hour',
    last_checked_at DATETIME,
    sources         TEXT NOT NULL DEFAULT '{}'  -- JSON map[string]string
);

-- Ingest chapter state: tracks what has been downloaded
CREATE TABLE IF NOT EXISTS ingest_chapters (
    slug          TEXT NOT NULL,
    language      TEXT NOT NULL,
    chapter_num   TEXT NOT NULL,
    sort_key      REAL NOT NULL DEFAULT 0,
    downloaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (slug, language, chapter_num)
);

-- Catalog: read model for API responses
CREATE TABLE IF NOT EXISTS manga (
    slug        TEXT PRIMARY KEY,
    uuid        TEXT NOT NULL DEFAULT '',
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT '',
    authors     TEXT NOT NULL DEFAULT '[]',  -- JSON array
    genres      TEXT NOT NULL DEFAULT '[]',  -- JSON array
    cover_url   TEXT NOT NULL DEFAULT '',
    synced_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_manga_uuid ON manga(uuid) WHERE uuid != '';

CREATE TABLE IF NOT EXISTS manga_langs (
    slug       TEXT NOT NULL REFERENCES manga(slug) ON DELETE CASCADE,
    language   TEXT NOT NULL,
    available  INT  NOT NULL DEFAULT 0,
    downloaded INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (slug, language)
);

CREATE TABLE IF NOT EXISTS manga_chapters (
    slug        TEXT    NOT NULL,
    language    TEXT    NOT NULL,
    chapter_num TEXT    NOT NULL,
    sort_key    REAL    NOT NULL DEFAULT 0,
    page_count  INT     NOT NULL DEFAULT 0,
    uploaded    BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TEXT,
    PRIMARY KEY (slug, language, chapter_num),
    FOREIGN KEY (slug, language) REFERENCES manga_langs(slug, language) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS chapter_pages (
    slug        TEXT NOT NULL,
    language    TEXT NOT NULL,
    chapter_num TEXT NOT NULL,
    page_index  INT  NOT NULL,
    s3_url      TEXT NOT NULL,
    PRIMARY KEY (slug, language, chapter_num, page_index),
    FOREIGN KEY (slug, language, chapter_num)
        REFERENCES manga_chapters(slug, language, chapter_num) ON DELETE CASCADE
);
