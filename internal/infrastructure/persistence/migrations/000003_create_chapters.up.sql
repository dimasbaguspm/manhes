CREATE TABLE
    IF NOT EXISTS chapters (
        id VARCHAR(36) PRIMARY KEY,
        manga_id VARCHAR(36) NOT NULL,
        name VARCHAR(500) NOT NULL DEFAULT '',
        chapter_order INT NOT NULL DEFAULT 0,
        lang VARCHAR(10) NOT NULL,
        image_src VARCHAR(1000) NOT NULL DEFAULT '',
        page_urls JSON NOT NULL DEFAULT ('[]'),
        page_count INT NOT NULL DEFAULT 0,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        INDEX idx_chapters_manga_lang (manga_id, lang),
        INDEX idx_chapters_manga_order (manga_id, chapter_order),
        CONSTRAINT fk_chapters_manga FOREIGN KEY (manga_id) REFERENCES manga (id) ON DELETE CASCADE
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
