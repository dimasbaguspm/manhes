-- 000003_create_manga.up.sql
CREATE TABLE manga (
    slug        VARCHAR(255) PRIMARY KEY,
    uuid        VARCHAR(36) NOT NULL DEFAULT '',
    title       VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT '',
    authors     JSON NOT NULL,
    genres      JSON NOT NULL,
    cover_url   VARCHAR(1000) NOT NULL DEFAULT '',
    synced_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_manga_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
