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
	Repo domain.Repository
	Disk domain.Storer
	S3   domain.ObjectStore
	DL   domain.Downloader
	Bus  domain.EventBus
	Cfg  *config.Config
}

// FileUploadService handles downloading chapter pages from sources, uploading
// them to S3, and cleaning up local files. It publishes ChapterDownloaded
// and ChapterUploaded events per chapter.
type FileUploadService struct {
	repo domain.Repository
	disk domain.Storer
	s3c  domain.ObjectStore
	dl   domain.Downloader
	bus  domain.EventBus
	cfg  *config.Config
	log  *slog.Logger
}

func NewFileUploadService(cfg FileUploadServiceConfig) *FileUploadService {
	return &FileUploadService{
		repo: cfg.Repo,
		disk: cfg.Disk,
		s3c:  cfg.S3,
		dl:   cfg.DL,
		bus:  cfg.Bus,
		cfg:  cfg.Cfg,
		log:  slog.With("service", "file_upload"),
	}
}

// HandleChaptersFound reacts to ChaptersFound events and runs the download→upload→cleanup
// pipeline for each new chapter, publishing ChapterDownloaded and ChapterUploaded per chapter.
// It is fully context-aware: cancellation propagates through all stages.
func (s *FileUploadService) HandleChaptersFound(ctx context.Context, e domain.ChaptersFound) error {
	type chapterJob struct {
		lang string
		ch   domain.Chapter
	}

	jobs := make([]chapterJob, 0)
	for lang, chapters := range e.Chapters {
		for _, ch := range chapters {
			if len(ch.PageURLs) == 0 {
				continue
			}
			jobs = append(jobs, chapterJob{lang: lang, ch: ch})
		}
	}

	if len(jobs) == 0 {
		return nil
	}

	sem := make(chan struct{}, s.cfg.IngestConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var toUpsert []domain.Chapter
	var firstErr error

	for _, job := range jobs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(mangaID, dictionaryID, lang string, ch domain.Chapter) {
			defer wg.Done()
			defer func() { <-sem }()

			// Download pages to disk.
			localPaths := s.downloadPages(ctx, mangaID, lang, ch.Number, ch.PageURLs)

			// Publish ChapterDownloaded after disk save (before S3 upload).
			if len(localPaths) > 0 {
				s.bus.Publish(ctx, s.cfg.Bus.ChapterDownloaded, domain.ChapterDownloaded{
					DictionaryID: dictionaryID,
					MangaID:      mangaID,
					Language:     lang,
					ChapterNum:   ch.Number,
					SortKey:      ch.SortKey,
					PageCount:    len(localPaths),
				})
			}

			// Upload pages to S3.
			s3URLs := s.uploadPages(ctx, mangaID, lang, ch.Number, localPaths)

			// Cleanup local files after upload completes.
			s.cleanupPages(localPaths)

			// Publish ChapterUploaded after S3 upload and cleanup.
			if len(s3URLs) > 0 {
				s.bus.Publish(ctx, s.cfg.Bus.ChapterUploaded, domain.ChapterUploaded{
					DictionaryID: dictionaryID,
					MangaID:      mangaID,
					Language:     lang,
					ChapterNum:   ch.Number,
					SortKey:      ch.SortKey,
					PageCount:    len(s3URLs),
					S3URLs:       s3URLs,
				})
			}

			mu.Lock()
			if firstErr == nil && len(s3URLs) > 0 {
				toUpsert = append(toUpsert, domain.Chapter{
					MangaID:  mangaID,
					Number:   ch.Number,
					SortKey:  ch.SortKey,
					Language: lang,
					Source:   fmt.Sprintf("%s/%s/%s", mangaID, lang, ch.Number),
					PageURLs: s3URLs,
				})
			}
			mu.Unlock()
		}(e.MangaID, e.DictionaryID, job.lang, job.ch)
	}

	wg.Wait()

	if len(toUpsert) > 0 {
		if err := s.repo.UpsertChapterBatch(ctx, toUpsert); err != nil {
			s.log.Warn("HandleChaptersFound: batch upsert", "mangaID", e.MangaID, "err", err)
		}
	}

	// Upload manga cover from disk to S3 if present.
	coverURL, err := s.uploadCoverFromDisk(ctx, e.MangaID)
	if err != nil {
		s.log.Warn("HandleChaptersFound: upload cover", "mangaID", e.MangaID, "err", err)
	}
	if coverURL != "" {
		if err := s.repo.UpdateMangaCover(ctx, e.MangaID, coverURL); err != nil {
			s.log.Warn("HandleChaptersFound: update manga cover", "mangaID", e.MangaID, "err", err)
		}
	}

	return firstErr
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
