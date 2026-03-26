package persistence

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"manga-engine/internal/domain"
)

func (r *SQLiteRepository) AddWatchlist(entry domain.WatchlistEntry) error {
	sourcesJSON, _ := json.Marshal(entry.Sources)
	id := entry.ID
	if id == "" {
		id = uuid.New().String()
	}
	_, err := r.db.Exec(`
		INSERT INTO watchlist (id, slug, title, sources, dictionary_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title=excluded.title, sources=excluded.sources,
			dictionary_id=excluded.dictionary_id`,
		id, entry.Slug, entry.Title, string(sourcesJSON), nullStr(entry.DictionaryID),
	)
	return err
}

func (r *SQLiteRepository) RemoveWatchlist(slug string) error {
	_, err := r.db.Exec(`DELETE FROM watchlist WHERE slug = ?`, slug)
	return err
}

func (r *SQLiteRepository) ListWatchlist() ([]domain.WatchlistEntry, error) {
	rows, err := r.db.Query(
		`SELECT COALESCE(id,''), slug, title, sources, last_checked_at, COALESCE(dictionary_id,'') FROM watchlist ORDER BY slug`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWatchlistRows(rows)
}

func (r *SQLiteRepository) UpdateLastChecked(slug string, t time.Time) error {
	_, err := r.db.Exec(
		`UPDATE watchlist SET last_checked_at = ? WHERE slug = ?`,
		t.UTC().Format(time.RFC3339), slug,
	)
	return err
}

func scanWatchlistRows(rows *sql.Rows) ([]domain.WatchlistEntry, error) {
	var entries []domain.WatchlistEntry
	for rows.Next() {
		var e domain.WatchlistEntry
		var sourcesJSON, dictID, id string
		var lastChecked sql.NullString
		if err := rows.Scan(&id, &e.Slug, &e.Title, &sourcesJSON, &lastChecked, &dictID); err != nil {
			return nil, err
		}
		e.ID = id
		e.DictionaryID = dictID
		json.Unmarshal([]byte(sourcesJSON), &e.Sources)
		if lastChecked.Valid {
			t, _ := time.Parse(time.RFC3339, lastChecked.String)
			e.LastChecked = &t
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
