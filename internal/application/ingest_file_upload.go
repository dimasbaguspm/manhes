package application

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"sync"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// FileUploadServiceConfig holds dependencies for FileUploadService.
type FileUploadServiceConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	Disk     domain.Storer
	S3       domain.ObjectStore
	DL       domain.Downloader
	Bus      domain.EventBus
	Cfg      *config.Config
}

// FileUploadService handles downloading chapter pages from sources, uploading
// them to S3, and cleaning up local files. It publishes ChapterDownloaded
// and ChapterUploaded events per chapter.
type FileUploadService struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	disk     domain.Storer
	s3c      domain.ObjectStore
	dl       domain.Downloader
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewFileUploadService(cfg FileUploadServiceConfig) *FileUploadService {
	return &FileUploadService{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		disk:     cfg.Disk,
		s3c:      cfg.S3,
		dl:       cfg.DL,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("service", "file_upload"),
	}
}

// HandleChaptersFound reacts to ChaptersFound events and runs the download→upload→cleanup
// pipeline for each new chapter, publishing ChapterDownloaded and ChapterUploaded per chapter.
// It fetches dictionary metadata and real page URLs from sources when chapters are placeholders.
// It is fully context-aware: cancellation propagates through all stages.
func (s *FileUploadService) HandleChaptersFound(ctx context.Context, e domain.ChaptersFound) error {
	s.log.Info("[FileUpload Service]: HandleChaptersFound: received event",
		"dictionaryID", e.DictionaryID,
		"mangaID", e.MangaID,
		"languages", func() []string {
			langs := make([]string, 0, len(e.Chapters))
			for l := range e.Chapters {
				langs = append(langs, l)
			}
			return langs
		}(),
	)

	// Fetch dictionary entry to get source IDs and best-source mapping.
	entry, found, err := s.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		s.log.Error("[FileUpload Service]: GetDictionary failed", "dictionaryID", e.DictionaryID, "err", err)
		return err
	}
	if !found {
		s.log.Warn("[FileUpload Service]: dictionary entry not found, skipping", "dictionaryID", e.DictionaryID)
		return nil
	}
	s.log.Info("[FileUpload Service]: dictionary entry fetched",
		"dictionaryID", entry.ID,
		"slug", entry.Slug,
		"sources", entry.Sources,
		"bestSource", entry.BestSource,
	)

	// Build effective source map: prefer BestSource, fall back to all Sources.
	sourceMap := entry.Sources
	if len(entry.BestSource) > 0 {
		sourceMap = make(map[string]string)
		for _, srcName := range entry.BestSource {
			if id, ok := entry.Sources[srcName]; ok {
				sourceMap[srcName] = id
			}
		}
	}

	// For each placeholder chapter, resolve PageURLs from sources.
	type chapterJob struct {
		lang     string
		ch       domain.Chapter
		pageURLs []string
	}
	jobs := make([]chapterJob, 0)

	for lang := range e.Chapters {
		// Resolve the best source for this language.
		srcName := entry.BestSource[lang]
		if srcName == "" {
			s.log.Warn("[FileUpload Service]: no best source for language, skipping",
				"mangaID", e.MangaID, "lang", lang)
			continue
		}
		srcID, ok := sourceMap[srcName]
		if !ok {
			s.log.Warn("[FileUpload Service]: no source ID for language, skipping",
				"mangaID", e.MangaID, "lang", lang, "source", srcName)
			continue
		}

		s.log.Info("[FileUpload Service]: fetching chapter list from source",
			"mangaID", e.MangaID, "lang", lang, "source", srcName, "sourceID", srcID)

		// Fetch the full chapter list for this language from the source.
		srcChapters, err := s.fetchChapterListFromSource(ctx, srcName, srcID)
		if err != nil {
			s.log.Warn("[FileUpload Service]: fetchChapterListFromSource failed, skipping language",
				"mangaID", e.MangaID, "lang", lang, "source", srcName, "err", err)
			continue
		}
		if len(srcChapters) == 0 {
			s.log.Info("[FileUpload Service]: no chapters returned from source, skipping language",
				"mangaID", e.MangaID, "lang", lang, "source", srcName)
			continue
		}
		s.log.Info("[FileUpload Service]: chapter list fetched from source",
			"mangaID", e.MangaID, "lang", lang, "source", srcName, "chapterCount", len(srcChapters))

		// Get already-uploaded chapter names for this language to diff.
		storedChapters, err := s.repo.GetChaptersByLang(ctx, e.MangaID, lang)
		if err != nil {
			s.log.Warn("[FileUpload Service]: GetChaptersByLang failed, skipping language",
				"mangaID", e.MangaID, "lang", lang, "err", err)
			continue
		}
		storedNames := make(map[string]bool)
		for _, sc := range storedChapters {
			storedNames[sc.Number] = true
		}
		s.log.Info("[FileUpload Service]: stored chapters loaded",
			"mangaID", e.MangaID, "lang", lang, "storedCount", len(storedNames))

		// Process each chapter from the source that hasn't been stored yet.
		for _, ch := range srcChapters {
			if storedNames[ch.Number] {
				s.log.Debug("[FileUpload Service]: chapter already stored, skipping",
					"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number)
				continue
			}

			s.log.Info("[FileUpload Service]: fetching page URLs for new chapter",
				"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number, "source", srcName)

			pageURLs, err := s.fetchPageURLsFromSource(ctx, srcName, srcID, ch.Number)
			if err != nil {
				s.log.Warn("[FileUpload Service]: fetchPageURLsFromSource failed, skipping chapter",
					"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number, "err", err)
				continue
			}
			if len(pageURLs) == 0 {
				s.log.Warn("[FileUpload Service]: no page URLs returned, skipping chapter",
					"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number)
				continue
			}

			// Double-check not already uploaded (race condition guard).
			uploaded, err := s.repo.GetChapterUploaded(ctx, e.MangaID, lang, ch.Number)
			if err != nil {
				s.log.Warn("[FileUpload Service]: GetChapterUploaded failed, treating as not uploaded",
					"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number, "err", err)
			}
			if uploaded {
				s.log.Info("[FileUpload Service]: chapter already uploaded, skipping",
					"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number)
				continue
			}

			s.log.Info("[FileUpload Service]: queuing chapter for download",
				"mangaID", e.MangaID, "lang", lang, "chapter", ch.Number, "pageCount", len(pageURLs))

			chWithURLs := ch
			chWithURLs.PageURLs = pageURLs
			jobs = append(jobs, chapterJob{lang: lang, ch: chWithURLs, pageURLs: pageURLs})
		}
	}

	if len(jobs) == 0 {
		s.log.Info("[FileUpload Service]: no chapters with page URLs to process")
		return nil
	}
	s.log.Info("[FileUpload Service]: starting download pipeline",
		"mangaID", e.MangaID,
		"jobCount", len(jobs),
		"concurrency", s.cfg.IngestConcurrency,
	)

	sem := make(chan struct{}, s.cfg.IngestConcurrency)
	var wg sync.WaitGroup
	var firstErr error

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			s.log.Warn("[FileUpload Service]: context cancelled, aborting", "mangaID", e.MangaID)
			return ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(mangaID, dictionaryID, lang string, ch domain.Chapter, pageURLs []string) {
			defer wg.Done()
			defer func() { <-sem }()

			s.log.Info("[FileUpload Service]: downloading chapter",
				"mangaID", mangaID, "lang", lang, "chapter", ch.Number, "pageCount", len(pageURLs))

			// Download pages to disk.
			localPaths := s.downloadPages(ctx, mangaID, lang, ch.Number, pageURLs)
			if len(localPaths) == 0 {
				s.log.Warn("[FileUpload Service]: downloadPages returned no paths, skipping chapter",
					"mangaID", mangaID, "lang", lang, "chapter", ch.Number)
				return
			}
			s.log.Info("[FileUpload Service]: pages downloaded to disk",
				"mangaID", mangaID, "lang", lang, "chapter", ch.Number,
				"localPaths", localPaths)

			// Publish ChapterDownloaded after disk save (before S3 upload).
			s.bus.Publish(ctx, s.cfg.Bus.ChapterDownloaded, domain.ChapterDownloaded{
				DictionaryID: dictionaryID,
				MangaID:      mangaID,
				Language:     lang,
				ChapterNum:   ch.Number,
				SortKey:      ch.SortKey,
				PageCount:    len(localPaths),
			})
			s.log.Info("[FileUpload Service]: ChapterDownloaded published",
				"mangaID", mangaID, "lang", lang, "chapter", ch.Number, "pageCount", len(localPaths))

			// Upload pages to S3.
			s3URLs := s.uploadPages(ctx, mangaID, lang, ch.Number, localPaths)
			s.log.Info("[FileUpload Service]: S3 upload complete",
				"mangaID", mangaID, "lang", lang, "chapter", ch.Number,
				"s3URLCount", len(s3URLs))

			// Cleanup local files after upload completes.
			s.cleanupPages(localPaths)

			// Upsert chapter with image_src before publishing ChapterUploaded,
			// so HandleChapterUploaded can correctly count available chapters.
			if len(s3URLs) > 0 {
				chapterImgSrc := fmt.Sprintf("%s/%s/%s", mangaID, lang, ch.Number)
				if err := s.repo.UpsertChapterBatch(ctx, []domain.Chapter{{
					MangaID:  mangaID,
					Number:   ch.Number,
					SortKey:  ch.SortKey,
					Language: lang,
					Source:   chapterImgSrc,
					PageURLs: s3URLs,
				}}); err != nil {
					s.log.Error("[FileUpload Service] upload: UpsertChapterBatch failed",
						"mangaID", mangaID, "lang", lang, "chapter", ch.Number, "err", err)
				}

				s.bus.Publish(ctx, s.cfg.Bus.ChapterUploaded, domain.ChapterUploaded{
					DictionaryID: dictionaryID,
					MangaID:      mangaID,
					Language:     lang,
					ChapterNum:   ch.Number,
					SortKey:      ch.SortKey,
					PageCount:    len(s3URLs),
					S3URLs:       s3URLs,
				})
				s.log.Info("[FileUpload Service]: ChapterUploaded published",
					"mangaID", mangaID, "lang", lang, "chapter", ch.Number, "pageCount", len(s3URLs))
			}
		}(e.MangaID, e.DictionaryID, job.lang, job.ch, job.pageURLs)
	}

	wg.Wait()
	s.log.Info("[FileUpload Service]: download pipeline complete",
		"mangaID", e.MangaID,
		"dictionaryID", e.DictionaryID,
		"firstErr", firstErr,
	)

	// Check if all expected chapters have been uploaded and publish MangaAvailable if so.
	if mangaAllUploaded, allUploaded := s.checkMangaAllUploaded(ctx, e.MangaID, e.DictionaryID); allUploaded {
		s.log.Info("[FileUpload Service]: all chapters uploaded, publishing MangaAvailable",
			"mangaID", e.MangaID, "dictionaryID", e.DictionaryID)
		if err := s.bus.Publish(ctx, s.cfg.Bus.MangaAvailable, domain.MangaAvailable{
			DictionaryID: e.DictionaryID,
			MangaID:      e.MangaID,
		}); err != nil {
			s.log.Error("[FileUpload Service]: publish MangaAvailable failed",
				"mangaID", e.MangaID, "dictionaryID", e.DictionaryID, "err", err)
		}
	} else {
		s.log.Info("[FileUpload Service]: not all chapters uploaded yet",
			"mangaID", e.MangaID,
			"reason", mangaAllUploaded,
		)
	}

	// Upload manga cover from disk to S3 if present.
	coverURL, err := s.uploadCoverFromDisk(ctx, e.MangaID)
	if err != nil {
		s.log.Warn("[FileUpload Service]: uploadCoverFromDisk failed", "mangaID", e.MangaID, "err", err)
	}
	if coverURL != "" {
		if err := s.repo.UpdateMangaCover(ctx, e.MangaID, coverURL); err != nil {
			s.log.Error("[FileUpload Service]: UpdateMangaCover failed", "mangaID", e.MangaID, "err", err)
		} else {
			s.log.Info("[FileUpload Service]: manga cover updated", "mangaID", e.MangaID, "coverURL", coverURL)
		}
	}

	return firstErr
}

// fetchChapterListFromSource fetches the full chapter list for a manga from the given source.
func (s *FileUploadService) fetchChapterListFromSource(ctx context.Context, srcName, srcID string) ([]domain.Chapter, error) {
	scraper := s.scraperFor(srcName)
	if scraper == nil {
		return nil, fmt.Errorf("scraper not registered: %s", srcName)
	}

	chapters, err := scraper.FetchChapterList(ctx, srcID)
	if err != nil {
		return nil, fmt.Errorf("FetchChapterList: %w", err)
	}

	return chapters, nil
}

// fetchPageURLsFromSource fetches all chapters for a manga from the given source,
// then returns the page URLs for the chapter matching the given chapter number.
func (s *FileUploadService) fetchPageURLsFromSource(ctx context.Context, srcName, srcID, chapterNum string) ([]string, error) {
	scraper := s.scraperFor(srcName)
	if scraper == nil {
		return nil, fmt.Errorf("scraper not registered: %s", srcName)
	}

	allChapters, err := scraper.FetchChapterList(ctx, srcID)
	if err != nil {
		return nil, fmt.Errorf("FetchChapterList: %w", err)
	}

	// Find the matching chapter by number.
	for _, ch := range allChapters {
		if ch.Number == chapterNum {
			// Fetch page URLs for this specific chapter.
			pages, err := scraper.FetchPageURLs(ctx, ch.SourceID)
			if err != nil {
				return nil, fmt.Errorf("FetchPageURLs for chapter %s: %w", chapterNum, err)
			}
			return pages, nil
		}
	}

	return nil, fmt.Errorf("chapter %s not found in source %s", chapterNum, srcName)
}

// checkMangaAllUploaded checks whether all expected chapters for a manga have been
// uploaded to S3. It returns (reason, true) if all are uploaded, or (reason, false) otherwise.
func (s *FileUploadService) checkMangaAllUploaded(ctx context.Context, mangaID, dictionaryID string) (string, bool) {
	// Get expected chapter counts from the manga's chapters_by_lang.
	mangaDetail, found, err := s.repo.GetMangaByDictionaryID(ctx, dictionaryID)
	if err != nil || !found {
		return "manga_not_found", false
	}

	// Get actual uploaded counts per language from the chapters table.
	uploadedByLang := make(map[string]int)
	for lang := range mangaDetail.ChaptersByLang {
		storedChapters, err := s.repo.GetChaptersByLang(ctx, mangaID, lang)
		if err != nil {
			return fmt.Sprintf("get_chapters_failed:%s", lang), false
		}
		for _, ch := range storedChapters {
			uploaded, err := s.repo.GetChapterUploaded(ctx, mangaID, lang, ch.Number)
			if err != nil {
				return fmt.Sprintf("get_chapter_uploaded_failed:%s:%s", lang, ch.Number), false
			}
			if uploaded {
				uploadedByLang[lang]++
			}
		}
	}

	// Compare uploaded counts against expected totals.
	for lang, expected := range mangaDetail.ChaptersByLang {
		uploaded := uploadedByLang[lang]
		if uploaded < expected.Total {
			return fmt.Sprintf("lang_%s: %d/%d uploaded", lang, uploaded, expected.Total), false
		}
	}

	return "", true
}

// scraperFor returns the scraper registered under the given source name, or nil.
func (s *FileUploadService) scraperFor(name string) domain.Scraper {
	for _, candidate := range s.registry.Ordered() {
		if candidate.Source() == name {
			return candidate
		}
	}
	return nil
}

// downloadPages downloads all pages for a chapter and returns local disk paths.
// It is context-aware and will abort early if the context is cancelled.
func (s *FileUploadService) downloadPages(ctx context.Context, mangaID, lang, chapterNum string, urls []string) []string {
	paths := make([]string, 0, len(urls))
	for i, u := range urls {
		select {
		case <-ctx.Done():
			return paths
		default:
		}

		data, err := s.dl.Download(ctx, u)
		if err != nil {
			s.log.Warn("downloadPages", "mangaID", mangaID, "lang", lang, "ch", chapterNum, "page", i, "err", err)
			continue
		}
		ext := extFor(u)
		path, err := s.disk.SavePage(mangaID, lang, chapterNum, i+1, data, ext)
		if err != nil {
			s.log.Warn("downloadPages: save", "mangaID", mangaID, "err", err)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

// uploadPages reads local files and uploads them to S3, returning the resulting S3 URLs.
// It is context-aware and will abort early if the context is cancelled.
func (s *FileUploadService) uploadPages(ctx context.Context, mangaID, lang, chapterNum string, localPaths []string) []string {
	keyPrefix := fmt.Sprintf("%s/%s/%s", mangaID, lang, chapterNum)
	urls := make([]string, 0, len(localPaths))
	for i, path := range localPaths {
		select {
		case <-ctx.Done():
			return urls
		default:
		}

		data, err := os.ReadFile(path)
		if err != nil {
			s.log.Warn("uploadPages: read", "path", path, "err", err)
			continue
		}
		ext := extFor(path)
		pageKey := fmt.Sprintf("%s/%03d%s", keyPrefix, i+1, ext)
		url, err := s.s3c.Upload(ctx, pageKey, data, mimeFor(ext))
		if err != nil {
			s.log.Warn("uploadPages: upload", "key", pageKey, "err", err)
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

// cleanupPages removes downloaded local files after successful S3 upload.
func (s *FileUploadService) cleanupPages(paths []string) {
	for _, path := range paths {
		os.Remove(path)
	}
}

var coverExtensions = []string{".jpg", ".jpeg", ".png", ".webp"}

// uploadCoverFromDisk uploads the manga cover from disk to S3 if not already uploaded.
// mangaDir is the manga UUID directory on disk.
func (s *FileUploadService) uploadCoverFromDisk(ctx context.Context, mangaDir string) (string, error) {
	for _, ext := range coverExtensions {
		path := filepath.Join(s.cfg.LibraryPath, mangaDir, "cover"+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "image/jpeg"
		}
		key := fmt.Sprintf("%s/cover%s", mangaDir, ext)
		url, err := s.s3c.Upload(ctx, key, data, ct)
		if err != nil {
			return "", err
		}
		os.Remove(path) // cleanup local cover after upload
		s.log.Info("file_upload: cover uploaded", "mangaID", mangaDir, "url", url)
		return url, nil
	}
	return "", nil
}

// extFor returns the file extension (with dot), defaulting to ".jpg".
func extFor(path string) string {
	if ext := filepath.Ext(path); ext != "" {
		return ext
	}
	return ".jpg"
}

// mimeFor returns the MIME type for a file extension, defaulting to image/jpeg.
func mimeFor(ext string) string {
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	return "image/jpeg"
}
