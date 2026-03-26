package persistence

import (
	"database/sql"
	"encoding/json"
	"time"

	"manga-engine/internal/domain"
)

func (r *SQLiteRepository) UpsertDictionary(entry domain.DictionaryEntry) error {
	sourcesJSON, _ := json.Marshal(entry.Sources)
	statsJSON, _ := json.Marshal(entry.SourceStats)
	bestJSON, _ := json.Marshal(entry.BestSource)
	createdAt := entry.CreatedAt.UTC().Format(time.RFC3339)
	if createdAt == (time.Time{}).UTC().Format(time.RFC3339) {
		createdAt = time.Now().UTC().Format(time.RFC3339)
	}
	var refreshedAt *string
	if entry.RefreshedAt != nil {
		s := entry.RefreshedAt.UTC().Format(time.RFC3339)
		refreshedAt = &s
	}
	_, err := r.db.Exec(`
		INSERT INTO dictionary (id, slug, title, sources, source_stats, best_source, state, cover_url, refreshed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title=excluded.title,
			sources=json_patch(dictionary.sources, excluded.sources),
			source_stats=CASE WHEN excluded.source_stats != '{}' THEN excluded.source_stats ELSE dictionary.source_stats END,
			best_source=CASE WHEN excluded.best_source != '{}' THEN excluded.best_source ELSE dictionary.best_source END,
			state=CASE WHEN dictionary.state = 'available' THEN 'available' ELSE excluded.state END,
			cover_url=CASE WHEN excluded.cover_url != '' THEN excluded.cover_url ELSE dictionary.cover_url END,
			refreshed_at=CASE WHEN excluded.refreshed_at IS NOT NULL THEN excluded.refreshed_at ELSE dictionary.refreshed_at END`,
		entry.ID, entry.Slug, entry.Title,
		string(sourcesJSON), string(statsJSON), string(bestJSON),
		string(entry.State), entry.CoverURL, refreshedAt, createdAt,
	)
	return err
}

func (r *SQLiteRepository) GetDictionary(id string) (domain.DictionaryEntry, bool, error) {
	return r.scanDictionaryRow(r.db.QueryRow(
		`SELECT id, slug, title, sources, source_stats, best_source, state, cover_url, refreshed_at, created_at
		 FROM dictionary WHERE id = ?`, id,
	))
}

func (r *SQLiteRepository) GetDictionaryBySlug(slug string) (domain.DictionaryEntry, bool, error) {
	return r.scanDictionaryRow(r.db.QueryRow(
		`SELECT id, slug, title, sources, source_stats, best_source, state, cover_url, refreshed_at, created_at
		 FROM dictionary WHERE slug = ?`, slug,
	))
}

func (r *SQLiteRepository) ListDictionary(filter domain.DictionaryFilter) (domain.DictionaryPage, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	const q = `
		WITH filtered AS (
			SELECT id, slug, title, sources, source_stats, best_source, state, cover_url, refreshed_at, created_at
			FROM dictionary
			WHERE (? = '' OR title LIKE '%' || ? || '%')
		), counted AS (SELECT COUNT(*) AS total FROM filtered)
		SELECT f.id, f.slug, f.title, f.sources, f.source_stats, f.best_source,
		       f.state, f.cover_url, f.refreshed_at, f.created_at, c.total
		FROM filtered f, counted c
		ORDER BY f.title
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(q, filter.Q, filter.Q, pageSize, offset)
	if err != nil {
		return domain.DictionaryPage{}, err
	}
	defer rows.Close()

	var entries []domain.DictionaryEntry
	var total int
	for rows.Next() {
		var e domain.DictionaryEntry
		var sourcesJSON, statsJSON, bestJSON, stateStr, createdAt string
		var refreshedAt sql.NullString
		if err := rows.Scan(
			&e.ID, &e.Slug, &e.Title, &sourcesJSON, &statsJSON, &bestJSON,
			&stateStr, &e.CoverURL, &refreshedAt, &createdAt, &total,
		); err != nil {
			return domain.DictionaryPage{}, err
		}
		if err := unmarshalDictEntry(&e, sourcesJSON, statsJSON, bestJSON, stateStr, refreshedAt, createdAt); err != nil {
			return domain.DictionaryPage{}, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return domain.DictionaryPage{}, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	return domain.DictionaryPage{
		Items:      entries,
		TotalItems: total,
		TotalPages: totalPages,
		PageSize:   pageSize,
		PageNumber: page,
	}, nil
}

func (r *SQLiteRepository) SetDictionaryState(id string, state domain.MangaState) error {
	_, err := r.db.Exec(`UPDATE dictionary SET state = ? WHERE id = ?`, string(state), id)
	return err
}

func (r *SQLiteRepository) SetDictionaryStateBySlug(slug string, state domain.MangaState) error {
	_, err := r.db.Exec(`UPDATE dictionary SET state = ? WHERE slug = ?`, string(state), slug)
	return err
}

func (r *SQLiteRepository) scanDictionaryRow(row *sql.Row) (domain.DictionaryEntry, bool, error) {
	var e domain.DictionaryEntry
	var sourcesJSON, statsJSON, bestJSON, stateStr, createdAt string
	var refreshedAt sql.NullString
	err := row.Scan(&e.ID, &e.Slug, &e.Title, &sourcesJSON, &statsJSON, &bestJSON, &stateStr, &e.CoverURL, &refreshedAt, &createdAt)
	if err == sql.ErrNoRows {
		return domain.DictionaryEntry{}, false, nil
	}
	if err != nil {
		return domain.DictionaryEntry{}, false, err
	}
	if err := unmarshalDictEntry(&e, sourcesJSON, statsJSON, bestJSON, stateStr, refreshedAt, createdAt); err != nil {
		return domain.DictionaryEntry{}, false, err
	}
	return e, true, nil
}

func unmarshalDictEntry(e *domain.DictionaryEntry, sourcesJSON, statsJSON, bestJSON, stateStr string, refreshedAt sql.NullString, createdAt string) error {
	json.Unmarshal([]byte(sourcesJSON), &e.Sources)
	json.Unmarshal([]byte(statsJSON), &e.SourceStats)
	json.Unmarshal([]byte(bestJSON), &e.BestSource)
	e.State = domain.MangaState(stateStr)
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		e.CreatedAt = t
	}
	if refreshedAt.Valid {
		if t, err := time.Parse(time.RFC3339, refreshedAt.String); err == nil {
			e.RefreshedAt = &t
		}
	}
	if e.Sources == nil {
		e.Sources = map[string]string{}
	}
	if e.SourceStats == nil {
		e.SourceStats = map[string]domain.SourceStat{}
	}
	if e.BestSource == nil {
		e.BestSource = map[string]string{}
	}
	return nil
}
