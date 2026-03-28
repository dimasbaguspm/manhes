package application

import (
	"context"
	"mime"
	"path/filepath"
	"time"

	"manga-engine/internal/domain"
	"manga-engine/pkg/concurrent"
)

func (s *DictionaryService) RunDaemon(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshAll(ctx)
		}
	}
}

func (s *DictionaryService) refreshAll(ctx context.Context) {
	// Fetch all entries in one large page (no UI pagination needed here).
	page, err := s.repo.ListDictionary(ctx, domain.DictionaryFilter{PageSize: 10_000})
	if err != nil {
		s.log.Error("dictionary daemon: list", "err", err)
		return
	}
	for _, e := range page.Items {
		if ctx.Err() != nil {
			return
		}
		if err := s.refresh(ctx, e.ID); err != nil {
			s.log.Warn("dictionary daemon: refresh", "id", e.ID, "slug", e.Slug, "err", err)
		}
	}
}

func (s *DictionaryService) refresh(ctx context.Context, id string) error {
	entry, found, err := s.repo.GetDictionary(ctx, id)
	if err != nil || !found {
		return err
	}

	ordered := s.registry.Ordered()

	// Build priority index for tie-breaking when two sources have equal chapter counts.
	priorityIndex := make(map[string]int, len(ordered))
	for i, src := range ordered {
		priorityIndex[src.Source()] = i
	}

	type sourceStatResult struct {
		name string
		stat domain.SourceStat
	}

	var g concurrent.Collect[sourceStatResult]
	for sourceName, sourceID := range entry.Sources {
		var src domain.Scraper
		for _, candidate := range ordered {
			if candidate.Source() == sourceName {
				src = candidate
				break
			}
		}
		g.Go(func() sourceStatResult {
			return sourceStatResult{name: sourceName, stat: s.fetchSourceStat(ctx, src, sourceID)}
		})
	}

	type best struct {
		source string
		count  int
		prio   int
	}
	newStats := make(map[string]domain.SourceStat, len(entry.Sources))
	// langBest tracks the highest-chapter-count source per language.
	langBest := make(map[string]best)
	for _, r := range g.Wait() {
		newStats[r.name] = r.stat
		prio := priorityIndex[r.name]
		for lang, count := range r.stat.ChaptersByLang {
			if b, seen := langBest[lang]; !seen || count > b.count || (count == b.count && prio < b.prio) {
				langBest[lang] = best{source: r.name, count: count, prio: prio}
			}
		}
	}

	newBest := make(map[string]string, len(langBest))
	for lang, b := range langBest {
		newBest[lang] = b.source
	}

	if entry.CoverURL == "" {
		entry.CoverURL = s.fetchCover(ctx, entry, ordered)
	}

	now := time.Now()
	entry.SourceStats = newStats
	entry.BestSource = newBest
	entry.RefreshedAt = &now
	if err := s.repo.UpsertDictionary(ctx, entry); err != nil {
		return err
	}

	// Populate manga table with scraped metadata for newly discovered entries
	// (before ingest has run). Skip if already ingested to avoid redundant HTTP calls.
	s.syncMangaMetadata(ctx, entry, ordered)
	return nil
}

// syncMangaMetadata populates the manga catalog table with metadata from the
// best available scraper source. It only runs when no manga entry exists yet
// (i.e., before ingest), so it does not override ingested data.
func (s *DictionaryService) syncMangaMetadata(ctx context.Context, entry domain.DictionaryEntry, ordered []domain.Scraper) {
	if _, found, err := s.repo.GetMangaBySlug(ctx, entry.Slug); err != nil || found {
		return
	}
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			continue
		}
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil || manga == nil {
			continue
		}
		manga.Slug = entry.Slug
		manga.CoverURL = entry.CoverURL // use S3 URL from dictionary
		if err := s.repo.UpsertManga(ctx, *manga); err != nil {
			s.log.Warn("dictionary refresh: sync manga metadata", "slug", entry.Slug, "err", err)
		}
		return
	}
}

func (s *DictionaryService) fetchSourceStat(ctx context.Context, src domain.Scraper, sourceID string) domain.SourceStat {
	stat := domain.SourceStat{
		ChaptersByLang: map[string]int{},
		FetchedAt:      time.Now(),
	}
	if src == nil {
		stat.Err = "scraper not registered"
		return stat
	}
	chapters, err := src.FetchChapterList(ctx, sourceID)
	if err != nil {
		stat.Err = err.Error()
		return stat
	}
	for _, ch := range chapters {
		stat.ChaptersByLang[ch.Language]++
	}
	return stat
}

func (s *DictionaryService) fetchCover(ctx context.Context, entry domain.DictionaryEntry, ordered []domain.Scraper) string {
	for _, src := range ordered {
		sourceID, ok := entry.Sources[src.Source()]
		if !ok {
			continue
		}
		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil || manga == nil || manga.CoverURL == "" {
			continue
		}
		data, err := s.dl.Download(ctx, manga.CoverURL)
		if err != nil {
			continue
		}
		ext := filepath.Ext(manga.CoverURL)
		if ext == "" {
			ext = ".jpg"
		}
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			ct = "image/jpeg"
		}
		url, err := s.s3c.Upload(ctx, entry.ID+"/cover"+ext, data, ct)
		if err != nil {
			s.log.Warn("dictionary refresh: upload cover", "id", entry.ID, "err", err)
			continue
		}
		return url
	}
	return ""
}
