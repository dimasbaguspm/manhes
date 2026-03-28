-- 000008_create_manga_chapters.up.sql
CREATE TABLE manga_chapters (
    slug        VARCHAR(255) NOT NULL,
    language    VARCHAR(10) NOT NULL,
    chapter_num VARCHAR(50) NOT NULL,
    sort_key    DOUBLE NOT NULL DEFAULT 0,
    page_count  INT NOT NULL DEFAULT 0,
    uploaded    BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMP NULL,
    PRIMARY KEY (slug, language, chapter_num),
    CONSTRAINT fk_manga_chapters_lang FOREIGN KEY (slug, language) REFERENCES manga_langs(slug, language) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
