-- dictionary_batch.sql: batch upsert for dictionary table

-- name: UpsertDictionaryBatch :exec
INSERT INTO dictionary (id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
ON DUPLICATE KEY UPDATE
    title=VALUES(title),
    sources=JSON_MERGE_PATCH(dictionary.sources, VALUES(sources)),
    source_stats=CASE WHEN VALUES(source_stats) != '{}' THEN VALUES(source_stats) ELSE dictionary.source_stats END,
    best_source=CASE WHEN VALUES(best_source) != '{}' THEN VALUES(best_source) ELSE dictionary.best_source END,
    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE dictionary.cover_url END,
    updated_at=CURRENT_TIMESTAMP;
