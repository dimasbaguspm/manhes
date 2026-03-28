-- ingest.sql: ingest_chapters table

-- name: IsChapterDownloaded :one
SELECT COUNT(*) FROM ingest_chapters WHERE slug = ? AND language = ? AND chapter_num = ?;

-- name: MarkChapterDownloaded :exec
INSERT INTO ingest_chapters (slug, language, chapter_num, sort_key)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE sort_key=VALUES(sort_key);

-- name: GetDownloadedByLang :many
SELECT language, COUNT(*) FROM ingest_chapters WHERE slug = ? GROUP BY language;

-- name: GetDownloadedChaptersByLang :many
SELECT chapter_num FROM ingest_chapters WHERE slug = ? AND language = ? ORDER BY sort_key ASC;
