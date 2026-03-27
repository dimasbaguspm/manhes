package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/storage"
)

var coverExtensions = []string{".jpg", ".jpeg", ".png", ".webp"}

type SyncConfig struct {
	LibraryPath string
	Interval    time.Duration
}

type SyncService struct {
	repo     domain.Repository
	disk     domain.Storer
	diskPath string
	interval time.Duration
	s3       domain.ObjectStore
	log      *slog.Logger
}

func NewSyncService(repo domain.Repository, disk domain.Storer, s3c domain.ObjectStore, cfg SyncConfig) *SyncService {
	return &SyncService{
		repo:     repo,
		disk:     disk,
		diskPath: cfg.LibraryPath,
		interval: cfg.Interval,
		s3:       s3c,
		log:      slog.With("service", "sync"),
	}
}

func (s *SyncService) HandleChapterDownloaded(ctx context.Context, e domain.ChapterDownloaded) error {
	dictID := s.dictIDForSlug(e.Slug)
	s.log.Info("chapter downloaded event",
		slog.String("dict_id", dictID),
		slog.String("slug", e.Slug),
		slog.String("lang", e.Language),
		slog.String("chapter", e.ChapterNum),
		slog.Int("pages", e.PageCount),
	)

	m, err := s.disk.ReadMetadata(e.Slug)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	if m != nil {
		coverURL, _ := s.uploadCover(ctx, e.Slug, dictID)
		if err := s.repo.UpsertManga(domain.Manga{
			Slug:        e.Slug,
			Title:       m.Title,
			Description: m.Description,
			Status:      m.Status,
			Authors:     m.Authors,
			Genres:      m.Genres,
			CoverURL:    coverURL,
		}); err != nil {
			return fmt.Errorf("upsert manga: %w", err)
		}
		if stat, ok := m.Languages[e.Language]; ok {
			if err := s.repo.UpsertLang(e.Slug, e.Language, stat.Available, stat.Downloaded); err != nil {
				s.log.Warn("upsert lang", "err", err)
			}
		}
	}

	if err := s.repo.UpsertChapter(e.Slug, e.Language, e.ChapterNum, e.SortKey, e.PageCount); err != nil {
		return fmt.Errorf("upsert chapter: %w", err)
	}

	return s.uploadChapter(ctx, domain.ChapterRef{
		DictionaryID: dictID,
		Slug:         e.Slug,
		Language:     e.Language,
		ChapterNum:   e.ChapterNum,
	})
}

// SyncAll runs an initial full scan and then ticks at the configured interval,
// catching any chapters that existed before the service started.
func (s *SyncService) SyncAll(ctx context.Context) {
	s.tick(ctx)
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.tick(ctx)
		}
	}
}

func (s *SyncService) tick(ctx context.Context) {
	s.log.Info("sync: starting tick")

	entries, err := os.ReadDir(s.diskPath)
	if err != nil {
		s.log.Error("sync: read library dir", "err", err)
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if err := s.syncSlug(ctx, e.Name()); err != nil {
			s.log.Error("sync: slug failed", "slug", e.Name(), "err", err)
		}
	}

	pending, err := s.repo.GetPendingChapters()
	if err != nil {
		s.log.Error("sync: get pending chapters", "err", err)
		return
	}
	for _, ref := range pending {
		if ctx.Err() != nil {
			return
		}
		if err := s.uploadChapter(ctx, ref); err != nil {
			s.log.Error("sync: upload chapter", "slug", ref.Slug, "lang", ref.Language, "chapter", ref.ChapterNum, "err", err)
		}
	}

	s.log.Info("sync: tick complete")
}

func (s *SyncService) syncSlug(ctx context.Context, slug string) error {
	m, err := s.disk.ReadMetadata(slug)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	if m == nil {
		return nil
	}

	coverURL, _ := s.uploadCover(ctx, slug, s.dictIDForSlug(slug))
	if err := s.repo.UpsertManga(domain.Manga{
		Slug:        slug,
		Title:       m.Title,
		Description: m.Description,
		Status:      m.Status,
		Authors:     m.Authors,
		Genres:      m.Genres,
		CoverURL:    coverURL,
	}); err != nil {
		return fmt.Errorf("upsert manga: %w", err)
	}

	for lang, stat := range m.Languages {
		if err := s.repo.UpsertLang(slug, lang, stat.Available, stat.Downloaded); err != nil {
			s.log.Warn("sync: upsert lang", "slug", slug, "lang", lang, "err", err)
			continue
		}

		langMeta, err := s.disk.ReadLangMetadata(slug, lang)
		if err != nil || langMeta == nil {
			continue
		}

		for _, chNum := range langMeta.Chapters {
			chDir := filepath.Join(s.diskPath, slug, lang, storage.ChapterDir(chNum))
			pageCount, err := countPageFiles(chDir)
			if err != nil {
				continue
			}
			// Skip if no local files — they were deleted after S3 upload; the
			// existing page_count in the DB must not be zeroed out.
			if pageCount == 0 {
				continue
			}
			sortKey := domain.ParseChapterSortKey(chNum)
			if err := s.repo.UpsertChapter(slug, lang, chNum, sortKey, pageCount); err != nil {
				s.log.Warn("sync: upsert chapter", "slug", slug, "lang", lang, "chapter", chNum, "err", err)
			}
		}
	}

	dictID := s.dictIDForSlug(slug)
	prefix := dictID
	if prefix == "" {
		prefix = slug
	}

	metaPath := filepath.Join(s.diskPath, slug, "metadata.json")
	if err := s.uploadJSONFile(ctx, prefix+"/metadata.json", metaPath); err != nil {
		s.log.Warn("sync: upload root metadata", "slug", slug, "err", err)
	}
	for lang := range m.Languages {
		langPath := filepath.Join(s.diskPath, slug, lang, "metadata.json")
		if err := s.uploadJSONFile(ctx, prefix+"/"+lang+"/metadata.json", langPath); err != nil {
			s.log.Warn("sync: upload lang metadata", "slug", slug, "lang", lang, "err", err)
		}
	}

	return nil
}

func (s *SyncService) dictIDForSlug(slug string) string {
	entry, found, err := s.repo.GetDictionaryBySlug(slug)
	if err != nil || !found {
		return ""
	}
	return entry.ID
}

func countPageFiles(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, f := range files {
		if !f.IsDir() && isImageFile(f.Name()) {
			count++
		}
	}
	return count, nil
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif":
		return true
	}
	return false
}
