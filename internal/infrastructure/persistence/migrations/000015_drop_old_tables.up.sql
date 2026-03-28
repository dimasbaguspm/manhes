-- 000015_drop_old_tables.up.sql
-- Drop tables replaced by the new chapters table.
DROP TABLE IF EXISTS chapter_pages;
DROP TABLE IF EXISTS manga_chapters;
DROP TABLE IF EXISTS manga_langs;
DROP TABLE IF EXISTS ingest_chapters;
