CREATE TABLE
    IF NOT EXISTS manga (
        id VARCHAR(36) PRIMARY KEY,
        dictionary_id VARCHAR(36) NOT NULL UNIQUE,
        title VARCHAR(500) NOT NULL DEFAULT '',
        description TEXT NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT '',
        authors JSON NOT NULL,
        genres JSON NOT NULL,
        cover_url VARCHAR(1000) NOT NULL DEFAULT '',
        state VARCHAR(20) NOT NULL DEFAULT 'unavailable',
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        INDEX idx_dictionary_id (dictionary_id),
        INDEX idx_state (state),
        INDEX idx_title (title)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;