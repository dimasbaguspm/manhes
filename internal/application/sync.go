package application

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"manga-engine/internal/domain"
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

// HandleChapterUploaded handles a ChapterUploaded event.
// In the new pipeline, chapters are already uploaded by the time this is called.
// This method handles the manga cover update and metadata finalization.
func (s *SyncService) HandleChapterUploaded(ctx context.Context, e domain.ChapterUploaded) error {
	dictID := s.dictIDForSlug(e.Slug)
	s.log.Info("chapter uploaded event",
		slog.String("dict_id", dictID),
		slog.String("slug", e.Slug),
		slog.String("lang", e.Language),
		slog.String("chapter", e.ChapterNum),
	)

	m, err := s.disk.ReadMetadata(e.Slug)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	if m == nil {
		return nil
	}

	// Upload cover if not already
	coverURL, _ := s.uploadCover(ctx, e.Slug, dictID)
	if coverURL != "" {
		if err := s.repo.UpdateMangaCover(ctx, e.Slug, coverURL); err != nil {
			s.log.Warn("update manga cover", "slug", e.Slug, "err", err)
		}
	}

	return nil
}

// SyncAll runs an initial full scan and then ticks at the configured interval.
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
	if err := s.repo.UpsertManga(ctx, domain.Manga{
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

func (s *SyncService) uploadCover(ctx context.Context, slug, dictID string) (string, error) {
	prefix := dictID
	if prefix == "" {
		prefix = slug
	}
	for _, ext := range coverExtensions {
		path := filepath.Join(s.diskPath, slug, "cover"+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "image/jpeg"
		}
		s.log.Info("uploading cover", slog.String("dict_id", dictID), slog.String("slug", slug), slog.String("ext", ext))
		url, err := s.s3.Upload(ctx, prefix+"/cover"+ext, data, ct)
		if err != nil {
			return "", err
		}
		s.log.Info("cover uploaded", slog.String("dict_id", dictID), slog.String("slug", slug), slog.String("url", url))
		if removeErr := os.Remove(path); removeErr != nil {
			s.log.Warn("sync: remove local cover", "path", path, "err", removeErr)
		}
		return url, nil
	}
	return "", nil
}

func (s *SyncService) uploadJSONFile(ctx context.Context, key, filePath string) error {
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := s.s3.Upload(ctx, key, data, "application/json"); err != nil {
		return err
	}
	if err := os.Remove(filePath); err != nil {
		s.log.Warn("sync: remove local json", "path", filePath, "err", err)
	}
	return nil
}

func (s *SyncService) dictIDForSlug(slug string) string {
	entry, found, err := s.repo.GetDictionaryBySlug(context.Background(), slug)
	if err != nil || !found {
		return ""
	}
	return entry.ID
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif":
		return true
	}
	return false
}
