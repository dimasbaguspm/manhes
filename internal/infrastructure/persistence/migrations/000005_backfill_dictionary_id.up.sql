-- 000005_backfill_dictionary_id.up.sql
ALTER TABLE watchlist
    ADD CONSTRAINT fk_watchlist_dictionary
    FOREIGN KEY (dictionary_id) REFERENCES dictionary(id) ON DELETE SET NULL;
