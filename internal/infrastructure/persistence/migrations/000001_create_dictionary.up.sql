-- 000001_create_dictionary.up.sql
CREATE TABLE dictionary (
    id           VARCHAR(36) PRIMARY KEY,
    slug         VARCHAR(255) NOT NULL UNIQUE,
    title        VARCHAR(500) NOT NULL,
    sources      JSON NOT NULL DEFAULT ('{}'),
    source_stats JSON NOT NULL DEFAULT ('{}'),
    best_source  JSON NOT NULL DEFAULT ('{}'),
    state        VARCHAR(20) NOT NULL DEFAULT 'unavailable',
    cover_url    VARCHAR(1000) NOT NULL DEFAULT '',
    refreshed_at TIMESTAMP NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_dictionary_state (state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
