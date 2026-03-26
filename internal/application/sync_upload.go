package application

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"time"

	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/storage"
)

func (s *SyncService) uploadChapter(ctx context.Context, ref domain.ChapterRef) error {
	chDir := filepath.Join(s.diskPath, ref.Slug, ref.Language, storage.ChapterDir(ref.ChapterNum))

	files, err := os.ReadDir(chDir)
	if err != nil {
		return fmt.Errorf("read chapter dir: %w", err)
	}

	prefix := ref.DictionaryID
	if prefix == "" {
		prefix = ref.Slug
	}

	upLog := s.log.With(
		slog.String("dict_id", ref.DictionaryID),
		slog.String("slug", ref.Slug),
		slog.String("lang", ref.Language),
		slog.Float64("chapter", ref.ChapterNum),
	)
	upLog.Info("uploading chapter")
	start := time.Now()

	pageIdx := 0
	for _, f := range files {
		if f.IsDir() || !isImageFile(f.Name()) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(chDir, f.Name()))
		if err != nil {
			return fmt.Errorf("read page %s: %w", f.Name(), err)
		}

		ext := filepath.Ext(f.Name())
		key := fmt.Sprintf("%s/%s/%s/%s",
			prefix,
			ref.Language,
			storage.ChapterDir(ref.ChapterNum),
			storage.PageFile(pageIdx, ext),
		)

		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "image/webp"
		}

		url, err := s.s3.Upload(ctx, key, data, contentType)
		if err != nil {
			return fmt.Errorf("upload page: %w", err)
		}

		if err := s.repo.UpsertPage(ref.Slug, ref.Language, ref.ChapterNum, pageIdx, url); err != nil {
			return fmt.Errorf("upsert page: %w", err)
		}

		pageIdx++
	}

	if err := s.repo.MarkChapterUploaded(ref.Slug, ref.Language, ref.ChapterNum); err != nil {
		return err
	}
	upLog.Info("chapter uploaded", slog.Int("pages", pageIdx), slog.Int64("duration_ms", time.Since(start).Milliseconds()))
	// Transition to available only when all chapters for this slug are uploaded
	// AND ingest is not still in progress (fetching state means some chapter
	// downloads failed and will be retried by the watchlist daemon).
	hasPending, err := s.repo.HasPendingChapters(ref.Slug)
	if err != nil {
		s.log.Warn("sync: check pending chapters", "slug", ref.Slug, "err", err)
	} else if !hasPending {
		entry, found, err := s.repo.GetDictionaryBySlug(ref.Slug)
		if err != nil || !found {
			s.log.Warn("sync: get dictionary for state check", "slug", ref.Slug, "err", err)
		} else if entry.State == domain.StateFetching {
			s.log.Info("sync: skipping available transition, ingest still in progress", "slug", ref.Slug)
		} else {
			if err := s.repo.SetDictionaryStateBySlug(ref.Slug, domain.StateAvailable); err != nil {
				s.log.Warn("sync: set dictionary state available", "slug", ref.Slug, "err", err)
			}
		}
	}
	for _, f := range files {
		if f.IsDir() || !isImageFile(f.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(chDir, f.Name())); err != nil {
			s.log.Warn("sync: remove local page", "file", f.Name(), "err", err)
		}
	}

	manifestPath := filepath.Join(chDir, "chapter.json")
	manifestKey := fmt.Sprintf("%s/%s/%s/chapter.json", prefix, ref.Language, storage.ChapterDir(ref.ChapterNum))
	if err := s.uploadJSONFile(ctx, manifestKey, manifestPath); err != nil {
		s.log.Warn("sync: upload chapter manifest", "slug", ref.Slug, "err", err)
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
