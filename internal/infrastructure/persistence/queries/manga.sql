-- manga.sql: manga table queries

-- name: UpsertManga :exec
INSERT INTO manga (id, dictionary_id, title, description, status, authors, genres, cover_url, state, chapters_by_lang, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON DUPLICATE KEY UPDATE
    title=VALUES(title),
    description=VALUES(description),
    status=VALUES(status),
    authors=VALUES(authors),
    genres=VALUES(genres),
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE cover_url END,
    state=VALUES(state),
    chapters_by_lang=VALUES(chapters_by_lang),
    updated_at=CURRENT_TIMESTAMP;

-- name: GetMangaByDictionaryID :one
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.chapters_by_lang, m.updated_at, m.created_at
FROM manga m
WHERE m.dictionary_id = ?;

-- name: GetMangaByID :one
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.chapters_by_lang, m.updated_at, m.created_at
FROM manga m
WHERE m.id = ?;

-- name: ListManga :many
SELECT m.id, m.dictionary_id, m.title, m.description, m.status, m.authors, m.genres,
       m.cover_url, m.state, m.chapters_by_lang, m.updated_at, m.created_at,
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
