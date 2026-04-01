CREATE TABLE IF NOT EXISTS tracker (
    id VARCHAR(36) PRIMARY KEY,
    manga_id VARCHAR(36) NOT NULL,
    chapter_id VARCHAR(36) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSON NOT NULL DEFAULT ('{}'),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_tracker_manga_chapter (manga_id, chapter_id),
    CONSTRAINT fk_tracker_manga FOREIGN KEY (manga_id) REFERENCES manga (id) ON DELETE CASCADE,
    CONSTRAINT fk_tracker_chapter FOREIGN KEY (chapter_id) REFERENCES chapters (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
