-- 000012_modify_watchlist_id.up.sql
ALTER TABLE watchlist MODIFY id VARCHAR(36) NOT NULL;
