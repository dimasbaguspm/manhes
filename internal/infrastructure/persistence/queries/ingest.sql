-- ingest.sql: chapter ingestion queries (replaced ingest_chapters)
-- Chapter existence is now tracked via the chapters table.

-- name: GetDownloadedByLang :many
-- Returns language and count of chapters ingested per language for a manga.
SELECT lang, COUNT(*) AS count FROM chapters WHERE manga_id = ? GROUP BY lang;

-- name: GetDownloadedChaptersByLang :many
-- Returns chapter_order values for chapters of a manga in a given language.
SELECT chapter_order FROM chapters WHERE manga_id = ? AND lang = ? ORDER BY chapter_order ASC;
