-- chapters.sql: unified chapters table

-- name: GetChapterUploaded :one
-- Returns the image_src if the chapter has been uploaded (image_src IS NOT NULL AND image_src != ''), empty string otherwise.
SELECT image_src FROM chapters WHERE manga_id = ? AND lang = ? AND name = ? AND image_src IS NOT NULL AND image_src != '' LIMIT 1;

-- name: UpsertChapter :exec
INSERT INTO chapters (id, manga_id, name, chapter_order, lang, image_src, page_urls, page_count)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    name=VALUES(name), chapter_order=VALUES(chapter_order),
    lang=VALUES(lang), image_src=VALUES(image_src),
    page_urls=VALUES(page_urls), page_count=VALUES(page_count);

-- name: GetChapterCountByLang :one
SELECT COUNT(*) FROM chapters WHERE manga_id = ? AND lang = ?;

-- name: GetChaptersByLang :many
SELECT id, manga_id, name, chapter_order, lang, image_src
FROM chapters WHERE manga_id = ? AND lang = ? ORDER BY chapter_order ASC;

-- name: GetUploadedChaptersByLang :many
SELECT id, manga_id, name, chapter_order, lang, image_src, page_urls, page_count, updated_at
FROM chapters WHERE manga_id = ? AND lang = ? AND image_src IS NOT NULL AND image_src != '' ORDER BY chapter_order ASC;

-- name: GetChaptersByManga :many
SELECT id, manga_id, name, chapter_order, lang, image_src
FROM chapters WHERE manga_id = ? ORDER BY lang, chapter_order ASC;

-- name: IsChapterIngested :one
SELECT COUNT(*) FROM chapters WHERE manga_id = ? AND lang = ? AND chapter_order = ?;

-- name: GetChapterByID :one
SELECT id, manga_id, name, chapter_order, lang, image_src, page_urls, page_count, updated_at
FROM chapters
WHERE id = ?
LIMIT 1;
