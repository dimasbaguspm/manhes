-- name: UpsertTracker :exec
INSERT INTO tracker (id, manga_id, chapter_id, is_read, metadata)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE is_read = VALUES(is_read), metadata = VALUES(metadata);

-- name: GetTracker :one
SELECT id, manga_id, chapter_id, is_read, metadata, updated_at, created_at
FROM tracker WHERE manga_id = ? AND chapter_id = ?;

-- name: GetTrackersByManga :many
SELECT id, manga_id, chapter_id, is_read, metadata, updated_at, created_at
FROM tracker WHERE manga_id = ?;

-- name: ListTracker :many
SELECT id, manga_id, chapter_id, is_read, metadata, updated_at, created_at
FROM tracker;
