-- 000013_create_watchlist_index.up.sql
CREATE INDEX idx_watchlist_last_checked ON watchlist(last_checked_at);
