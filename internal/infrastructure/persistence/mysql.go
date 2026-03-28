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

// Catalog methods

func (r *MySQLRepository) UpsertManga(ctx context.Context, m domain.Manga) error {
	authors, _ := json.Marshal(m.Authors)
	genres, _ := json.Marshal(m.Genres)
	return r.q.UpsertManga(ctx, queries.UpsertMangaParams{
		Uuid:        m.DictionaryID,
		Slug:        m.Slug,
		Title:       m.Title,
		Description: m.Description,
		Status:      m.Status,
		Authors:     authors,
		Genres:      genres,
		CoverUrl:    m.CoverURL,
	})
}

func (r *MySQLRepository) ListManga(ctx context.Context, filter domain.MangaFilter) (domain.MangaPage, error) {
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

	rows, err := r.q.ListManga(ctx, queries.ListMangaParams{
		Column1: filter.Title,
		CONCAT:  filter.Title,
		Column3: filter.State,
		State:   filter.State,
		Column5: filter.Status,
		Status:  filter.Status,
		Column7: filter.HideUnavailable,
		Column8: sortBy,
		Column9: sortBy,
		Limit:   int32(pageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		return domain.MangaPage{}, err
	}
	if len(rows) == 0 {
		return domain.MangaPage{Items: []domain.Manga{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	items := make([]domain.Manga, 0, len(rows))
	for _, row := range rows {
		var authors, genres, langs []string
		json.Unmarshal(row.Authors, &authors)
		json.Unmarshal(row.Genres, &genres)
		json.Unmarshal(row.Languages, &langs)
		var updatedAt *time.Time
		t := row.SyncedAt
		updatedAt = &t
		var dictID string
		if row.DictID != "" {
			dictID = row.DictID
		}
		var state string
		if row.State != "" {
			state = row.State
		}
		var coverURL string
		if row.CoverUrl != nil {
			coverURL, _ = row.CoverUrl.(string)
		}
		items = append(items, domain.Manga{
			DictionaryID: dictID,
			Slug:         row.Slug,
			Title:        row.Title,
			Description:  row.Description,
			Status:       row.Status,
			CoverURL:     coverURL,
			Authors:      authors,
			Genres:       genres,
			Languages:    langs,
			UpdatedAt:    updatedAt,
			State:        domain.MangaState(state),
		})
	}

	return domain.MangaPage{Items: items, Total: int(rows[0].Total), Page: page, PageSize: pageSize}, nil
}

func (r *MySQLRepository) GetMangaBySlug(ctx context.Context, slug string) (domain.MangaDetail, bool, error) {
	row, err := r.q.GetMangaBySlug(ctx, slug)
	if err == sql.ErrNoRows {
		return domain.MangaDetail{}, false, nil
	}
	if err != nil {
		return domain.MangaDetail{}, false, err
	}

	var authors, genres []string
	json.Unmarshal(row.Authors, &authors)
	json.Unmarshal(row.Genres, &genres)

	var dictID, state string
	if row.DictID.Valid {
		dictID = row.DictID.String
	}
	if row.State.Valid {
		state = row.State.String
	}

	d := domain.MangaDetail{
		Manga: domain.Manga{
			DictionaryID: dictID,
			Slug:         row.Slug,
			Title:        row.Title,
			Description:  row.Description,
			Status:       row.Status,
			CoverURL:     row.CoverUrl,
			Authors:      authors,
			Genres:       genres,
			State:        domain.MangaState(state),
		},
	}
	return r.fillMangaDetail(ctx, d)
}

func (r *MySQLRepository) fillMangaDetail(ctx context.Context, d domain.MangaDetail) (domain.MangaDetail, bool, error) {
	chRows, err := r.q.GetChaptersByManga(ctx, d.Slug)
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
			Slug:       ch.MangaID,
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
			MangaSlug: ch.MangaID,
			Number:    ch.Name,
			SortKey:   float64(ch.ChapterOrder),
			Language:  ch.Lang,
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
			MangaSlug: ch.MangaID,
			Number:    ch.Name,
			SortKey:   float64(ch.ChapterOrder),
			Language:  ch.Lang,
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
	var refreshedAt sql.NullTime
	if entry.RefreshedAt != nil {
		refreshedAt = sql.NullTime{Time: *entry.RefreshedAt, Valid: true}
	}
	return r.q.UpsertDictionary(ctx, queries.UpsertDictionaryParams{
		ID:          entry.ID,
		Slug:        entry.Slug,
		Title:       entry.Title,
		Sources:     sourcesJSON,
		SourceStats: statsJSON,
		BestSource:  bestJSON,
		State:       string(entry.State),
		CoverUrl:    entry.CoverURL,
		RefreshedAt: refreshedAt,
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
	return r.scanToEntry(row), true, nil
}

func (r *MySQLRepository) GetDictionaryBySlug(ctx context.Context, slug string) (domain.DictionaryEntry, bool, error) {
	row, err := r.q.GetDictionaryBySlug(ctx, slug)
	if err == sql.ErrNoRows {
		return domain.DictionaryEntry{}, false, nil
	}
	if err != nil {
		return domain.DictionaryEntry{}, false, err
	}
	return r.scanToEntry(row), true, nil
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
			State:       domain.MangaState(row.State),
		}
		e.CreatedAt = row.CreatedAt
		if row.RefreshedAt.Valid {
			t := row.RefreshedAt.Time
			e.RefreshedAt = &t
		}
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

func (r *MySQLRepository) SetDictionaryState(ctx context.Context, id string, state domain.MangaState) error {
	return r.q.SetDictionaryState(ctx, queries.SetDictionaryStateParams{
		State: string(state),
		ID:    id,
	})
}

func (r *MySQLRepository) SetDictionaryStateBySlug(ctx context.Context, slug string, state domain.MangaState) error {
	return r.q.SetDictionaryStateBySlug(ctx, queries.SetDictionaryStateBySlugParams{
		State: string(state),
		Slug:  slug,
	})
}

func (r *MySQLRepository) scanToEntry(row queries.Dictionary) domain.DictionaryEntry {
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
		State:       domain.MangaState(row.State),
	}
	e.CreatedAt = row.CreatedAt
	if row.RefreshedAt.Valid {
		t := row.RefreshedAt.Time
		e.RefreshedAt = &t
	}
	return e
}

// Watchlist methods

func (r *MySQLRepository) ListWatchlist(ctx context.Context) ([]domain.WatchlistEntry, error) {
	rows, err := r.q.ListWatchlist(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]domain.WatchlistEntry, 0, len(rows))
	for _, row := range rows {
		var sources map[string]string
		json.Unmarshal(row.Sources, &sources)
		e := domain.WatchlistEntry{
			ID:           row.ID,
			Slug:         row.Slug,
			Title:        row.Title,
			DictionaryID: row.DictionaryID,
			Sources:      sources,
		}
		if row.LastCheckedAt.Valid {
			t := row.LastCheckedAt.Time
			e.LastChecked = &t
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (r *MySQLRepository) AddWatchlist(ctx context.Context, entry domain.WatchlistEntry) error {
	sourcesJSON, _ := json.Marshal(entry.Sources)
	id := entry.ID
	if id == "" {
		id = uuid.New().String()
	}
	return r.q.AddWatchlist(ctx, queries.AddWatchlistParams{
		ID:           id,
		Slug:         entry.Slug,
		Title:        entry.Title,
		Sources:      sourcesJSON,
		DictionaryID: sql.NullString{String: entry.DictionaryID, Valid: entry.DictionaryID != ""},
	})
}

func (r *MySQLRepository) RemoveWatchlist(ctx context.Context, slug string) error {
	return r.q.RemoveWatchlist(ctx, slug)
}

func (r *MySQLRepository) UpdateLastChecked(ctx context.Context, slug string, t time.Time) error {
	return r.q.UpdateLastChecked(ctx, queries.UpdateLastCheckedParams{
		LastCheckedAt: sql.NullTime{Time: t, Valid: true},
		Slug:          slug,
	})
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
