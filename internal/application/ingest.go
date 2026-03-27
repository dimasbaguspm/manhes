package application

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"manga-engine/internal/domain"
)

type IngestService struct {
	repo      domain.Repository
	registry  domain.SourceRegistry
	dl        domain.Downloader
	disk      domain.Storer
	publisher domain.EventPublisher
	log       *slog.Logger
	metaMu    sync.Mutex
}

func NewIngestService(
	repo domain.Repository,
	reg domain.SourceRegistry,
	dl domain.Downloader,
	disk domain.Storer,
	pub domain.EventPublisher,
) *IngestService {
	return &IngestService{
		repo:      repo,
		registry:  reg,
		dl:        dl,
		disk:      disk,
		publisher: pub,
		log:       slog.With("service", "ingest"),
	}
}

func (s *IngestService) Ingest(ctx context.Context, e domain.IngestRequested) error {
	sources := s.registry.Ordered()
	if len(sources) == 0 {
		return fmt.Errorf("no sources registered")
	}

	var dictID string
	if entry, found, err := s.repo.GetDictionaryBySlug(e.Slug); err == nil && found {
		dictID = entry.ID
	}

	s.log.Info("ingest started",
		slog.String("dict_id", dictID),
		slog.String("slug", e.Slug),
		slog.Int("sources", len(e.Sources)),
	)

	type sourceResult struct {
		manga       *domain.Manga
		availByLang map[string]int
		hadErrors   bool // true if any chapter download failed
		err         error
	}

	results := make([]sourceResult, len(sources))
	var wg sync.WaitGroup
	for i, src := range sources {
		sourceID, ok := e.Sources[src.Source()]
		if !ok {
			continue
		}
		wg.Add(1)
		go func(i int, src domain.Scraper, sourceID string) {
			defer wg.Done()
			manga, avail, hadErr, err := s.runSource(ctx, e.Slug, dictID, src, sourceID, e.LangToSource)
			results[i] = sourceResult{manga: manga, availByLang: avail, hadErrors: hadErr, err: err}
		}(i, src, sourceID)
	}
	wg.Wait()

	var primaryManga *domain.Manga
	mergedAvail := make(map[string]int)
	allFailed := true
	anyDownloadErrors := false
	for i, r := range results {
		if r.err != nil {
			s.log.Warn("source failed", "source", sources[i].Source(), "slug", e.Slug, "err", r.err)
			continue
		}
		allFailed = false
		if r.hadErrors {
			anyDownloadErrors = true
		}
		if primaryManga == nil {
			primaryManga = r.manga
		}
		for lang, count := range r.availByLang {
			if count > mergedAvail[lang] {
				mergedAvail[lang] = count
			}
		}
	}
	if allFailed {
		return fmt.Errorf("all sources failed for %s", e.Slug)
	}

	s.log.Info("ingest complete",
		slog.String("dict_id", dictID),
		slog.String("slug", e.Slug),
	)

	if err := s.writeMetadata(e.Slug, primaryManga, mergedAvail); err != nil {
		return err
	}
	for lang, avail := range mergedAvail {
		if err := s.writeLangMetadata(e.Slug, lang, avail); err != nil {
			s.log.Warn("final lang metadata write failed", "lang", lang, "err", err)
		}
	}

	// Stamp updated_at / refreshed_at to record that we fetched from the
	// source, regardless of whether new chapters were downloaded.
	now := time.Now()
	if primaryManga != nil {
		if err := s.repo.UpsertManga(domain.Manga{
			Slug:        e.Slug,
			Title:       primaryManga.Title,
			Description: primaryManga.Description,
			Status:      primaryManga.Status,
			Authors:     primaryManga.Authors,
			Genres:      primaryManga.Genres,
			// CoverURL left empty: ON CONFLICT clause preserves the existing S3 URL.
		}); err != nil {
			s.log.Warn("ingest: stamp manga updated_at", "slug", e.Slug, "err", err)
		}
	}
	if dictID != "" {
		if entry, found, err := s.repo.GetDictionary(dictID); err == nil && found {
			entry.RefreshedAt = &now
			if err := s.repo.UpsertDictionary(entry); err != nil {
				s.log.Warn("ingest: stamp dictionary refreshed_at", "dict_id", dictID, "err", err)
			}
		}
	}

	if anyDownloadErrors {
		s.log.Warn("ingest: some chapters failed to download, keeping state fetching for retry",
			"slug", e.Slug, "dict_id", dictID,
		)
	} else {
		if err := s.repo.SetDictionaryStateBySlug(e.Slug, domain.StateUploading); err != nil {
			s.log.Warn("ingest: set state uploading", "slug", e.Slug, "err", err)
		}
		// If every chapter was already uploaded (no new downloads this run),
		// no ChapterDownloaded events will fire and the sync service will never
		// advance the state. Transition directly to available in that case.
		hasPending, err := s.repo.HasPendingChapters(e.Slug)
		if err != nil {
			s.log.Warn("ingest: check pending chapters", "slug", e.Slug, "err", err)
		} else if !hasPending {
			if err := s.repo.SetDictionaryStateBySlug(e.Slug, domain.StateAvailable); err != nil {
				s.log.Warn("ingest: set state available (no pending)", "slug", e.Slug, "err", err)
			}
		}
	}
	return nil
}

func (s *IngestService) runSource(
	ctx context.Context,
	slug, dictID string,
	src domain.Scraper,
	sourceID string,
	langToSource map[string]string,
) (*domain.Manga, map[string]int, bool, error) {

	s.log.Info("fetching manga detail", "dict_id", dictID, "source", src.Source(), "id", sourceID)
	manga, err := src.FetchMangaDetail(ctx, sourceID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("fetch detail: %w", err)
	}
	manga.Slug = slug

	if manga.CoverURL != "" {
		if data, err := s.dl.Download(ctx, manga.CoverURL); err == nil {
			ext := filepath.Ext(manga.CoverURL)
			if _, err := s.disk.SaveCover(slug, data, ext); err != nil {
				s.log.Warn("save cover failed", "err", err)
			}
		} else {
			s.log.Warn("download cover failed", "err", err)
		}
	}

	s.log.Info("fetching chapter list", "dict_id", dictID, "source", src.Source(), "slug", slug)
	chapters, err := src.FetchChapterList(ctx, sourceID)
	if err != nil {
		return manga, nil, false, fmt.Errorf("fetch chapters: %w", err)
	}

	byLang := make(map[string][]domain.Chapter)
	for _, ch := range chapters {
		// Skip languages owned by a different source to avoid duplicate content.
		if len(langToSource) > 0 {
			if bestSrc, ok := langToSource[ch.Language]; ok && bestSrc != src.Source() {
				continue
			}
		}
		byLang[ch.Language] = append(byLang[ch.Language], ch)
	}
	availByLang := make(map[string]int, len(byLang))
	for lang, chs := range byLang {
		availByLang[lang] = len(chs)
	}
	s.log.Info("chapter list fetched", "dict_id", dictID, "source", src.Source(), "total", len(chapters))

	s.writeMetadataLocked(slug, manga, availByLang)
	for lang, avail := range availByLang {
		s.writeLangMetadataLocked(slug, lang, avail)
	}

	hadErrors := false
	for _, lang := range orderedLangs(byLang) {
		for _, ch := range byLang[lang] {
			ch.MangaSlug = slug
			if err := s.downloadChapter(ctx, slug, dictID, lang, ch, src, manga, availByLang); err != nil {
				hadErrors = true
				s.log.Error("chapter download failed",
					"source", src.Source(),
					"dict_id", dictID,
					"slug", slug,
					"lang", lang,
					"num", ch.Number,
					"err", err,
				)
			}
		}
	}

	return manga, availByLang, hadErrors, nil
}

func (s *IngestService) downloadChapter(
	ctx context.Context,
	slug, dictID, lang string,
	ch domain.Chapter,
	src domain.Scraper,
	manga *domain.Manga,
	availByLang map[string]int,
) error {
	done, err := s.repo.IsChapterDownloaded(slug, lang, ch.Number)
	if err != nil {
		return fmt.Errorf("check chapter: %w", err)
	}
	if done {
		s.log.Debug("chapter already done", "source", src.Source(), "dict_id", dictID, "slug", slug, "lang", lang, "num", ch.Number)
		return nil
	}

	chLog := s.log.With(
		slog.String("source", src.Source()),
		slog.String("dict_id", dictID),
		slog.String("slug", slug),
		slog.String("lang", lang),
		slog.String("num", ch.Number),
	)

	chLog.Info("downloading chapter")
	pageURLs, err := src.FetchPageURLs(ctx, ch.SourceID)
	if err != nil {
		return fmt.Errorf("fetch page urls: %w", err)
	}
	ch.PageURLs = pageURLs

	start := time.Now()
	total := len(pageURLs)
	for i, u := range pageURLs {
		chLog.Debug("downloading page", slog.Int("page", i+1), slog.Int("total", total))
		data, err := s.dl.Download(ctx, u)
		if err != nil {
			return fmt.Errorf("page %d: %w", i+1, err)
		}
		ext := filepath.Ext(u)
		if ext == "" {
			ext = ".jpg"
		}
		if _, err := s.disk.SavePage(slug, lang, ch.Number, i+1, data, ext); err != nil {
			return fmt.Errorf("save page %d: %w", i+1, err)
		}
	}

	if err := s.disk.WriteChapterManifest(slug, lang, &ch); err != nil {
		chLog.Warn("write chapter manifest failed", "err", err)
	}
	if err := s.repo.MarkChapterDownloaded(slug, lang, ch.Number, ch.SortKey); err != nil {
		return fmt.Errorf("mark downloaded: %w", err)
	}

	chLog.Info("chapter done", slog.Int("pages", total), slog.Int64("duration_ms", time.Since(start).Milliseconds()))

	s.writeMetadataLocked(slug, manga, availByLang)
	s.writeLangMetadataLocked(slug, lang, availByLang[lang])

	if err := s.publisher.PublishChapterDownloaded(ctx, domain.ChapterDownloaded{
		Slug:       slug,
		Language:   lang,
		ChapterNum: ch.Number,
		SortKey:    ch.SortKey,
		PageCount:  len(pageURLs),
	}); err != nil {
		s.log.Warn("publish ChapterDownloaded failed", "err", err)
	}

	return nil
}
