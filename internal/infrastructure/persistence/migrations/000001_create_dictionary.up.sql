CREATE TABLE
    IF NOT EXISTS dictionary (
        id VARCHAR(36) PRIMARY KEY,
        slug VARCHAR(255) NOT NULL UNIQUE,
        title VARCHAR(500) NOT NULL,
        cover_url VARCHAR(1000) NOT NULL DEFAULT '',
        sources JSON NOT NULL,
        source_stats JSON NOT NULL,
        best_source JSON NOT NULL,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        INDEX idx_slug (slug),
        INDEX idx_title (title)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;