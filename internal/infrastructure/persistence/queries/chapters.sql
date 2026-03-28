-- chapters.sql: unified chapters table

-- name: UpsertChapter :exec
INSERT INTO chapters (id, manga_id, name, chapter_order, lang, image_src)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    name=VALUES(name), chapter_order=VALUES(chapter_order),
    lang=VALUES(lang), image_src=VALUES(image_src);

-- name: GetChapterCountByLang :one
SELECT COUNT(*) FROM chapters WHERE manga_id = ? AND lang = ?;

-- name: GetChaptersByLang :many
SELECT id, manga_id, name, chapter_order, lang, image_src
FROM chapters WHERE manga_id = ? AND lang = ? ORDER BY chapter_order ASC;

-- name: GetChaptersByManga :many
SELECT id, manga_id, name, chapter_order, lang, image_src
FROM chapters WHERE manga_id = ? ORDER BY lang, chapter_order ASC;

-- name: IsChapterIngested :one
SELECT COUNT(*) FROM chapters WHERE manga_id = ? AND lang = ? AND chapter_order = ?;
