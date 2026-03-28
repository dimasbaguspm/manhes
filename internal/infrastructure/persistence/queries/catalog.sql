-- catalog.sql: manga table only (chapter data moved to chapters table)

-- name: UpsertManga :exec
INSERT INTO manga (uuid, slug, title, description, status, authors, genres, cover_url, synced_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON DUPLICATE KEY UPDATE
    title=VALUES(title), description=VALUES(description),
    status=VALUES(status), authors=VALUES(authors), genres=VALUES(genres),
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE cover_url END,
    synced_at=CURRENT_TIMESTAMP;

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
counted AS (SELECT COUNT(*) AS total FROM filtered),
langs AS (SELECT manga_id, JSON_ARRAYAGG(lang) AS langs FROM (SELECT DISTINCT manga_id, lang FROM chapters) t GROUP BY manga_id)
SELECT f.dict_id, f.slug, f.state, f.title, f.status, f.description,
       f.cover_url, f.authors, f.genres, f.synced_at,
       COALESCE(l.langs, '[]') AS languages,
       c.total
FROM filtered f, counted c
LEFT JOIN langs l ON l.manga_id = f.slug
ORDER BY CASE WHEN ? = 'last_update' THEN f.synced_at END DESC,
         CASE WHEN ? != 'last_update' THEN f.title END ASC
LIMIT ? OFFSET ?;

-- name: GetMangaBySlug :one
SELECT m.slug, m.title, m.status, m.description, m.authors, m.genres, m.cover_url,
       d.id AS dict_id, d.state
FROM manga m
LEFT JOIN dictionary d ON d.slug = m.slug
WHERE m.slug = ?;
