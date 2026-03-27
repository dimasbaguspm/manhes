package persistence

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"manga-engine/internal/domain"
)

func (r *SQLiteRepository) UpsertManga(m domain.Manga) error {
	authors, _ := json.Marshal(m.Authors)
	genres, _ := json.Marshal(m.Genres)
	_, err := r.db.Exec(`
		INSERT INTO manga (uuid, slug, title, description, status, authors, genres, cover_url, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(slug) DO UPDATE SET
			title=excluded.title, description=excluded.description,
			status=excluded.status, authors=excluded.authors, genres=excluded.genres,
			cover_url=CASE WHEN excluded.cover_url != '' THEN excluded.cover_url ELSE cover_url END,
			synced_at=CURRENT_TIMESTAMP`,
		uuid.New().String(), m.Slug, m.Title, m.Description, m.Status, string(authors), string(genres), m.CoverURL,
	)
	return err
}

func (r *SQLiteRepository) ListManga(filter domain.MangaFilter) (domain.MangaPage, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "title"
	}

	const q = `
		WITH filtered AS (
			SELECT d.id AS dict_id, d.slug, d.state,
			       COALESCE(m.title, d.title) AS title,
			       COALESCE(m.status, '') AS status,
			       COALESCE(m.description, '') AS description,
			       COALESCE(NULLIF(m.cover_url, ''), NULLIF(d.cover_url, ''), '') AS cover_url,
			       COALESCE(m.authors, '[]') AS authors,
			       COALESCE(m.genres, '[]') AS genres,
			       COALESCE(m.synced_at, d.created_at) AS synced_at,
			       COALESCE((SELECT json_group_array(language) FROM manga_langs WHERE slug = d.slug ORDER BY language), '[]') AS languages,
			       COALESCE((SELECT json_group_object(language, available) FROM manga_langs WHERE slug = d.slug), '{}') AS chapters_by_lang
			FROM dictionary d
			LEFT JOIN manga m ON m.slug = d.slug
			WHERE (? = '' OR d.title LIKE '%' || ? || '%')
			  AND (? = '' OR d.state = ?)
			  AND (? = '' OR m.status = ?)
			  AND (? = FALSE OR d.state != 'unavailable')
		), counted AS (SELECT COUNT(*) AS total FROM filtered)
		SELECT f.dict_id, f.slug, f.state, f.title, f.status, f.description,
		       f.cover_url, f.authors, f.genres, f.synced_at, f.languages, f.chapters_by_lang, c.total
		FROM filtered f, counted c
		ORDER BY CASE WHEN ? = 'last_update' THEN f.synced_at END DESC,
		         CASE WHEN ? != 'last_update' THEN f.title END ASC
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(q,
		filter.Title, filter.Title,
		filter.State, filter.State,
		filter.Status, filter.Status,
		filter.HideUnavailable,
		sortBy, sortBy,
		pageSize, offset,
	)
	if err != nil {
		return domain.MangaPage{}, err
	}
	defer rows.Close()

	var items []domain.Manga
	var total int
	for rows.Next() {
		var m domain.Manga
		var authorsJSON, genresJSON, langsJSON, chapsByLangJSON string
		var state string
		var syncedAt sql.NullString
		if err := rows.Scan(
			&m.DictionaryID, &m.Slug, &state,
			&m.Title, &m.Status, &m.Description, &m.CoverURL,
			&authorsJSON, &genresJSON, &syncedAt, &langsJSON, &chapsByLangJSON, &total,
		); err != nil {
			return domain.MangaPage{}, err
		}
		m.State = domain.MangaState(state)
		json.Unmarshal([]byte(authorsJSON), &m.Authors)
		json.Unmarshal([]byte(genresJSON), &m.Genres)
		json.Unmarshal([]byte(langsJSON), &m.Languages)
		json.Unmarshal([]byte(chapsByLangJSON), &m.ChaptersByLang)
		if syncedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", syncedAt.String); err == nil {
				m.UpdatedAt = &t
			}
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return domain.MangaPage{}, err
	}
	return domain.MangaPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *SQLiteRepository) GetMangaBySlug(slug string) (domain.MangaDetail, bool, error) {
	var d domain.MangaDetail
	var authorsJSON, genresJSON string
	err := r.db.QueryRow(
		`SELECT slug, title, status, description, authors, genres, cover_url FROM manga WHERE slug = ?`, slug,
	).Scan(&d.Slug, &d.Title, &d.Status, &d.Description, &authorsJSON, &genresJSON, &d.CoverURL)
	if err == sql.ErrNoRows {
		return domain.MangaDetail{}, false, nil
	}
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	json.Unmarshal([]byte(authorsJSON), &d.Authors)
	json.Unmarshal([]byte(genresJSON), &d.Genres)
	detail, err := r.fillMangaDetail(d)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}
	return detail, true, nil
}

func (r *SQLiteRepository) fillMangaDetail(d domain.MangaDetail) (domain.MangaDetail, error) {
	langRows, err := r.db.Query(
		`SELECT language, available, downloaded FROM manga_langs WHERE slug = ? ORDER BY language`, d.Slug,
	)
	if err != nil {
		return domain.MangaDetail{}, err
	}
	defer langRows.Close()
	for langRows.Next() {
		var l domain.MangaLang
		if err := langRows.Scan(&l.Language, &l.Available, &l.Fetched); err != nil {
			return domain.MangaDetail{}, err
		}
		d.Languages = append(d.Languages, l)
	}

	chRows, err := r.db.Query(
		`SELECT slug, language, chapter_num, page_count, uploaded, uploaded_at
		 FROM manga_chapters WHERE slug = ? ORDER BY language, sort_key`, d.Slug,
	)
	if err != nil {
		return domain.MangaDetail{}, err
	}
	defer chRows.Close()
	for chRows.Next() {
		var ch domain.MangaChapter
		var uploadedAt sql.NullString
		if err := chRows.Scan(&ch.Slug, &ch.Language, &ch.ChapterNum, &ch.PageCount, &ch.Uploaded, &uploadedAt); err != nil {
			return domain.MangaDetail{}, err
		}
		if uploadedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
				ch.UploadedAt = &t
			}
		}
		d.Chapters = append(d.Chapters, ch)
	}

	return d, nil
}

func (r *SQLiteRepository) HasUploadedChapters(slug string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM manga_chapters WHERE slug = ? AND uploaded = TRUE`, slug,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteRepository) HasPendingChapters(slug string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM manga_chapters WHERE slug = ? AND uploaded = FALSE`, slug,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLiteRepository) UpsertLang(slug, lang string, available, downloaded int) error {
	_, err := r.db.Exec(`
		INSERT INTO manga_langs (slug, language, available, downloaded)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(slug, language) DO UPDATE SET
			available=excluded.available, downloaded=excluded.downloaded`,
		slug, lang, available, downloaded,
	)
	return err
}

func (r *SQLiteRepository) UpsertChapter(slug, lang, num string, sortKey float64, pageCount int) error {
	_, err := r.db.Exec(`
		INSERT INTO manga_chapters (slug, language, chapter_num, sort_key, page_count)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(slug, language, chapter_num) DO UPDATE SET sort_key=excluded.sort_key, page_count=excluded.page_count`,
		slug, lang, num, sortKey, pageCount,
	)
	return err
}

func (r *SQLiteRepository) MarkChapterUploaded(slug, lang, num string) error {
	_, err := r.db.Exec(
		`UPDATE manga_chapters SET uploaded=TRUE, uploaded_at=CURRENT_TIMESTAMP WHERE slug=? AND language=? AND chapter_num=?`,
		slug, lang, num,
	)
	return err
}

func (r *SQLiteRepository) UpsertPage(slug, lang, num string, idx int, url string) error {
	_, err := r.db.Exec(`
		INSERT INTO chapter_pages (slug, language, chapter_num, page_index, s3_url)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(slug, language, chapter_num, page_index) DO UPDATE SET s3_url=excluded.s3_url`,
		slug, lang, num, idx, url,
	)
	return err
}

func (r *SQLiteRepository) GetChapterPages(slug, lang, num string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT s3_url FROM chapter_pages WHERE slug=? AND language=? AND chapter_num=? ORDER BY page_index`,
		slug, lang, num,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, rows.Err()
}

func (r *SQLiteRepository) GetChaptersByLang(slug, lang string) ([]domain.MangaChapter, error) {
	rows, err := r.db.Query(
		`SELECT slug, language, chapter_num, page_count, uploaded_at
		 FROM manga_chapters WHERE slug=? AND language=? AND uploaded=TRUE ORDER BY sort_key ASC`,
		slug, lang,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chapters []domain.MangaChapter
	for rows.Next() {
		var ch domain.MangaChapter
		var uploadedAt sql.NullString
		if err := rows.Scan(&ch.Slug, &ch.Language, &ch.ChapterNum, &ch.PageCount, &uploadedAt); err != nil {
			return nil, err
		}
		ch.Uploaded = true
		if uploadedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", uploadedAt.String); err == nil {
				ch.UploadedAt = &t
			}
		}
		chapters = append(chapters, ch)
	}
	return chapters, rows.Err()
}

func (r *SQLiteRepository) GetPendingChapters() ([]domain.ChapterRef, error) {
	rows, err := r.db.Query(
		`SELECT mc.slug, mc.language, mc.chapter_num, COALESCE(d.id, '')
		 FROM manga_chapters mc
		 LEFT JOIN dictionary d ON d.slug = mc.slug
		 WHERE mc.uploaded=FALSE ORDER BY mc.slug, mc.language, mc.sort_key`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.ChapterRef
	for rows.Next() {
		var ref domain.ChapterRef
		if err := rows.Scan(&ref.Slug, &ref.Language, &ref.ChapterNum, &ref.DictionaryID); err != nil {
			return nil, err
		}
		result = append(result, ref)
	}
	return result, rows.Err()
}
