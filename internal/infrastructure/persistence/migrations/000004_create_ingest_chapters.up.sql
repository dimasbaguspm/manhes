-- 000004_create_ingest_chapters.up.sql
CREATE TABLE ingest_chapters (
    slug          VARCHAR(255) NOT NULL,
    language      VARCHAR(10) NOT NULL,
    chapter_num   VARCHAR(50) NOT NULL,
    sort_key      DOUBLE NOT NULL DEFAULT 0,
    downloaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (slug, language, chapter_num),
    INDEX idx_ingest_chapters_slug_lang (slug, language)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
