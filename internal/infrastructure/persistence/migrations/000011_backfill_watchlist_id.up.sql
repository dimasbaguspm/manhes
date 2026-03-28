-- 000011_backfill_watchlist_id.up.sql
UPDATE watchlist SET id = UUID() WHERE id IS NULL;
