package persistence

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/persistence/queries"
)

var _ domain.Repository = (*MySQLRepository)(nil)

//go:embed migrations/*.sql
var fs embed.FS

func NewMySQL(ctx context.Context, cfg config.DatabaseConfig) (*MySQLRepository, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns())
	db.SetMaxIdleConns(cfg.MaxIdleConns())

	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("create migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	repo := &MySQLRepository{
		db: db,
		q:  queries.New(db),
	}

	return repo, nil
}

func (r *MySQLRepository) Close() error { return r.db.Close() }

// MySQLRepository implements domain.Repository via queries.Queries.
type MySQLRepository struct {
	db *sql.DB
	q  *queries.Queries
}

// Manga methods

func (r *MySQLRepository) UpsertManga(ctx context.Context, m domain.Manga) error {
	authors, _ := json.Marshal(m.Authors)
	genres, _ := json.Marshal(m.Genres)
	chaptersByLang, _ := json.Marshal(m.ChaptersByLang)
	state := string(m.State)
	if state == "" {
		state = string(domain.StateUnavailable)
	}
	return r.q.UpsertManga(ctx, queries.UpsertMangaParams{
		ID:             m.ID,
		DictionaryID:   m.DictionaryID,
		Title:          m.Title,
		Description:    m.Description,
		Status:         m.Status,
		Authors:        authors,
		Genres:         genres,
		CoverUrl:       m.CoverURL,
		State:          state,
		ChaptersByLang: chaptersByLang,
	})
}

func (r *MySQLRepository) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
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
	sortOrder := filter.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}

	stateVal := ""
	if len(filter.States) == 1 {
		stateVal = filter.States[0]
	}
	dictIDVal := ""
	if len(filter.IDs) == 1 {
		dictIDVal = filter.IDs[0]
	}

	rows, err := r.q.ListManga(ctx, queries.ListMangaParams{
		Column1:      nil,
		DictionaryID: dictIDVal,
		Column3:      filter.Q,
		CONCAT:       filter.Q,
		CONCAT_2:     filter.Q,
		Column6:      nil,
		State:        stateVal,
		Column8:      sortBy,
		Column9:      sortOrder,
		Column10:     sortBy,
		Column11:     sortOrder,
		Column12:     sortBy,
		Column13:     sortOrder,
		Column14:     sortBy,
		Column15:     sortOrder,
		Column16:     sortBy,
		Column17:     sortOrder,
		Column18:     sortBy,
		Column19:     sortOrder,
		Limit:        int32(pageSize),
		Offset:       int32(offset),
	})
	if err != nil {
		return domain.MangaPage{}, err
	}
	if len(rows) == 0 {
		return domain.MangaPage{Items: []domain.Manga{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	items := make([]domain.Manga, 0, len(rows))
	for _, row := range rows {
		var authors, genres []string
		json.Unmarshal(row.Authors, &authors)
		json.Unmarshal(row.Genres, &genres)
		var chaptersByLang map[string]domain.ChapterStats
		json.Unmarshal(row.ChaptersByLang, &chaptersByLang)
		if chaptersByLang == nil {
			chaptersByLang = map[string]domain.ChapterStats{}
		}

		var updatedAt *time.Time
		if row.UpdatedAt.Valid {
			t := row.UpdatedAt.Time
			updatedAt = &t
		}
		var createdAt time.Time
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time
		}

		items = append(items, domain.Manga{
			ID:             row.ID,
			DictionaryID:   row.DictionaryID,
			Title:          row.Title,
			Description:    row.Description,
			Status:         row.Status,
			CoverURL:       row.CoverUrl,
			Authors:        authors,
			Genres:         genres,
			State:          domain.MangaState(row.State),
			ChaptersByLang: chaptersByLang,
			UpdatedAt:      updatedAt,
			CreatedAt:      createdAt,
		})
	}

	// Apply array filters in-memory
	if len(filter.Genres) > 0 || len(filter.Authors) > 0 || len(filter.States) > 1 {
		items = applyMangaFilters(items, filter)
	}
	if len(filter.IDs) > 1 {
		idSet := map[string]bool{}
		for _, id := range filter.IDs {
			idSet[id] = true
		}
		filtered := make([]domain.Manga, 0, len(items))
		for _, m := range items {
			if idSet[m.DictionaryID] {
				filtered = append(filtered, m)
			}
		}
		items = filtered
	}

	var total int
	if rows[0].Total != nil {
		switch v := rows[0].Total.(type) {
		case int64:
			total = int(v)
		case int:
			total = v
		}
	}

	return domain.MangaPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func applyMangaFilters(items []domain.Manga, filter domain.MangaFilter) []domain.Manga {
	genreSet := map[string]bool{}
	for _, g := range filter.Genres {
		genreSet[g] = true
	}
	authorSet := map[string]bool{}
	for _, a := range filter.Authors {
		authorSet[a] = true
	}
	stateSet := map[string]bool{}
	for _, s := range filter.States {
		stateSet[s] = true
	}

	result := make([]domain.Manga, 0, len(items))
	for _, m := range items {
		if len(genreSet) > 0 {
			match := false
			for _, g := range m.Genres {
				if genreSet[g] {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if len(authorSet) > 0 {
			match := false
			for _, a := range m.Authors {
				if authorSet[a] {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if len(stateSet) > 0 && !stateSet[string(m.State)] {
			continue
		}
		result = append(result, m)
	}
	return result
}

func (r *MySQLRepository) GetMangaByDictionaryID(ctx context.Context, dictionaryID string) (domain.MangaDetail, bool, error) {
	row, err := r.q.GetMangaByDictionaryID(ctx, dictionaryID)
	if err == sql.ErrNoRows {
		return domain.MangaDetail{}, false, nil
	}
	if err != nil {
		return domain.MangaDetail{}, false, err
	}

	var authors, genres []string
	json.Unmarshal(row.Authors, &authors)
	json.Unmarshal(row.Genres, &genres)
	var chaptersByLang map[string]domain.ChapterStats
	json.Unmarshal(row.ChaptersByLang, &chaptersByLang)
	if chaptersByLang == nil {
		chaptersByLang = map[string]domain.ChapterStats{}
	}

	var updatedAt *time.Time
	if row.UpdatedAt.Valid {
		t := row.UpdatedAt.Time
		updatedAt = &t
	}
	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	d := domain.MangaDetail{
		Manga: domain.Manga{
			ID:             row.ID,
			DictionaryID:   row.DictionaryID,
			Title:          row.Title,
			Description:    row.Description,
			Status:         row.Status,
			CoverURL:       row.CoverUrl,
			Authors:        authors,
			Genres:         genres,
			State:          domain.MangaState(row.State),
			ChaptersByLang: chaptersByLang,
			UpdatedAt:      updatedAt,
			CreatedAt:      createdAt,
		},
	}
	return r.fillMangaDetail(ctx, d)
}

func (r *MySQLRepository) GetMangaByID(ctx context.Context, id string) (domain.MangaDetail, bool, error) {
	row, err := r.q.GetMangaByID(ctx, id)
	if err == sql.ErrNoRows {
		return domain.MangaDetail{}, false, nil
	}
	if err != nil {
		return domain.MangaDetail{}, false, err
	}

	var authors, genres []string
	json.Unmarshal(row.Authors, &authors)
	json.Unmarshal(row.Genres, &genres)
	var chaptersByLang map[string]domain.ChapterStats
	json.Unmarshal(row.ChaptersByLang, &chaptersByLang)
	if chaptersByLang == nil {
		chaptersByLang = map[string]domain.ChapterStats{}
	}

	var updatedAt *time.Time
	if row.UpdatedAt.Valid {
		t := row.UpdatedAt.Time
		updatedAt = &t
	}
	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	d := domain.MangaDetail{
		Manga: domain.Manga{
			ID:             row.ID,
			DictionaryID:   row.DictionaryID,
			Title:          row.Title,
			Description:    row.Description,
			Status:         row.Status,
			CoverURL:       row.CoverUrl,
			Authors:        authors,
			Genres:         genres,
			State:          domain.MangaState(row.State),
			ChaptersByLang: chaptersByLang,
			UpdatedAt:      updatedAt,
			CreatedAt:      createdAt,
		},
	}
	return r.fillMangaDetail(ctx, d)
}

func (r *MySQLRepository) fillMangaDetail(ctx context.Context, d domain.MangaDetail) (domain.MangaDetail, bool, error) {
	chRows, err := r.q.GetChaptersByManga(ctx, d.ID)
	if err != nil {
		return domain.MangaDetail{}, false, err
	}

	// Group chapters by language
	langMap := make(map[string]*domain.MangaLang)
	chapterMap := make(map[string][]domain.MangaChapter)

	for _, ch := range chRows {
		lang := ch.Lang
		if _, ok := langMap[lang]; !ok {
			langMap[lang] = &domain.MangaLang{Language: lang}
		}
		langMap[lang].Available++

		pageCount := 1
		if ch.ImageSrc != "" {
			pageCount = 1 // imageSrc is base path; actual page count derived from manifest
		}

		mc := domain.MangaChapter{
			MangaID:    ch.MangaID,
			Language:   lang,
			ChapterNum: ch.Name,
			PageCount:  pageCount,
			Uploaded:   ch.ImageSrc != "",
		}
		chapterMap[lang] = append(chapterMap[lang], mc)
	}

	for _, l := range langMap {
		d.Languages = append(d.Languages, *l)
	}
	for _, chs := range chapterMap {
		d.Chapters = append(d.Chapters, chs...)
	}

	return d, true, nil
}

// Chapter methods (replaces manga_chapters, manga_langs, chapter_pages, ingest_chapters)

func (r *MySQLRepository) UpsertChapter(ctx context.Context, id, mangaID, name string, chapterOrder int, lang, imageSrc string) error {
	return r.q.UpsertChapter(ctx, queries.UpsertChapterParams{
		ID:           id,
		MangaID:      mangaID,
		Name:         name,
		ChapterOrder: int32(chapterOrder),
		Lang:         lang,
		ImageSrc:     imageSrc,
	})
}

// UpsertChapterBatch inserts or updates multiple chapters in a single transaction.
// It auto-generates IDs for chapters that don't have one.
func (r *MySQLRepository) UpsertChapterBatch(ctx context.Context, chapters []domain.Chapter) error {
	if len(chapters) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chapters (id, manga_id, name, chapter_order, lang, image_src) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE name=VALUES(name), chapter_order=VALUES(chapter_order), lang=VALUES(lang), image_src=VALUES(image_src)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, ch := range chapters {
		id := ch.ID
		if id == "" {
			id = uuid.New().String()
		}
		_, err := stmt.ExecContext(ctx, id, ch.MangaID, ch.Number, int(ch.SortKey), ch.Language, ch.Source)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) GetChapterCountByLang(ctx context.Context, mangaID, lang string) (int, error) {
	n, err := r.q.GetChapterCountByLang(ctx, queries.GetChapterCountByLangParams{
		MangaID: mangaID,
		Lang:    lang,
	})
	return int(n), err
}

func (r *MySQLRepository) GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]domain.Chapter, error) {
	rows, err := r.q.GetChaptersByLang(ctx, queries.GetChaptersByLangParams{
		MangaID: mangaID,
		Lang:    lang,
	})
	if err != nil {
		return nil, err
	}
	chapters := make([]domain.Chapter, 0, len(rows))
	for _, ch := range rows {
		chapters = append(chapters, domain.Chapter{
			MangaID:  ch.MangaID,
			Number:   ch.Name,
			SortKey:  float64(ch.ChapterOrder),
			Language: ch.Lang,
		})
	}
	return chapters, nil
}

func (r *MySQLRepository) GetChaptersByManga(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	rows, err := r.q.GetChaptersByManga(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	chapters := make([]domain.Chapter, 0, len(rows))
	for _, ch := range rows {
		chapters = append(chapters, domain.Chapter{
			MangaID:  ch.MangaID,
			Number:   ch.Name,
			SortKey:  float64(ch.ChapterOrder),
			Language: ch.Lang,
		})
	}
	return chapters, nil
}

func (r *MySQLRepository) IsChapterIngested(ctx context.Context, mangaID, lang string, chapterOrder int) (bool, error) {
	n, err := r.q.IsChapterIngested(ctx, queries.IsChapterIngestedParams{
		MangaID:      mangaID,
		Lang:         lang,
		ChapterOrder: int32(chapterOrder),
	})
	return n > 0, err
}

// Dictionary methods

func (r *MySQLRepository) UpsertDictionary(ctx context.Context, entry domain.DictionaryEntry) error {
	sourcesJSON, _ := json.Marshal(entry.Sources)
	statsJSON, _ := json.Marshal(entry.SourceStats)
	bestJSON, _ := json.Marshal(entry.BestSource)
	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return r.q.UpsertDictionary(ctx, queries.UpsertDictionaryParams{
		ID:          entry.ID,
		Slug:        entry.Slug,
		Title:       entry.Title,
		Sources:     sourcesJSON,
		SourceStats: statsJSON,
		BestSource:  bestJSON,
		CoverUrl:    entry.CoverURL,
		CreatedAt:   createdAt,
	})
}

func (r *MySQLRepository) GetDictionary(ctx context.Context, id string) (domain.DictionaryEntry, bool, error) {
	row, err := r.q.GetDictionary(ctx, id)
	if err == sql.ErrNoRows {
		return domain.DictionaryEntry{}, false, nil
	}
	if err != nil {
		return domain.DictionaryEntry{}, false, err
	}
	return r.scanToDictionaryRow(row), true, nil
}

func (r *MySQLRepository) GetDictionaryBySlug(ctx context.Context, slug string) (domain.DictionaryEntry, bool, error) {
	row, err := r.q.GetDictionaryBySlug(ctx, slug)
	if err == sql.ErrNoRows {
		return domain.DictionaryEntry{}, false, nil
	}
	if err != nil {
		return domain.DictionaryEntry{}, false, err
	}
	return r.scanToDictionaryBySlugRow(row), true, nil
}

func (r *MySQLRepository) ListDictionary(ctx context.Context, filter domain.DictionaryFilter) (domain.DictionaryPage, error) {
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	rows, err := r.q.ListDictionary(ctx, queries.ListDictionaryParams{
		Column1: filter.Q,
		CONCAT:  filter.Q,
		Limit:   int32(pageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		return domain.DictionaryPage{}, err
	}

	entries := make([]domain.DictionaryEntry, 0, len(rows))
	var totalItems int
	for i, row := range rows {
		if i == 0 {
			totalItems = int(row.Total)
		}
		var sources, best map[string]string
		var sourceStats map[string]domain.SourceStat
		json.Unmarshal([]byte(row.Sources), &sources)
		json.Unmarshal([]byte(row.SourceStats), &sourceStats)
		json.Unmarshal([]byte(row.BestSource), &best)
		e := domain.DictionaryEntry{
			ID:          row.ID,
			Slug:        row.Slug,
			Title:       row.Title,
			CoverURL:    row.CoverUrl,
			Sources:     sources,
			SourceStats: sourceStats,
			BestSource:  best,
		}
		e.CreatedAt = row.CreatedAt
		entries = append(entries, e)
	}

	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	return domain.DictionaryPage{
		Items:      entries,
		TotalItems: totalItems,
		TotalPages: totalPages,
		PageSize:   pageSize,
		PageNumber: page,
	}, nil
}

func (r *MySQLRepository) UpsertDictionaryBatch(ctx context.Context, entries []domain.DictionaryEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, entry := range entries {
		sourcesJSON, _ := json.Marshal(entry.Sources)
		statsJSON, _ := json.Marshal(entry.SourceStats)
		bestJSON, _ := json.Marshal(entry.BestSource)
		createdAt := entry.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO dictionary (id, slug, title, sources, source_stats, best_source, cover_url, updated_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
			ON DUPLICATE KEY UPDATE
			    title=VALUES(title),
			    sources=JSON_MERGE_PATCH(dictionary.sources, VALUES(sources)),
			    source_stats=CASE WHEN VALUES(source_stats) != '{}' THEN VALUES(source_stats) ELSE dictionary.source_stats END,
			    best_source=CASE WHEN VALUES(best_source) != '{}' THEN VALUES(best_source) ELSE dictionary.best_source END,
			    cover_url=CASE WHEN VALUES(cover_url) != '' THEN VALUES(cover_url) ELSE dictionary.cover_url END,
			    updated_at=CURRENT_TIMESTAMP`,
			entry.ID, entry.Slug, entry.Title, sourcesJSON, statsJSON, bestJSON, entry.CoverURL, createdAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) scanToDictionaryRow(row queries.GetDictionaryRow) domain.DictionaryEntry {
	return scanToDictionaryEntryImpl(row.ID, row.Slug, row.Title, row.CoverUrl, row.Sources, row.SourceStats, row.BestSource, row.CreatedAt)
}

func (r *MySQLRepository) scanToDictionaryBySlugRow(row queries.GetDictionaryBySlugRow) domain.DictionaryEntry {
	return scanToDictionaryEntryImpl(row.ID, row.Slug, row.Title, row.CoverUrl, row.Sources, row.SourceStats, row.BestSource, row.CreatedAt)
}

func scanToDictionaryEntryImpl(id, slug, title, coverUrl string, sources, sourceStats, bestSource []byte, createdAt time.Time) domain.DictionaryEntry {
	var srcs, best map[string]string
	var stats map[string]domain.SourceStat
	json.Unmarshal(sources, &srcs)
	json.Unmarshal(sourceStats, &stats)
	json.Unmarshal(bestSource, &best)

	return domain.DictionaryEntry{
		ID:          id,
		Slug:        slug,
		Title:       title,
		CoverURL:    coverUrl,
		Sources:     srcs,
		SourceStats: stats,
		BestSource:  best,
		CreatedAt:   createdAt,
	}
}

// Ingest methods

func (r *MySQLRepository) GetDownloadedByLang(ctx context.Context, slug string) (map[string]int, error) {
	rows, err := r.q.GetDownloadedByLang(ctx, slug)
	if err != nil {
		return nil, err
	}
	m := make(map[string]int, len(rows))
	for _, row := range rows {
		m[row.Lang] = int(row.Count)
	}
	return m, nil
}

func (r *MySQLRepository) GetDownloadedChaptersByLang(ctx context.Context, slug, lang string) ([]string, error) {
	rows, err := r.q.GetDownloadedChaptersByLang(ctx, queries.GetDownloadedChaptersByLangParams{
		MangaID: slug,
		Lang:    lang,
	})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, r := range rows {
		result = append(result, fmt.Sprintf("%d", r))
	}
	return result, nil
}

// Manga methods

func (r *MySQLRepository) UpdateMangaCover(ctx context.Context, slug, coverURL string) error {
	// This is a simple UPDATE for the cover_url field.
	// We use an exec since there's no generated query for it.
	_, err := r.db.ExecContext(ctx, "UPDATE manga SET cover_url = ? WHERE slug = ?", coverURL, slug)
	return err
}
