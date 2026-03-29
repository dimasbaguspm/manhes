-- dictionary.sql: dictionary table

-- name: UpsertDictionary :exec
INSERT INTO dictionary (id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
ON DUPLICATE KEY UPDATE
    title=VALUES(title),
    sources=JSON_MERGE_PATCH(dictionary.sources, VALUES(sources)),
    source_stats=CASE WHEN VALUES(source_stats) != '{}' THEN VALUES(source_stats) ELSE dictionary.source_stats END,
    best_source=CASE WHEN VALUES(best_source) != '{}' THEN VALUES(best_source) ELSE dictionary.best_source END,
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE dictionary.cover_url END,
    updated_at=CURRENT_TIMESTAMP;

-- name: GetDictionary :one
SELECT id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at
FROM dictionary WHERE id = ?;

-- name: GetDictionaryBySlug :one
SELECT id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at
FROM dictionary WHERE slug = ?;

-- name: GetDictionariesByIDs :many
SELECT id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at
FROM dictionary WHERE id IN (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListDictionary :many
WITH filtered AS (
    SELECT id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at
    FROM dictionary
    WHERE (? = '' OR title LIKE CONCAT('%', ?, '%'))
),
counted AS (SELECT COUNT(*) AS total FROM filtered)
SELECT f.id, f.slug, f.title, f.sources, f.source_stats, f.best_source, f.cover_url, f.updated_at, f.created_at, c.total
FROM filtered f, counted c
ORDER BY f.title
LIMIT ? OFFSET ?;
