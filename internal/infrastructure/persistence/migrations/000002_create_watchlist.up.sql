-- 000002_create_watchlist.up.sql
CREATE TABLE watchlist (
    slug            VARCHAR(255) PRIMARY KEY,
    title           VARCHAR(500) NOT NULL DEFAULT '',
    interval_hours  INT NOT NULL DEFAULT 6,
    period          VARCHAR(20) NOT NULL DEFAULT 'hour',
    last_checked_at TIMESTAMP NULL,
    sources         JSON NOT NULL DEFAULT ('{}'),
    dictionary_id   VARCHAR(36) NULL,
    id              VARCHAR(36) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
