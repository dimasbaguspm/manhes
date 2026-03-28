-- 000018_restructure_manga.up.sql
-- Restructure manga table: drop slug, add id (UUID) and dictionary_id, add chapters_by_lang.
DROP TABLE IF EXISTS manga;

CREATE TABLE manga (
    id              VARCHAR(36) PRIMARY KEY,
    dictionary_id   VARCHAR(36) NOT NULL UNIQUE,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    status          VARCHAR(50) NOT NULL DEFAULT '',
    authors         JSON NOT NULL DEFAULT ('[]'),
    genres          JSON NOT NULL DEFAULT ('[]'),
    cover_url       VARCHAR(1000) NOT NULL DEFAULT '',
    state           VARCHAR(20) NOT NULL DEFAULT 'unavailable',
    chapters_by_lang JSON NOT NULL DEFAULT ('{}'),
    updated_at      TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_dictionary_id (dictionary_id),
    INDEX idx_state (state),
    INDEX idx_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
