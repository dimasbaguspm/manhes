package persistence

import "manga-engine/internal/domain"

func (r *SQLiteRepository) IsChapterDownloaded(slug, lang, num string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM ingest_chapters WHERE slug = ? AND language = ? AND chapter_num = ?`,
		slug, lang, num,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteRepository) MarkChapterDownloaded(slug, lang, num string, sortKey float64) error {
	_, err := r.db.Exec(`
		INSERT INTO ingest_chapters (slug, language, chapter_num, sort_key)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(slug, language, chapter_num) DO NOTHING`,
		slug, lang, num, sortKey,
	)
	return err
}

func (r *SQLiteRepository) GetDownloadedByLang(slug string) (map[string]int, error) {
	rows, err := r.db.Query(
		`SELECT language, COUNT(*) FROM ingest_chapters WHERE slug = ? GROUP BY language`, slug,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int)
	for rows.Next() {
		var lang string
		var count int
		if err := rows.Scan(&lang, &count); err != nil {
			return nil, err
		}
		m[lang] = count
	}
	return m, rows.Err()
}

func (r *SQLiteRepository) GetDownloadedChaptersByLang(slug, lang string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT chapter_num FROM ingest_chapters WHERE slug = ? AND language = ? ORDER BY sort_key ASC`,
		slug, lang,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nums []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, rows.Err()
}

var _ domain.Repository = (*SQLiteRepository)(nil) // compile-time interface check
