-- catalog.sql: manga, manga_langs, manga_chapters, chapter_pages

-- name: UpsertManga :exec
INSERT INTO manga (uuid, slug, title, description, status, authors, genres, cover_url, synced_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON DUPLICATE KEY UPDATE
    title=VALUES(title), description=VALUES(description),
    status=VALUES(status), authors=VALUES(authors), genres=VALUES(genres),
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE cover_url END,
    synced_at=CURRENT_TIMESTAMP;

-- name: UpsertLang :exec
INSERT INTO manga_langs (slug, language, available, downloaded)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE available=VALUES(available), downloaded=VALUES(downloaded);

-- name: UpsertChapter :exec
INSERT INTO manga_chapters (slug, language, chapter_num, sort_key, page_count)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE sort_key=VALUES(sort_key), page_count=VALUES(page_count);

-- name: MarkChapterUploaded :exec
UPDATE manga_chapters SET uploaded=TRUE, uploaded_at=CURRENT_TIMESTAMP WHERE slug=? AND language=? AND chapter_num=?;

-- name: UpsertPage :exec
INSERT INTO chapter_pages (slug, language, chapter_num, page_index, s3_url)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE s3_url=VALUES(s3_url);

-- name: GetChapterPages :many
SELECT s3_url FROM chapter_pages WHERE slug=? AND language=? AND chapter_num=? ORDER BY page_index;

-- name: GetChaptersByLang :many
SELECT slug, language, chapter_num, page_count, uploaded_at
FROM manga_chapters WHERE slug=? AND language=? AND uploaded=TRUE ORDER BY sort_key ASC;

-- name: HasUploadedChapters :one
SELECT COUNT(*) FROM manga_chapters WHERE slug=? AND uploaded=TRUE;

-- name: HasPendingChapters :one
SELECT COUNT(*) FROM manga_chapters WHERE slug=? AND uploaded=FALSE;

-- name: GetPendingChapters :many
SELECT mc.slug, mc.language, mc.chapter_num, COALESCE(d.id, '') AS dictionary_id
FROM manga_chapters mc
LEFT JOIN dictionary d ON d.slug = mc.slug
WHERE mc.uploaded=FALSE
ORDER BY mc.slug, mc.language, mc.sort_key;

-- name: GetMangaLangsBySlug :many
SELECT language, available, downloaded FROM manga_langs WHERE slug=? ORDER BY language;

-- name: GetChaptersBySlug :many
SELECT slug, language, chapter_num, page_count, uploaded, uploaded_at
FROM manga_chapters WHERE slug=? ORDER BY language, sort_key;

-- name: ListManga :many
WITH filtered AS (
    SELECT d.id AS dict_id, d.slug, d.state,
           COALESCE(m.title, d.title) AS title,
           COALESCE(m.status, '') AS status,
           COALESCE(m.description, '') AS description,
           COALESCE(NULLIF(m.cover_url, ''), NULLIF(d.cover_url, ''), '') AS cover_url,
           COALESCE(m.authors, '[]') AS authors,
           COALESCE(m.genres, '[]') AS genres,
           COALESCE(m.synced_at, d.created_at) AS synced_at
    FROM dictionary d
    LEFT JOIN manga m ON m.slug = d.slug
    WHERE (? = '' OR d.title LIKE CONCAT('%', ?, '%'))
      AND (? = '' OR d.state = ?)
      AND (? = '' OR m.status = ?)
      AND (? = FALSE OR d.state != 'unavailable')
),
counted AS (SELECT COUNT(*) AS total FROM filtered)
SELECT f.dict_id, f.slug, f.state, f.title, f.status, f.description,
       f.cover_url, f.authors, f.genres, f.synced_at,
       COALESCE((SELECT JSON_ARRAYAGG(language) FROM manga_langs WHERE slug = f.slug), '[]') AS languages,
       c.total
FROM filtered f, counted c
ORDER BY CASE WHEN ? = 'last_update' THEN f.synced_at END DESC,
         CASE WHEN ? != 'last_update' THEN f.title END ASC
LIMIT ? OFFSET ?;

-- name: GetMangaBySlug :one
SELECT m.slug, m.title, m.status, m.description, m.authors, m.genres, m.cover_url,
       d.id AS dict_id, d.state
FROM manga m
LEFT JOIN dictionary d ON d.slug = m.slug
WHERE m.slug = ?;
