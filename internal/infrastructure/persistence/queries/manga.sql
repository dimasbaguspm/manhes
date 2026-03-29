-- manga.sql: manga table queries

-- name: UpsertManga :exec
INSERT INTO manga (id, dictionary_id, title, description, status, authors, genres, cover_url, state)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    title=VALUES(title),
    description=VALUES(description),
    status=VALUES(status),
    authors=VALUES(authors),
    genres=VALUES(genres),
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE cover_url END,
    state=VALUES(state);

-- name: GetMangaByDictionaryID :one
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.updated_at, m.created_at
FROM manga m
WHERE m.dictionary_id = ?;

-- name: GetMangaByID :one
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.updated_at, m.created_at
FROM manga m
WHERE m.id = ?;

-- name: GetMangaChapterStats :many
-- Returns per-language chapter stats for a manga, computed via JOIN on chapters table.
SELECT
    c.lang,
    COUNT(*) AS total,
    SUM(CASE WHEN c.image_src IS NOT NULL AND c.image_src != '' THEN 1 ELSE 0 END) AS available,
    MAX(c.updated_at) AS latest_updated
FROM chapters c
WHERE c.manga_id = ?
GROUP BY c.lang;

-- name: ListManga :many
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.updated_at, m.created_at,
       COUNT(*) OVER () AS total
FROM manga m
WHERE (? = '' OR m.dictionary_id = ?)
  AND (? = '' OR m.title LIKE CONCAT('%', ?, '%') OR m.description LIKE CONCAT('%', ?, '%'))
  AND (? = '' OR m.state = ?)
ORDER BY
    CASE WHEN ? = 'title' AND ? = 'asc' THEN m.title END ASC,
    CASE WHEN ? = 'title' AND ? = 'desc' THEN m.title END DESC,
    CASE WHEN ? = 'updatedAt' AND ? = 'asc' THEN m.updated_at END ASC,
    CASE WHEN ? = 'updatedAt' AND ? = 'desc' THEN m.updated_at END DESC,
    CASE WHEN ? = 'createdAt' AND ? = 'asc' THEN m.created_at END ASC,
    CASE WHEN ? = 'createdAt' AND ? = 'desc' THEN m.created_at END DESC
LIMIT ? OFFSET ?;
