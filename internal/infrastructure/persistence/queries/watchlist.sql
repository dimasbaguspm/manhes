-- watchlist.sql: watchlist table

-- name: AddWatchlist :exec
INSERT INTO watchlist (id, slug, title, sources, dictionary_id)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    title=VALUES(title), sources=VALUES(sources),
    dictionary_id=VALUES(dictionary_id);

-- name: RemoveWatchlist :exec
DELETE FROM watchlist WHERE slug = ?;

-- name: ListWatchlist :many
SELECT COALESCE(id, ''), slug, title, sources, last_checked_at, COALESCE(dictionary_id, '') FROM watchlist ORDER BY slug;

-- name: UpdateLastChecked :exec
UPDATE watchlist SET last_checked_at = ? WHERE slug = ?;
